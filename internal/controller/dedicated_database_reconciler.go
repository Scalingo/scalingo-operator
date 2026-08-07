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
	Object() client.Object
	Meta() *metav1.ObjectMeta
	AuthSecret() apiv1.AuthSecretSpec
	ConnInfoSecretTarget() apiv1.SecretTargetSpec
	Networking() apiv1.NetworkingSpec
	Region() string
	DatabaseID() string
	SetDatabaseID(string)
	Conditions() *[]metav1.Condition
}

type databaseSecretWriter interface {
	SetSecret(context.Context, domain.Secret) error
}

type dedicatedDatabaseConfig struct {
	newResource   func() dedicatedDatabaseResource
	finalizerName string
	databaseType  domain.DatabaseType
	toDatabase    func(context.Context, dedicatedDatabaseResource) (domain.Database, error)
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

func (r *dedicatedDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	resource := r.config.newResource()

	err := r.Get(ctx, req.NamespacedName, resource.Object())
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

	secretManager := helpers.NewSecretManager(r.Client, resource.Object())
	authSecret := resource.AuthSecret()
	authSecretRef := domain.Secret{Namespace: req.Namespace, Name: authSecret.Name, Key: authSecret.Key}
	log.Info("Get auth secret", "secret", authSecretRef)

	apiToken, err := secretManager.GetSecret(ctx, authSecretRef)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "get auth secret")
	}

	dbManager, err := databasebase.NewManager(ctx, r.config.databaseType, apiToken, resource.Region())
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "create database manager")
	}

	expectedDB, err := r.config.toDatabase(ctx, resource)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(ctx, err, "bad custom resource format")
	}

	state := databaseState(resource)
	log.Info("Current state",
		"database", resource.DatabaseID(),
		"deletion_requested", state.deletionRequested,
		"available", state.available,
		"provisioning", state.provisioning,
		"running", state.running)

	result, err := r.reconcileDatabase(ctx, req.Namespace, resource, dbManager, secretManager, expectedDB, state)
	if err != nil {
		return ctrl.Result{}, err
	}

	state = databaseState(resource)
	networkingSpec := resource.Networking()
	netPeeringReconciler := networking.NetPeeringReconciler{Client: r.Client, Scheme: r.Scheme}
	netPeeringRequeue, err := netPeeringReconciler.Reconcile(
		ctx,
		dbManager,
		networking.DatabaseResource{
			Name:       resource.Object().GetName(),
			Namespace:  resource.Object().GetNamespace(),
			Owner:      resource.Object(),
			DatabaseID: resource.DatabaseID(),
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
		log.Info("Update resource status", "statusConditions", *resource.Conditions())
		err = r.Status().Update(ctx, resource.Object())
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
	object := resource.Object()
	meta := resource.Meta()
	conditions := resource.Conditions()

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
		available:         helpers.IsDatabaseAvailable(*resource.Conditions()),
		provisioning:      helpers.IsDatabaseProvisioning(*resource.Conditions()),
		running:           helpers.IsDatabaseRunning(*resource.Meta()),
		deletionRequested: helpers.IsDatabaseDeletionRequested(*resource.Meta()),
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
	databaseID := resource.DatabaseID()

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
	databaseID := resource.DatabaseID()
	log.Info("Delete database")

	if databaseID == "" {
		log.Info("Database provisioning requested but no database created yet, skip database deletion")
	} else {
		exists, err := dbManager.CheckDatabaseExists(ctx, databaseID)
		if err != nil {
			return errors.Wrapf(ctx, err, "check database %s exists", databaseID)
		}
		if !exists {
			log.Info("Scalingo database not found, skip database deletion", "database", databaseID)
		} else {
			networkingSpec := resource.Networking()
			if networkingSpec.IsOutscaleOKSNetPeeringEnabled() {
				netPeeringReconciler := networking.NetPeeringReconciler{Client: r.Client, Scheme: r.Scheme}
				err = netPeeringReconciler.DeleteNetPeerings(ctx, dbManager, networking.DatabaseResource{
					Name:       resource.Object().GetName(),
					Namespace:  resource.Object().GetNamespace(),
					Owner:      resource.Object(),
					DatabaseID: databaseID,
					Networking: networkingSpec,
				})
				if err != nil {
					return errors.Wrap(ctx, err, "delete net peering resources")
				}
			}

			err = dbManager.DeleteDatabase(ctx, databaseID)
			if err != nil {
				return errors.Wrapf(ctx, err, "delete database id %s", databaseID)
			}
		}
	}

	controllerutil.RemoveFinalizer(resource.Object(), r.config.finalizerName)
	err := r.Update(ctx, resource.Object())
	if err != nil {
		return errors.Wrap(ctx, err, "remove resource finalizer")
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
		return dedicatedDatabaseResult{}, errors.Wrapf(ctx, err, "create database %s", expectedDB.Name)
	}

	resource.SetDatabaseID(newDB.ID)
	helpers.SetDatabaseStatusProvisioning(resource.Conditions())
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

	dbStatus, err := dbManager.UpdateDatabase(ctx, resource.DatabaseID(), expectedDB)
	if err != nil {
		log.Error(err, "Update database", "database", expectedDB)
		return dedicatedDatabaseResult{}, errors.Wrapf(ctx, err, "update database %s", expectedDB.Name)
	}
	if dbStatus != domain.DatabaseStatusProvisioning {
		return dedicatedDatabaseResult{}, nil
	}

	log.Info("Waiting for database being provisioned")
	helpers.SetDatabaseStatusProvisioning(resource.Conditions())
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
	databaseID := resource.DatabaseID()

	if available {
		_, err := dbManager.UpdateDatabase(ctx, databaseID, expectedDB)
		if err != nil {
			log.Error(err, "Update database while provisioning", "database", databaseID)
			return dedicatedDatabaseResult{}, errors.Wrapf(ctx, err, "update database %s while provisioning", expectedDB.Name)
		}
	}

	currentDB, err := dbManager.GetDatabase(ctx, databaseID)
	if err != nil {
		return dedicatedDatabaseResult{}, errors.Wrapf(ctx, err, "get current database %s", databaseID)
	}
	if currentDB.Status != domain.DatabaseStatusRunning {
		log.Info("Waiting for database being provisioned")
		return dedicatedDatabaseResult{requeueAfter: helpers.RequeueLongDelay}, nil
	}

	log.Info("Database is provisioned")
	helpers.SetDatabaseStatusProvisioned(resource.Conditions())
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

	secretTarget := resource.ConnInfoSecretTarget()
	connectionSecret := domain.Secret{
		Namespace: namespace,
		Name:      secretTarget.Name,
		Key:       domain.ComposeConnectionURLName(secretTarget.Prefix, dbURL.Name),
		Value:     dbURL.Value,
	}
	log.Info("Write connection info secret", "secret", connectionSecret)
	err = secretWriter.SetSecret(ctx, connectionSecret)
	if err != nil {
		return errors.Wrapf(ctx, err, "set secret %s", connectionSecret.Key)
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
			return errors.Wrapf(ctx, err, "set secret %s", endpointSecret.Key)
		}
	}
	return nil
}
