package networking

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Scalingo/go-utils/errors/v3"
	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/controller/helpers"
	databaseusecases "github.com/Scalingo/scalingo-operator/internal/usecases/database"
)

type NetPeeringReconciler struct {
	client.Client

	Scheme *runtime.Scheme
}

type DatabaseResource struct {
	Name       string
	Namespace  string
	Owner      client.Object
	DatabaseID string
	Networking apiv1.NetworkingSpec
}

type DatabaseState struct {
	DeletionRequested bool
	Available         bool
	Provisioning      bool
}

func (r NetPeeringReconciler) Reconcile(ctx context.Context, dbManager databaseusecases.Manager, resource DatabaseResource, state DatabaseState) (time.Duration, error) {
	if state.DeletionRequested {
		return 0, nil
	}

	log := logf.FromContext(ctx)
	if !resource.Networking.IsOutscaleOKSNetPeeringEnabled() {
		err := r.DeleteNetPeerings(ctx, dbManager, resource)
		if err != nil {
			return 0, errors.Wrap(ctx, err, "delete net peering resources")
		}
		return 0, nil
	}

	if !state.Available || state.Provisioning || resource.DatabaseID == "" {
		return 0, nil
	}

	log.Info("Reconcile Outscale OKS net peering")
	netPeeringID, err := r.ensureOKSNetPeeringRequest(ctx, dbManager, resource)
	if err != nil {
		return 0, errors.Wrap(ctx, err, "reconcile net peering request")
	}

	if netPeeringID == "" {
		log.Info("Waiting for NetPeeringRequest to be provisioned")
		return helpers.RequeueLongDelay, nil
	}

	err = dbManager.EnsureDatabaseNetPeering(ctx, resource.DatabaseID, netPeeringID)
	if err != nil {
		return 0, errors.Wrapf(ctx, err, "ensure database net peering id %s", resource.DatabaseID)
	}
	return 0, nil
}

func (r NetPeeringReconciler) DeleteNetPeerings(ctx context.Context, dbManager databaseusecases.Manager, resource DatabaseResource) error {
	log := logf.FromContext(ctx)
	if resource.DatabaseID == "" {
		return nil
	}

	netPeerings, err := r.listExistingNetPeeringsForDatabase(ctx, dbManager, resource)
	if err != nil {
		return errors.Wrap(ctx, err, "list existing net peerings for database")
	}

	for _, netPeering := range netPeerings {
		log = log.WithValues("netPeering", netPeering.Name())
		log.Info("Delete Outscale NetPeering resource")
		err := r.Delete(ctx, netPeering.Object())
		if err != nil {
			log.Error(err, "Delete net peering resource", "netPeering", netPeering.Name())
		}
		log.Info("Outscale NetPeering resource deleted")
	}

	return nil
}

func (r NetPeeringReconciler) ensureOKSNetPeeringRequest(ctx context.Context, dbManager databaseusecases.Manager, resource DatabaseResource) (string, error) {
	existingNetPeeringRequestsForDatabase, err := r.listExistingNetPeeringRequestsForDatabase(ctx, resource)
	if err != nil {
		return "", errors.Wrap(ctx, err, "list existing net peering requests for database")
	}

	var netPeeringRequest NetPeeringRequest
	switch len(existingNetPeeringRequestsForDatabase) {
	case 0:
		netPeeringRequest, err = r.createOKSNetPeeringRequest(ctx, dbManager, resource)
		if err != nil {
			return "", errors.Wrap(ctx, err, "create net peering request")
		}
	case 1:
		netPeeringRequest = existingNetPeeringRequestsForDatabase[0]
	default:
		return "", errors.New(ctx, "multiple net peering requests found for database")
	}

	netPeeringID := netPeeringRequest.NetPeeringID()
	return netPeeringID, nil
}

func (r NetPeeringReconciler) createOKSNetPeeringRequest(ctx context.Context, dbManager databaseusecases.Manager, resource DatabaseResource) (NetPeeringRequest, error) {
	log := logf.FromContext(ctx)
	databaseNetworkConfig, err := dbManager.GetDatabaseNetworkConfiguration(ctx, resource.DatabaseID)
	if err != nil {
		return NetPeeringRequest{}, errors.Wrapf(ctx, err, "get database network configuration id %s", resource.DatabaseID)
	}

	log.Info("Create Outscale NetPeeringRequest resource", "outscaleNetID", databaseNetworkConfig.OutscaleNetID, "outscaleAccountID", databaseNetworkConfig.OutscaleAccountID)
	netPeeringRequest := NewNetPeeringRequest(resource, databaseNetworkConfig)

	err = netPeeringRequest.SetControllerReference(resource.Owner, r.Scheme)
	if err != nil {
		return NetPeeringRequest{}, errors.Wrap(ctx, err, "set controller reference")
	}

	err = r.Create(ctx, netPeeringRequest.Object())
	if err != nil {
		return NetPeeringRequest{}, errors.Wrap(ctx, err, "create net peering request")
	}
	log.Info("Outscale NetPeeringRequest resource created")

	return netPeeringRequest, nil
}

func (r NetPeeringReconciler) listExistingNetPeeringRequestsForDatabase(ctx context.Context, resource DatabaseResource) ([]NetPeeringRequest, error) {
	netPeeringRequests := NewNetPeeringRequestList()

	err := r.List(
		ctx,
		netPeeringRequests.ObjectList(),
		client.InNamespace(resource.Namespace),
		NetPeeringRequestMatchingLabels(resource.DatabaseID),
	)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "list net peering requests")
	}

	return netPeeringRequests.Items(), nil
}

func (r NetPeeringReconciler) listExistingNetPeeringsForDatabase(ctx context.Context, dbManager databaseusecases.Manager, resource DatabaseResource) ([]NetPeering, error) {
	netPeerings := NewNetPeeringList()
	err := r.List(
		ctx,
		netPeerings.ObjectList(),
		client.InNamespace(resource.Namespace),
	)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "list net peerings")
	}

	databaseNetworkConfig, err := dbManager.GetDatabaseNetworkConfiguration(ctx, resource.DatabaseID)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "get database network configuration id %s", resource.DatabaseID)
	}

	return netPeerings.ItemsMatching(databaseNetworkConfig), nil
}
