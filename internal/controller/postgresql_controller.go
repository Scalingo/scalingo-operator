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
	"time"

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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PostgreSQL object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
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

	// Add finalizer.
	if !controllerutil.ContainsFinalizer(&postgresql, helpers.PostgreSQLFinalizerName) {
		log.Info("Add finalizer to resource", "finalizer", helpers.PostgreSQLFinalizerName)
		controllerutil.AddFinalizer(&postgresql, helpers.PostgreSQLFinalizerName)
		if err := r.Update(ctx, &postgresql); err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "add PostgreSQL finalizer")
		}
	}

	// Initialize status conditions and annotations if not present.
	if !metav1.HasAnnotation(postgresql.ObjectMeta, helpers.DatabaseAnnotationIsRunning) {
		log.Info("Initialize status conditions and annotations")

		helpers.SetDatabaseInitialState(&postgresql.ObjectMeta, &postgresql.Status.Conditions)
		if err := r.Update(ctx, &postgresql); err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "set intial state")
		}
	}

	// Read secret token.
	log.Info("Get auth secret", "namespace", req.Namespace, "name", postgresql.Spec.AuthSecret.Name, "key", postgresql.Spec.AuthSecret.Key)
	apiToken, err := helpers.GetSecret(ctx, r.Client, req.Namespace, postgresql.Spec.AuthSecret.Name, postgresql.Spec.AuthSecret.Key)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "get auth secret")
	}

	// Create database manager.
	dbManager, err := databasebase.NewManager(ctx, domain.DatabaseTypePostgreSQL, apiToken, postgresql.Spec.Region)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "create database manager")
	}

	// Create, modify, delete database.
	expectedDB := adapters.PostgreSQLToDatabase(postgresql)
	isDatabaseRunning := helpers.IsDatabaseRunning(postgresql.ObjectMeta)
	isDatabaseAvailable := helpers.IsDatabaseAvailable(postgresql.Status.Conditions)
	IsDatabaseProvisionned := helpers.IsDatabaseAvailable(postgresql.Status.Conditions)
	requeue := false

	log.Info("Current state",
		"database", postgresql.Status.ScalingoDatabaseID,
		"running", isDatabaseRunning,
		"available", isDatabaseAvailable,
		"provisionned", IsDatabaseProvisionned)

	if !isDatabaseRunning && postgresql.Status.ScalingoDatabaseID == "" {
		// Create database.
		log.Info("Create new database")

		newDB, err := dbManager.CreateDatabase(ctx, expectedDB)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "create new database")
		}

		postgresql.Status.ScalingoDatabaseID = newDB.ID
		helpers.SetDatabaseStatusProvisionning(&postgresql.Status.Conditions)
		if err = r.Status().Update(ctx, &postgresql); err != nil {
			return ctrl.Result{}, errors.Wrapf(ctx, err, "update database %s status", newDB.ID)
		}
		requeue = true
	} else if postgresql.Status.ScalingoDatabaseID != "" {
		// Wait for database creation.
		currentDB, err := dbManager.GetDatabase(ctx, postgresql.Status.ScalingoDatabaseID)
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(ctx, err, "get current database %s", postgresql.Status.ScalingoDatabaseID)
		}

		if currentDB.Status == domain.AddonStatusRunning {
			log.Info("Database is provisionned")
			helpers.SetDatabaseStatusProvisionned(&postgresql.ObjectMeta, &postgresql.Status.Conditions)
			if err = r.Update(ctx, &postgresql); err != nil {
				return ctrl.Result{}, errors.Wrapf(ctx, err, "update database %s status", currentDB.ID)
			}

			// TODO(david): set conn info secret.
		} else {
			log.Info("Waiting for database being provisionned")
			requeue = true
		}
	}

	if requeue {
		delay := 30 * time.Second
		log.Info("Requeue after delay", "delay", delay)
		return ctrl.Result{RequeueAfter: delay}, nil
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
