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
	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/controller/helpers"
	"github.com/Scalingo/scalingo-operator/internal/controller/networking"
	"github.com/Scalingo/scalingo-operator/internal/domain"
	databaseusecases "github.com/Scalingo/scalingo-operator/internal/usecases/database"
	databasebase "github.com/Scalingo/scalingo-operator/internal/usecases/database/base"
)

type dedicatedDatabaseResource interface {
	object() client.Object
	meta() *metav1.ObjectMeta
	toDatabase(ctx context.Context) (domain.Database, error)
	authSecret() apiv1.AuthSecretSpec
	connInfoSecretTarget() apiv1.SecretTargetSpec
	networking() apiv1.NetworkingSpec
	region() string
	databaseID() string
	setDatabaseID(id string)
	conditions() *[]metav1.Condition
}

type databaseSecretWriter interface {
	SetSecret(ctx context.Context, secret domain.Secret) error
}

type dedicatedDatabaseConfig struct {
	newResource   func() dedicatedDatabaseResource
	finalizerName string
	databaseType  domain.DatabaseType
}

type dedicatedDatabaseReconciler struct {
	client.Client

	Scheme *runtime.Scheme
	config dedicatedDatabaseConfig
}

type dedicatedDatabaseState struct {
	available         bool
	provisioning      bool
	running           bool
	deletionRequested bool
}

type dedicatedDatabaseResult struct {
	statusUpdate bool
	requeueAfter time.Duration
}

func (r *dedicatedDatabaseReconciler) reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	resource := r.config.newResource()

	err := r.Get(ctx, req.NamespacedName, resource.object())
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	initialized, err := r.initializeResource(ctx, resource)
	if err != nil {
		return ctrl.Result{}, err
	}
	if initialized {
		return ctrl.Result{RequeueAfter: helpers.RequeueShortDelay}, nil
	}

	secretManager := helpers.NewSecretManager(r.Client, resource.object())
	authSecret := resource.authSecret()
	authSecretRef := domain.Secret{Namespace: req.Namespace, Name: authSecret.Name, Key: authSecret.Key}
	log.Info("Get auth secret", "secret", authSecretRef)

	apiToken, err := secretManager.GetSecret(ctx, authSecretRef)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "get auth secret")
	}

	dbManager, err := databasebase.NewManager(ctx, r.config.databaseType, apiToken, resource.region())
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "create database manager")
	}

	expectedDB, err := resource.toDatabase(ctx)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "bad custom resource format")
	}

	state := databaseState(resource)
	log.Info("Current state",
		"database", resource.databaseID(),
		"deletion_requested", state.deletionRequested,
		"available", state.available,
		"provisioning", state.provisioning,
		"running", state.running)

	result, err := r.reconcileDatabase(ctx, req.Namespace, resource, dbManager, secretManager, expectedDB, state)
	if err != nil {
		return ctrl.Result{}, err
	}

	state = databaseState(resource)
	networkingSpec := resource.networking()
	netPeeringReconciler := networking.NetPeeringReconciler{Client: r.Client, Scheme: r.Scheme}
	netPeeringRequeue, err := netPeeringReconciler.Reconcile(
		ctx,
		dbManager,
		networking.DatabaseResource{
			Name:       resource.object().GetName(),
			Namespace:  resource.object().GetNamespace(),
			Owner:      resource.object(),
			DatabaseID: resource.databaseID(),
			Networking: networkingSpec,
		},
		networking.DatabaseState{
			DeletionRequested: state.deletionRequested,
			Available:         state.available,
			Provisioning:      state.provisioning,
		},
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	if netPeeringRequeue > 0 {
		result.requeueAfter = netPeeringRequeue
	}

	if result.statusUpdate {
		log.Info("Update resource status", "statusConditions", *resource.conditions())
		err = r.Status().Update(ctx, resource.object())
		if err != nil {
			return ctrl.Result{}, errors.Wrap(ctx, err, "update database resource status")
		}
	}
	if result.requeueAfter > 0 {
		log.Info("Requeue after", "delay", result.requeueAfter)
		return ctrl.Result{RequeueAfter: result.requeueAfter}, nil
	}

	log.Info("Ready")
	return ctrl.Result{}, nil
}

func (r *dedicatedDatabaseReconciler) initializeResource(ctx context.Context, resource dedicatedDatabaseResource) (bool, error) {
	log := logf.FromContext(ctx)
	object := resource.object()
	meta := resource.meta()
	conditions := resource.conditions()

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
	case !controllerutil.ContainsFinalizer(object, r.config.finalizerName):
		log.Info("Add finalizer to resource", "finalizer", r.config.finalizerName)
		controllerutil.AddFinalizer(object, r.config.finalizerName)
		err := r.Update(ctx, object)
		if err != nil {
			return false, errors.Wrap(ctx, err, "add resource finalizer")
		}
	case !helpers.IsDatabaseInitialized(*conditions):
		log.Info("Initialize resource status conditions")
		helpers.SetDatabaseInitialStatus(conditions)
		err := r.Status().Update(ctx, object)
		if err != nil {
			return false, errors.Wrap(ctx, err, "update database resource status")
		}
	case !metav1.HasAnnotation(*meta, helpers.DatabaseAnnotationIsRunning):
		log.Info("Initialize resource annotations")
		helpers.SetDatabaseIsNotRunning(meta)
		err := r.Update(ctx, object)
		if err != nil {
			return false, errors.Wrap(ctx, err, "update database resource")
		}
	case helpers.IsDatabaseAvailable(*conditions) && !helpers.IsDatabaseRunning(*meta):
		log.Info("Update database resource running annotation")
		helpers.SetDatabaseIsRunning(meta)
		err := r.Update(ctx, object)
		if err != nil {
			return false, errors.Wrap(ctx, err, "update database resource")
		}
	default:
		return false, nil
	}

	return true, nil
}

func databaseState(resource dedicatedDatabaseResource) dedicatedDatabaseState {
	return dedicatedDatabaseState{
		available:         helpers.IsDatabaseAvailable(*resource.conditions()),
		provisioning:      helpers.IsDatabaseProvisioning(*resource.conditions()),
		running:           helpers.IsDatabaseRunning(*resource.meta()),
		deletionRequested: helpers.IsDatabaseDeletionRequested(*resource.meta()),
	}
}

func (r *dedicatedDatabaseReconciler) reconcileDatabase(
	ctx context.Context,
	namespace string,
	resource dedicatedDatabaseResource,
	dbManager databaseusecases.Manager,
	secretWriter databaseSecretWriter,
	expectedDB domain.Database,
	state dedicatedDatabaseState,
) (dedicatedDatabaseResult, error) {
	databaseID := resource.databaseID()

	switch {
	case state.deletionRequested:
		err := r.deleteDatabase(ctx, resource, dbManager)
		return dedicatedDatabaseResult{}, err
	case !state.available && databaseID == "":
		return r.createDatabase(ctx, resource, dbManager, expectedDB)
	case state.available && !state.provisioning && databaseID != "":
		return r.updateDatabase(ctx, resource, dbManager, expectedDB)
	case state.provisioning && databaseID != "":
		return r.reconcileDatabaseProvisioning(ctx, namespace, resource, dbManager, secretWriter, expectedDB, state.available)
	default:
		return dedicatedDatabaseResult{}, nil
	}
}

func (r *dedicatedDatabaseReconciler) deleteDatabase(
	ctx context.Context,
	resource dedicatedDatabaseResource,
	dbManager databaseusecases.Manager,
) error {
	log := logf.FromContext(ctx)
	databaseID := resource.databaseID()
	log.Info("Delete database")

	if databaseID == "" {
		log.Info("Database provisioning requested but no database created yet, skip database deletion")
	} else {
		err := r.deleteScalingoDatabase(ctx, resource, dbManager, databaseID)
		if err != nil {
			return err
		}
	}

	controllerutil.RemoveFinalizer(resource.object(), r.config.finalizerName)
	err := r.Update(ctx, resource.object())
	if err != nil {
		return errors.Wrap(ctx, err, "remove resource finalizer")
	}
	return nil
}

func (r *dedicatedDatabaseReconciler) deleteScalingoDatabase(
	ctx context.Context,
	resource dedicatedDatabaseResource,
	dbManager databaseusecases.Manager,
	databaseID string,
) error {
	log := logf.FromContext(ctx)
	exists, err := dbManager.CheckDatabaseExists(ctx, databaseID)
	if err != nil {
		return errors.Wrap(ctx, err, "check database exists")
	}
	if !exists {
		log.Info("Scalingo database not found, skip database deletion", "database", databaseID)
		return nil
	}

	networkingSpec := resource.networking()
	if networkingSpec.IsOutscaleOKSNetPeeringEnabled() {
		netPeeringReconciler := networking.NetPeeringReconciler{Client: r.Client, Scheme: r.Scheme}
		err = netPeeringReconciler.DeleteNetPeerings(ctx, dbManager, networking.DatabaseResource{
			Name:       resource.object().GetName(),
			Namespace:  resource.object().GetNamespace(),
			Owner:      resource.object(),
			DatabaseID: databaseID,
			Networking: networkingSpec,
		})
		if err != nil {
			return errors.Wrap(ctx, err, "delete net peering resources")
		}
	}

	err = dbManager.DeleteDatabase(ctx, databaseID)
	if err != nil {
		return errors.Wrap(ctx, err, "delete database")
	}
	return nil
}

func (r *dedicatedDatabaseReconciler) createDatabase(
	ctx context.Context,
	resource dedicatedDatabaseResource,
	dbManager databaseusecases.Manager,
	expectedDB domain.Database,
) (dedicatedDatabaseResult, error) {
	log := logf.FromContext(ctx)
	log.Info("Create database")

	newDB, err := dbManager.CreateDatabase(ctx, expectedDB)
	if err != nil {
		log.Error(err, "Create database", "database", expectedDB)
		return dedicatedDatabaseResult{}, errors.Wrap(ctx, err, "create database")
	}

	resource.setDatabaseID(newDB.ID)
	helpers.SetDatabaseStatusProvisioning(resource.conditions())
	return dedicatedDatabaseResult{statusUpdate: true, requeueAfter: helpers.RequeueLongDelay}, nil
}

func (r *dedicatedDatabaseReconciler) updateDatabase(
	ctx context.Context,
	resource dedicatedDatabaseResource,
	dbManager databaseusecases.Manager,
	expectedDB domain.Database,
) (dedicatedDatabaseResult, error) {
	log := logf.FromContext(ctx)
	log.Info("Update database")

	dbStatus, err := dbManager.UpdateDatabase(ctx, resource.databaseID(), expectedDB)
	if err != nil {
		log.Error(err, "Update database", "database", expectedDB)
		return dedicatedDatabaseResult{}, errors.Wrap(ctx, err, "update database")
	}
	if dbStatus != domain.DatabaseStatusProvisioning {
		return dedicatedDatabaseResult{}, nil
	}

	log.Info("Waiting for database being provisioned")
	helpers.SetDatabaseStatusProvisioning(resource.conditions())
	return dedicatedDatabaseResult{statusUpdate: true}, nil
}

func (r *dedicatedDatabaseReconciler) reconcileDatabaseProvisioning(
	ctx context.Context,
	namespace string,
	resource dedicatedDatabaseResource,
	dbManager databaseusecases.Manager,
	secretWriter databaseSecretWriter,
	expectedDB domain.Database,
	available bool,
) (dedicatedDatabaseResult, error) {
	log := logf.FromContext(ctx)
	databaseID := resource.databaseID()

	if available {
		_, err := dbManager.UpdateDatabase(ctx, databaseID, expectedDB)
		if err != nil {
			log.Error(err, "Update database while provisioning", "database", databaseID)
			return dedicatedDatabaseResult{}, errors.Wrap(ctx, err, "update database while provisioning")
		}
	}

	currentDB, err := dbManager.GetDatabase(ctx, databaseID)
	if err != nil {
		return dedicatedDatabaseResult{}, errors.Wrap(ctx, err, "get current database")
	}
	if currentDB.Status != domain.DatabaseStatusRunning {
		log.Info("Waiting for database being provisioned")
		return dedicatedDatabaseResult{requeueAfter: helpers.RequeueLongDelay}, nil
	}

	log.Info("Database is provisioned")
	helpers.SetDatabaseStatusProvisioned(resource.conditions())
	err = r.writeConnectionSecrets(ctx, namespace, resource, dbManager, secretWriter, currentDB)
	if err != nil {
		return dedicatedDatabaseResult{}, err
	}

	return dedicatedDatabaseResult{statusUpdate: true, requeueAfter: helpers.RequeueShortDelay}, nil
}

func (r *dedicatedDatabaseReconciler) writeConnectionSecrets(
	ctx context.Context,
	namespace string,
	resource dedicatedDatabaseResource,
	dbManager databaseusecases.Manager,
	secretWriter databaseSecretWriter,
	currentDB domain.Database,
) error {
	log := logf.FromContext(ctx)
	dbURL, err := dbManager.GetDatabaseURL(ctx, currentDB)
	if err != nil {
		return errors.Wrap(ctx, err, "get database url")
	}

	secretTarget := resource.connInfoSecretTarget()
	connectionSecret := domain.Secret{
		Namespace: namespace,
		Name:      secretTarget.Name,
		Key:       domain.ComposeConnectionURLName(secretTarget.Prefix, dbURL.Name),
		Value:     dbURL.Value,
	}
	log.Info("Write connection info secret", "secret", connectionSecret)
	err = secretWriter.SetSecret(ctx, connectionSecret)
	if err != nil {
		return errors.Wrap(ctx, err, "set secret")
	}

	endpoints, err := dbManager.GetDatabaseEndpoints(ctx, currentDB.ID)
	if err != nil {
		return errors.Wrap(ctx, err, "get database endpoints")
	}
	for _, endpoint := range endpoints {
		endpointURL, err := domain.ComposeEndpointConnectionURL(ctx, dbURL.Value, endpoint)
		if err != nil {
			return errors.Wrap(ctx, err, "compose endpoint connection url")
		}

		endpointSecret := domain.Secret{
			Namespace: namespace,
			Name:      secretTarget.Name,
			Key:       domain.ComposeEndpointConnectionURLName(secretTarget.Prefix, dbURL.Name, endpoint.Type),
			Value:     endpointURL,
		}
		log.Info("Write endpoint connection info secret", "secret", endpointSecret)
		err = secretWriter.SetSecret(ctx, endpointSecret)
		if err != nil {
			return errors.Wrap(ctx, err, "set secret")
		}
	}
	return nil
}
