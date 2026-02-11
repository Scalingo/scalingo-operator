/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Scalingo/go-utils/errors/v3"
	databasesv1alpha1 "github.com/Scalingo/scalingo-operator/api/v1alpha1"
	"github.com/Scalingo/scalingo-operator/internal/controller/adapters"
	"github.com/Scalingo/scalingo-operator/internal/controller/helpers"
	"github.com/Scalingo/scalingo-operator/internal/domain"
	databasebase "github.com/Scalingo/scalingo-operator/internal/usecases/database/base"
)

// PostgreSQLReconciler reconciles a PostgreSQL object
type PostgreSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/reconcile
func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the instance.
	var postgresql databasesv1alpha1.PostgreSQL
	err := r.Get(ctx, req.NamespacedName, &postgresql)
	if err != nil {
		// Handle error, if it's not found, there's nothing to do.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO(david): TO REMOVE
	log.Info("### start", "is running annotation", postgresql.Annotations[helpers.DatabaseAnnotationIsRunning])
	log.Info("### start", "status", postgresql.Status)

	// Initialize:
	// 1/ add finalizer + requeue
	// 2/ set initial Status.Condition + requeue
	// 3/ set initial Annotations + requeue
	//
	// Note: the `requeue` prevents resource update conflicts.
	containsFinalizer := controllerutil.ContainsFinalizer(&postgresql, helpers.PostgreSQLFinalizerName)
	hasRunningAnnotation := metav1.HasAnnotation(postgresql.ObjectMeta, helpers.DatabaseAnnotationIsRunning)
	isDatabaseStatusIntialized := helpers.IsDatabaseInitialized(postgresql.Status.Conditions)

	switch {
	case !containsFinalizer:
		log.Info("Add finalizer to resource", "finalizer", helpers.PostgreSQLFinalizerName)

		controllerutil.AddFinalizer(&postgresql, helpers.PostgreSQLFinalizerName)
		err := r.Update(ctx, &postgresql)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "add PostgreSQL finalizer")
		}
		return ctrl.Result{Requeue: true}, nil
	case !hasRunningAnnotation && !isDatabaseStatusIntialized:
		log.Info("Initialize status conditions")

		orig := postgresql.DeepCopy()
		helpers.SetDatabaseInitialStatus(&postgresql.Status.Conditions)

		err := r.Status().Patch(ctx, &postgresql, client.MergeFrom(orig))
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "patch database resource initial status")
		}
		return ctrl.Result{Requeue: true}, nil
	case !hasRunningAnnotation && isDatabaseStatusIntialized:
		log.Info("Initialize annotations")

		orig := postgresql.DeepCopy()
		helpers.SetDatabaseIsNotRunning(&postgresql.ObjectMeta)

		err := r.Patch(ctx, &postgresql, client.MergeFrom(orig))
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "patch database resource initial annotations")
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if true {
		return ctrl.Result{}, nil
	}

	// Read secret token.
	secretManager := helpers.NewSecretManager(r.Client, &postgresql)

	authSecret := domain.Secret{Namespace: req.Namespace, Name: postgresql.Spec.AuthSecret.Name, Key: postgresql.Spec.AuthSecret.Key}
	log.Info("Get auth secret", "secret", authSecret)

	apiToken, err := secretManager.GetSecret(ctx, authSecret)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "get auth secret")
	}

	// Create database manager.
	dbManager, err := databasebase.NewManager(ctx, domain.DatabaseTypePostgreSQL, apiToken, postgresql.Spec.Region)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "create database manager")
	}

	expectedDB, err := adapters.PostgreSQLToDatabase(ctx, postgresql)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "bad custom resource format")
	}
	isDatabaseDeletionRequested := helpers.IsDatabaseDeletionRequested(postgresql.ObjectMeta)
	isDatabaseAvailable := helpers.IsDatabaseAvailable(postgresql.Status.Conditions)
	isDatabaseProvisioning := helpers.IsDatabaseProvisioning(postgresql.Status.Conditions)
	isDatabaseRunning := helpers.IsDatabaseRunning(postgresql.ObjectMeta)

	log.Info("Current state",
		"database", postgresql.Status.ScalingoDatabaseID,
		"deletion_requested", isDatabaseDeletionRequested,
		"available", isDatabaseAvailable,
		"provisioning", isDatabaseProvisioning,
		"running", isDatabaseRunning)

	requeueLater := false

	switch {
	case isDatabaseDeletionRequested:
		// Delete database.
		log.Info("Delete database resource")

		err := dbManager.DeleteDatabase(ctx, postgresql.Status.ScalingoDatabaseID)
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(ctx, err, "delete database id %s", postgresql.Status.ScalingoDatabaseID)
		}
		controllerutil.RemoveFinalizer(&postgresql, helpers.PostgreSQLFinalizerName)
		err = r.Update(ctx, &postgresql)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "remove finalizer")
		}
	case !isDatabaseAvailable && postgresql.Status.ScalingoDatabaseID == "":
		// Create database.
		log.Info("Create database resource")

		newDB, err := dbManager.CreateDatabase(ctx, expectedDB)
		if err != nil {
			log.Error(err, "Create database", "database", expectedDB)
			return ctrl.Result{}, errors.Wrapf(ctx, err, "create database %s", expectedDB.Name)
		}

		orig := postgresql.DeepCopy()
		postgresql.Status.ScalingoDatabaseID = newDB.ID
		helpers.SetDatabaseStatusProvisioning(&postgresql.Status.Conditions)

		err = r.Status().Patch(ctx, &postgresql, client.MergeFrom(orig))
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "patch database resource status")
		}
		requeueLater = true
	case isDatabaseAvailable && !isDatabaseRunning:
		// Update running annotation.
		log.Info("Update database running annotation")

		orig := postgresql.DeepCopy()
		helpers.SetDatabaseIsRunning(&postgresql.ObjectMeta)

		err = r.Patch(ctx, &postgresql, client.MergeFrom(orig))
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "patch database resource running annotation")
		}
	case isDatabaseAvailable && !isDatabaseProvisioning && postgresql.Status.ScalingoDatabaseID != "":
		// Update database.
		log.Info("Update database resource")

		err = dbManager.UpdateDatabase(ctx, postgresql.Status.ScalingoDatabaseID, expectedDB)
		if err != nil {
			log.Error(err, "Update database", "database", expectedDB)
			return ctrl.Result{}, errors.Wrapf(ctx, err, "update database %s", expectedDB.Name)
		}

		orig := postgresql.DeepCopy()
		helpers.SetDatabaseStatusProvisioned(&postgresql.Status.Conditions)

		err = r.Patch(ctx, &postgresql, client.MergeFrom(orig))
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "patch database resource status")
		}
	case isDatabaseProvisioning && postgresql.Status.ScalingoDatabaseID != "":
		// Wait for database creation.
		currentDB, err := dbManager.GetDatabase(ctx, postgresql.Status.ScalingoDatabaseID)
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(ctx, err, "get current database %s", postgresql.Status.ScalingoDatabaseID)
		}

		if currentDB.Status == domain.DatabaseStatusRunning {
			log.Info("Database is provisioned")

			orig := postgresql.DeepCopy()
			helpers.SetDatabaseStatusProvisioned(&postgresql.Status.Conditions)

			err = r.Status().Patch(ctx, &postgresql, client.MergeFrom(orig))
			if err != nil {
				return ctrl.Result{}, errors.Wrap(ctx, err, "update database resource status")
			}

			// Write connection info in secret
			dbURL, err := dbManager.GetDatabaseURL(ctx, currentDB)
			if err != nil {
				return ctrl.Result{}, errors.Wrap(ctx, err, "get database url")
			}

			connInfoSecret := domain.Secret{
				Namespace: req.Namespace,
				Name:      postgresql.Spec.ConnInfoSecretTarget.Name,
				Key:       domain.ComposeConnectionURLName(postgresql.Spec.ConnInfoSecretTarget.Prefix, dbURL.Name),
				Value:     dbURL.Value,
			}
			log.Info("Write connection info secret", "secret", connInfoSecret)

			err = secretManager.SetSecret(ctx, connInfoSecret)
			if err != nil {
				return ctrl.Result{}, errors.Wrapf(ctx, err, "set secret %s", connInfoSecret.Key)
			}
			return ctrl.Result{Requeue: true}, nil // Requeue for the is running annotation.
		} else {
			log.Info("Waiting for database being provisioned")
			requeueLater = true
		}
	}

	if requeueLater {
		log.Info("Requeue after delay", "delay", helpers.RequeueDelay)
		return ctrl.Result{RequeueAfter: helpers.RequeueDelay}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasesv1alpha1.PostgreSQL{}).
		Named("postgresql").
		Complete(r)
}
