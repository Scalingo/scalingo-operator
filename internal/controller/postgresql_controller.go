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

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Scalingo/go-utils/errors/v3"
	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
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

const (
	netPeeringRequestSuffix = "-net-peering-request"
)

var netPeeringRequestGVK = schema.GroupVersionKind{
	Group:   "oks.dev",
	Version: "v1beta",
	Kind:    "NetPeeringRequest",
}

// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oks.dev,resources=netpeeringrequests,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/reconcile
func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the instance.
	var postgresql apiv1.PostgreSQL
	err := r.Get(ctx, req.NamespacedName, &postgresql)
	if err != nil {
		// Handle error, if it's not found, there's nothing to do.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// State information.
	containsFinalizer := controllerutil.ContainsFinalizer(&postgresql, helpers.PostgreSQLFinalizerName)
	hasRunningAnnotation := metav1.HasAnnotation(postgresql.ObjectMeta, helpers.DatabaseAnnotationIsRunning)
	isDatabaseStatusInitialized := helpers.IsDatabaseInitialized(postgresql.Status.Conditions)

	isDatabaseProvisioning := helpers.IsDatabaseProvisioning(postgresql.Status.Conditions)
	isDatabaseAvailable := helpers.IsDatabaseAvailable(postgresql.Status.Conditions)
	isDatabaseRunning := helpers.IsDatabaseRunning(postgresql.ObjectMeta)
	isDatabaseDeletionRequested := helpers.IsDatabaseDeletionRequested(postgresql.ObjectMeta)

	// Trigger variables.
	var (
		triggerUpdate       bool
		triggerStatusUpdate bool
		triggerRequeueLater time.Duration
	)

	// Initialization and resource updates.
	// (no Scalingo client interaction)
	//
	// Initialize steps:
	// 1/ add finalizer + requeue
	// 2/ set initial Status.Condition + requeue
	// 3/ set initial Annotations + requeue
	//
	// Note: the `requeue` prevents resource update conflicts.
	switch {
	case !containsFinalizer:
		log.Info("Add finalizer to resource", "finalizer", helpers.PostgreSQLFinalizerName)

		controllerutil.AddFinalizer(&postgresql, helpers.PostgreSQLFinalizerName)
		err := r.Update(ctx, &postgresql)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "add resource finalizer")
		}
		return ctrl.Result{RequeueAfter: helpers.RequeueShortDelay}, nil

	case !isDatabaseStatusInitialized:
		log.Info("Initialize resource status conditions")

		helpers.SetDatabaseInitialStatus(&postgresql.Status.Conditions)
		triggerStatusUpdate = true

	case isDatabaseStatusInitialized && !hasRunningAnnotation:
		log.Info("Initialize resource annotations")

		helpers.SetDatabaseIsNotRunning(&postgresql.ObjectMeta)
		triggerUpdate = true

	case isDatabaseAvailable && !isDatabaseRunning:
		// Update running annotation.
		log.Info("Update database resource running annotation")

		helpers.SetDatabaseIsRunning(&postgresql.ObjectMeta)
		triggerUpdate = true
	}

	// Apply triggered resource updates with short delay requeue.
	switch {
	case triggerStatusUpdate:
		err := r.Status().Update(ctx, &postgresql)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "update database resource status")
		}
		return ctrl.Result{RequeueAfter: helpers.RequeueShortDelay}, nil

	case triggerUpdate:
		err := r.Update(ctx, &postgresql)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "update database resource")
		}
		return ctrl.Result{RequeueAfter: helpers.RequeueShortDelay}, nil
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

	// Requested/expected database resource.
	expectedDB, err := adapters.PostgreSQLToDatabase(ctx, postgresql)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "bad custom resource format")
	}

	log.Info("Current state",
		"database", postgresql.Status.ScalingoDatabaseID,
		"deletion_requested", isDatabaseDeletionRequested,
		"available", isDatabaseAvailable,
		"provisioning", isDatabaseProvisioning,
		"running", isDatabaseRunning)

	isOutscaleOKSNetPeeringEnabled := postgresql.Spec.Networking.IsOutscaleOKSNetPeeringEnabled()

	// Create/update/delete database.
	switch {
	case isDatabaseDeletionRequested:
		log.Info("Delete database")

		if isOutscaleOKSNetPeeringEnabled {
			log.Info("Delete database net peering and NetPeeringRequest")

			if postgresql.Status.ScalingoDatabaseID != "" {
				err := dbManager.DeleteDatabaseNetPeerings(ctx, postgresql.Status.ScalingoDatabaseID)
				if err != nil {
					return ctrl.Result{}, errors.Wrapf(ctx, err, "delete database net peerings id %s", postgresql.Status.ScalingoDatabaseID)
				}
			}

			err = r.deleteNetPeeringRequest(ctx, postgresql)
			if err != nil {
				return ctrl.Result{}, errors.Wrap(ctx, err, "delete net peering request")
			}
		}

		if postgresql.Status.ScalingoDatabaseID == "" {
			log.Info("Database provisioning requested but no database created yet, skip database deletion")
		} else {
			err := dbManager.DeleteDatabase(ctx, postgresql.Status.ScalingoDatabaseID)
			if err != nil {
				return ctrl.Result{}, errors.Wrapf(ctx, err, "delete database id %s", postgresql.Status.ScalingoDatabaseID)
			}
		}

		controllerutil.RemoveFinalizer(&postgresql, helpers.PostgreSQLFinalizerName)
		err = r.Update(ctx, &postgresql)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "remove resource finalizer")
		}

	case !isDatabaseAvailable && postgresql.Status.ScalingoDatabaseID == "":
		log.Info("Create database")

		newDB, err := dbManager.CreateDatabase(ctx, expectedDB)
		if err != nil {
			log.Error(err, "Create database", "database", expectedDB)
			return ctrl.Result{}, errors.Wrapf(ctx, err, "create database %s", expectedDB.Name)
		}

		postgresql.Status.ScalingoDatabaseID = newDB.ID
		helpers.SetDatabaseStatusProvisioning(&postgresql.Status.Conditions)
		triggerStatusUpdate = true
		triggerRequeueLater = helpers.RequeueLongDelay

	case isDatabaseAvailable && !isDatabaseProvisioning && postgresql.Status.ScalingoDatabaseID != "":
		log.Info("Update database")

		dbStatus, err := dbManager.UpdateDatabase(ctx, postgresql.Status.ScalingoDatabaseID, expectedDB)
		if err != nil {
			log.Error(err, "Update database", "database", expectedDB)
			return ctrl.Result{}, errors.Wrapf(ctx, err, "update database %s", expectedDB.Name)
		}

		if dbStatus == domain.DatabaseStatusProvisioning {
			log.Info("Waiting for database being provisioned")
			helpers.SetDatabaseStatusProvisioning(&postgresql.Status.Conditions)
			triggerStatusUpdate = true
		}

	case isDatabaseProvisioning && postgresql.Status.ScalingoDatabaseID != "":
		// Keep applying compatible updates (e.g firewall rules) while the database is
		// available (created) and provisioning.
		if isDatabaseAvailable {
			_, err := dbManager.UpdateDatabase(ctx, postgresql.Status.ScalingoDatabaseID, expectedDB)
			if err != nil {
				log.Error(err, "Update database while provisioning", "database", expectedDB)
				return ctrl.Result{}, errors.Wrapf(ctx, err, "update database %s while provisioning", expectedDB.Name)
			}
		}

		// Wait for database creation/plan update completion.
		currentDB, err := dbManager.GetDatabase(ctx, postgresql.Status.ScalingoDatabaseID)
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(ctx, err, "get current database %s", postgresql.Status.ScalingoDatabaseID)
		}

		if currentDB.Status == domain.DatabaseStatusRunning {
			log.Info("Database is provisioned")

			helpers.SetDatabaseStatusProvisioned(&postgresql.Status.Conditions)
			triggerStatusUpdate = true

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
			triggerRequeueLater = helpers.RequeueShortDelay // Requeue for the is running annotation.

		} else {
			log.Info("Waiting for database being provisioned")
			triggerRequeueLater = helpers.RequeueLongDelay
		}
	}

	if !isDatabaseDeletionRequested {
		if !isOutscaleOKSNetPeeringEnabled {
			log.Info("Outscale OKS net peering disabled, cleanup resources")

			if postgresql.Status.ScalingoDatabaseID != "" {
				err := dbManager.DeleteDatabaseNetPeerings(ctx, postgresql.Status.ScalingoDatabaseID)
				if err != nil {
					return ctrl.Result{}, errors.Wrapf(ctx, err, "delete database net peerings id %s", postgresql.Status.ScalingoDatabaseID)
				}
			}

			err = r.deleteNetPeeringRequest(ctx, postgresql)
			if err != nil {
				return ctrl.Result{}, errors.Wrap(ctx, err, "delete net peering request")
			}
		} else if isDatabaseAvailable && !isDatabaseProvisioning && postgresql.Status.ScalingoDatabaseID != "" {
			log.Info("Reconcile Outscale OKS net peering")

			databaseNetworkConfig, err := dbManager.GetDatabaseNetworkConfiguration(ctx, postgresql.Status.ScalingoDatabaseID)
			if err != nil {
				return ctrl.Result{}, errors.Wrapf(ctx, err, "get database network configuration id %s", postgresql.Status.ScalingoDatabaseID)
			}

			netPeeringID, err := r.reconcileNetPeeringRequest(ctx, postgresql, databaseNetworkConfig.OutscaleNetID, databaseNetworkConfig.OutscaleAccountID)
			if err != nil {
				return ctrl.Result{}, errors.Wrap(ctx, err, "reconcile net peering request")
			}

			if netPeeringID == "" {
				log.Info("Waiting for NetPeeringRequest to be provisioned")
				triggerRequeueLater = helpers.RequeueLongDelay
			} else {
				err = dbManager.EnsureDatabaseNetPeering(ctx, postgresql.Status.ScalingoDatabaseID, netPeeringID)
				if err != nil {
					return ctrl.Result{}, errors.Wrapf(ctx, err, "ensure database net peering id %s", postgresql.Status.ScalingoDatabaseID)
				}
			}
		}
	}

	// Apply triggered resource updates and requeue later.
	if triggerStatusUpdate {
		log.Info("Update resource status", "statusConditions", postgresql.Status.Conditions)
		err := r.Status().Update(ctx, &postgresql)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "update database resource status")
		}
	}
	if triggerRequeueLater > 0 {
		log.Info("Requeue after", "delay", triggerRequeueLater)
		return ctrl.Result{RequeueAfter: triggerRequeueLater}, nil
	}

	log.Info("Ready")
	return ctrl.Result{}, nil
}

func (r *PostgreSQLReconciler) reconcileNetPeeringRequest(ctx context.Context, postgresql apiv1.PostgreSQL, accepterNetID, accepterOwnerID string) (string, error) {
	netPeeringRequest := &unstructured.Unstructured{}
	netPeeringRequest.SetGroupVersionKind(netPeeringRequestGVK)
	netPeeringRequest.SetNamespace(postgresql.Namespace)
	netPeeringRequest.SetName(postgresql.Name + netPeeringRequestSuffix)

	err := r.Get(ctx, client.ObjectKeyFromObject(netPeeringRequest), netPeeringRequest)
	switch {
	case k8serrors.IsNotFound(err):
		netPeeringRequest.Object = map[string]any{
			"apiVersion": "oks.dev/v1beta",
			"kind":       "NetPeeringRequest",
			"metadata": map[string]any{
				"name":      postgresql.Name + netPeeringRequestSuffix,
				"namespace": postgresql.Namespace,
			},
			"spec": map[string]any{
				"accepterNetId":   accepterNetID,
				"accepterOwnerId": accepterOwnerID,
			},
		}

		if err := controllerutil.SetControllerReference(&postgresql, netPeeringRequest, r.Scheme); err != nil {
			return "", errors.Wrap(ctx, err, "set controller reference")
		}

		if err := r.Create(ctx, netPeeringRequest); err != nil {
			return "", errors.Wrap(ctx, err, "create net peering request")
		}
		return "", nil
	case err != nil:
		return "", errors.Wrap(ctx, err, "get net peering request")
	}

	desiredSpec := map[string]any{
		"accepterNetId":   accepterNetID,
		"accepterOwnerId": accepterOwnerID,
	}
	currentSpec, _, _ := unstructured.NestedMap(netPeeringRequest.Object, "spec")

	if !equalNetPeeringSpec(currentSpec, desiredSpec) {
		if err := unstructured.SetNestedMap(netPeeringRequest.Object, desiredSpec, "spec"); err != nil {
			return "", errors.Wrap(ctx, err, "set net peering request spec")
		}

		if err := r.Update(ctx, netPeeringRequest); err != nil {
			return "", errors.Wrap(ctx, err, "update net peering request")
		}
		return "", nil
	}

	netPeeringID := extractNetPeeringID(netPeeringRequest)
	return netPeeringID, nil
}

func (r *PostgreSQLReconciler) deleteNetPeeringRequest(ctx context.Context, postgresql apiv1.PostgreSQL) error {
	netPeeringRequest := &unstructured.Unstructured{}
	netPeeringRequest.SetGroupVersionKind(netPeeringRequestGVK)
	netPeeringRequest.SetNamespace(postgresql.Namespace)
	netPeeringRequest.SetName(postgresql.Name + netPeeringRequestSuffix)

	err := r.Delete(ctx, netPeeringRequest)
	if k8serrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return errors.Wrap(ctx, err, "delete net peering request")
	}
	return nil
}

func extractNetPeeringID(netPeeringRequest *unstructured.Unstructured) string {
	for _, path := range [][]string{
		{"status", "netPeeringId"},
	} {
		value, found, _ := unstructured.NestedString(netPeeringRequest.Object, path...)
		if found && value != "" {
			return value
		}
	}
	return ""
}

func equalNetPeeringSpec(currentSpec map[string]any, desiredSpec map[string]any) bool {
	if currentSpec == nil {
		return false
	}
	currentNetID, _, _ := unstructured.NestedString(map[string]any{"spec": currentSpec}, "spec", "accepterNetId")
	currentOwnerID, _, _ := unstructured.NestedString(map[string]any{"spec": currentSpec}, "spec", "accepterOwnerId")

	return currentNetID == desiredSpec["accepterNetId"] && currentOwnerID == desiredSpec["accepterOwnerId"]
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.PostgreSQL{}).
		Named("postgresql").
		Complete(r)
}
