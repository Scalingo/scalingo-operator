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
	"fmt"
	"time"

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
	databaseusecases "github.com/Scalingo/scalingo-operator/internal/usecases/database"
	databasebase "github.com/Scalingo/scalingo-operator/internal/usecases/database/base"
)

// PostgreSQLReconciler reconciles a PostgreSQL object
type PostgreSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	netPeeringRequestSuffix = "net-peering-request"
)

var netPeeringRequestGVK = schema.GroupVersionKind{
	Group:   "oks.dev",
	Version: "v1beta",
	Kind:    "NetPeeringRequest",
}

var netPeeringRequestListGVK = schema.GroupVersionKind{
	Group:   "oks.dev",
	Version: "v1beta",
	Kind:    "NetPeeringRequestList",
}

var netPeeringGVK = schema.GroupVersionKind{
	Group:   "oks.dev",
	Version: "v1beta",
	Kind:    "NetPeering",
}

var netPeeringListGVK = schema.GroupVersionKind{
	Group:   "oks.dev",
	Version: "v1beta",
	Kind:    "NetPeeringList",
}

// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oks.dev,resources=netpeeringrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oks.dev,resources=netpeerings,verbs=get;list;delete

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
			err = r.deleteNetPeerings(ctx, dbManager, postgresql)
			if err != nil {
				return ctrl.Result{}, errors.Wrap(ctx, err, "delete net peering resources")
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

			endpoints, err := dbManager.GetDatabaseEndpoints(ctx, currentDB.ID)
			if err != nil {
				return ctrl.Result{}, errors.Wrap(ctx, err, "get database endpoints")
			}

			for _, endpoint := range endpoints {
				endpointURL, err := domain.ComposeEndpointConnectionURL(ctx, dbURL.Value, endpoint)
				if err != nil {
					return ctrl.Result{}, errors.Wrap(ctx, err, "compose endpoint connection url")
				}

				endpointConnInfoSecret := domain.Secret{
					Namespace: req.Namespace,
					Name:      postgresql.Spec.ConnInfoSecretTarget.Name,
					Key:       domain.ComposeEndpointConnectionURLName(postgresql.Spec.ConnInfoSecretTarget.Prefix, dbURL.Name, endpoint.Type),
					Value:     endpointURL,
				}
				log.Info("Write endpoint connection info secret", "secret", endpointConnInfoSecret)

				err = secretManager.SetSecret(ctx, endpointConnInfoSecret)
				if err != nil {
					return ctrl.Result{}, errors.Wrapf(ctx, err, "set secret %s", endpointConnInfoSecret.Key)
				}
			}
			triggerRequeueLater = helpers.RequeueShortDelay // Requeue for the is running annotation.

		} else {
			log.Info("Waiting for database being provisioned")
			triggerRequeueLater = helpers.RequeueLongDelay
		}
	}

	netPeeringRequeue, err := r.reconcileOutscaleOKSNetPeering(
		ctx,
		postgresql,
		dbManager,
		isDatabaseDeletionRequested,
		isOutscaleOKSNetPeeringEnabled,
		isDatabaseAvailable,
		isDatabaseProvisioning,
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	if netPeeringRequeue > 0 {
		triggerRequeueLater = netPeeringRequeue
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

func (r *PostgreSQLReconciler) reconcileOutscaleOKSNetPeering(
	ctx context.Context,
	postgresql apiv1.PostgreSQL,
	dbManager databaseusecases.Manager,
	isDatabaseDeletionRequested bool,
	isOutscaleOKSNetPeeringEnabled bool,
	isDatabaseAvailable bool,
	isDatabaseProvisioning bool,
) (time.Duration, error) {
	if isDatabaseDeletionRequested {
		return 0, nil
	}

	log := logf.FromContext(ctx)
	if !isOutscaleOKSNetPeeringEnabled {
		err := r.deleteNetPeerings(ctx, dbManager, postgresql)
		if err != nil {
			return 0, errors.Wrap(ctx, err, "delete net peering resources")
		}
		return 0, nil
	}

	if !isDatabaseAvailable || isDatabaseProvisioning || postgresql.Status.ScalingoDatabaseID == "" {
		return 0, nil
	}

	log.Info("Reconcile Outscale OKS net peering")
	netPeeringID, err := r.ensureOKSNetPeeringRequest(ctx, dbManager, postgresql)
	if err != nil {
		return 0, errors.Wrap(ctx, err, "reconcile net peering request")
	}

	if netPeeringID == "" {
		log.Info("Waiting for NetPeeringRequest to be provisioned")
		return helpers.RequeueLongDelay, nil
	}

	err = dbManager.EnsureDatabaseNetPeering(ctx, postgresql.Status.ScalingoDatabaseID, netPeeringID)
	if err != nil {
		return 0, errors.Wrapf(ctx, err, "ensure database net peering id %s", postgresql.Status.ScalingoDatabaseID)
	}
	return 0, nil
}

func (r *PostgreSQLReconciler) ensureOKSNetPeeringRequest(ctx context.Context, dbManager databaseusecases.Manager, postgresql apiv1.PostgreSQL) (string, error) {
	existingNetPeeringRequestsForDatabase, err := r.listExistingNetPeeringRequestsForDatabase(ctx, postgresql)
	if err != nil {
		return "", errors.Wrap(ctx, err, "list existing net peering requests for database")
	}

	netPeeringRequest := &unstructured.Unstructured{}
	if len(existingNetPeeringRequestsForDatabase) == 0 {
		netPeeringRequest, err = r.createOKSNetPeeringRequest(ctx, dbManager, postgresql)
		if err != nil {
			return "", errors.Wrap(ctx, err, "create net peering request")
		}
	} else if len(existingNetPeeringRequestsForDatabase) == 1 {
		netPeeringRequest = &existingNetPeeringRequestsForDatabase[0]
	} else {
		return "", errors.New(ctx, "multiple net peering requests found for database")
	}

	netPeeringID := extractNetPeeringID(*netPeeringRequest)
	return netPeeringID, nil
}

func (r *PostgreSQLReconciler) createOKSNetPeeringRequest(ctx context.Context, dbManager databaseusecases.Manager, postgresql apiv1.PostgreSQL) (*unstructured.Unstructured, error) {
	log := logf.FromContext(ctx)
	databaseNetworkConfig, err := dbManager.GetDatabaseNetworkConfiguration(ctx, postgresql.Status.ScalingoDatabaseID)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "get database network configuration id %s", postgresql.Status.ScalingoDatabaseID)
	}

	log.Info("Create Outscale NetPeeringRequest resource", "outscaleNetID", databaseNetworkConfig.OutscaleNetID, "outscaleAccountID", databaseNetworkConfig.OutscaleAccountID)
	now := time.Now().UnixMilli()
	resourceName := fmt.Sprintf("%s-%s-%d", postgresql.Name, netPeeringRequestSuffix, now)
	netPeeringRequest := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "oks.dev/v1beta",
			"kind":       "NetPeeringRequest",
			"metadata": map[string]any{
				"name":      resourceName,
				"namespace": postgresql.Namespace,
				"labels": map[string]any{
					"app.kubernetes.io/managed-by": "scalingo-operator",
					"scalingo.com/database-id":     postgresql.Status.ScalingoDatabaseID,
				},
			},
			"spec": map[string]any{
				"accepterNetId":   databaseNetworkConfig.OutscaleNetID,
				"accepterOwnerId": databaseNetworkConfig.OutscaleAccountID,
			},
		},
	}

	err = controllerutil.SetControllerReference(&postgresql, netPeeringRequest, r.Scheme)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "set controller reference")
	}

	err = r.Create(ctx, netPeeringRequest)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "create net peering request")
	}
	log.Info("Outscale NetPeeringRequest resource created")

	return netPeeringRequest, nil
}

func (r *PostgreSQLReconciler) deleteNetPeerings(ctx context.Context, dbManager databaseusecases.Manager, postgresql apiv1.PostgreSQL) error {
	log := logf.FromContext(ctx)
	if postgresql.Status.ScalingoDatabaseID == "" {
		return nil
	}

	netPeerings, err := r.listExistingNetPeeringsForDatabase(ctx, dbManager, postgresql)
	if err != nil {
		return errors.Wrap(ctx, err, "list existing net peerings for database")
	}
	// No peering to cleanup, skip deletion.
	if len(netPeerings) == 0 {
		return nil
	}

	for _, netPeering := range netPeerings {
		log = log.WithValues("netPeering", netPeering.GetName())
		log.Info("Delete Outscale NetPeering resource")
		err := r.Delete(ctx, &netPeering)
		if err != nil {
			log.Error(err, "Delete net peering resource", "netPeering", netPeering.GetName())
		}
		log.Info("Outscale NetPeering resource deleted")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.PostgreSQL{}).
		Named("postgresql").
		Complete(r)
}

func (r *PostgreSQLReconciler) listExistingNetPeeringRequestsForDatabase(ctx context.Context, postgresql apiv1.PostgreSQL) ([]unstructured.Unstructured, error) {
	netPeeringRequests := &unstructured.UnstructuredList{}
	netPeeringRequests.SetGroupVersionKind(netPeeringRequestListGVK)

	err := r.List(
		ctx,
		netPeeringRequests,
		client.InNamespace(postgresql.Namespace),
		client.MatchingLabels{"scalingo.com/database-id": postgresql.Status.ScalingoDatabaseID},
	)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "list net peering requests")
	}

	return netPeeringRequests.Items, nil
}

func (r *PostgreSQLReconciler) listExistingNetPeeringsForDatabase(ctx context.Context, dbManager databaseusecases.Manager, postgresql apiv1.PostgreSQL) ([]unstructured.Unstructured, error) {
	netPeerings := &unstructured.UnstructuredList{}
	netPeerings.SetGroupVersionKind(netPeeringListGVK)
	err := r.List(
		ctx,
		netPeerings,
		client.InNamespace(postgresql.Namespace),
	)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "list net peerings")
	}

	databaseNetworkConfig, err := dbManager.GetDatabaseNetworkConfiguration(ctx, postgresql.Status.ScalingoDatabaseID)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "get database network configuration id %s", postgresql.Status.ScalingoDatabaseID)
	}

	// Select the net peering from current OKS cluster joined with the database.
	// This should be unique since there can't be 2 net peerings with the same accepterNetId and accepterOwnerId for a given database.
	netPeeringsForDatabase := make([]unstructured.Unstructured, 0, 1)
	for _, netPeering := range netPeerings.Items {
		npState, _, _ := unstructured.NestedString(netPeering.Object, "status", "netPeeringState")
		npAccepterNetId, _, _ := unstructured.NestedString(netPeering.Object, "status", "accepterNetId")
		npAccepterOwnerId, _, _ := unstructured.NestedString(netPeering.Object, "status", "accepterOwnerId")

		if npState == "active" && npAccepterNetId == databaseNetworkConfig.OutscaleNetID && npAccepterOwnerId == databaseNetworkConfig.OutscaleAccountID {
			netPeeringsForDatabase = append(netPeeringsForDatabase, netPeering)
		}
	}

	return netPeeringsForDatabase, nil
}

func extractNetPeeringID(netPeeringRequest unstructured.Unstructured) string {
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
