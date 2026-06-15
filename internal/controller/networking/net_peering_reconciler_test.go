package networking

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Scalingo/go-utils/errors/v3"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/controller/helpers"
	"github.com/Scalingo/scalingo-operator/internal/domain"
	"github.com/Scalingo/scalingo-operator/internal/usecases/database/databasemock"
)

func TestReconcileOKSNetPeering(t *testing.T) {
	t.Run("ensures database net peering from provisioned request", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)

		clientStub := &netPeeringResourceClient{
			items: []*unstructured.Unstructured{
				newNetPeeringRequest("net-peering-request", "pcx-1234"),
			},
		}
		databaseManager := databasemock.NewMockManager(ctrl)

		reconciler := NetPeeringReconciler{
			Client: clientStub,
		}
		resource := DatabaseResource{
			Name:       "db-resource",
			Namespace:  "default",
			DatabaseID: "db-123",
			Networking: netPeeringEnabledNetworkingSpec(),
		}

		databaseManager.EXPECT().GetDatabaseNetworkConfiguration(ctx, "db-123").Return(domain.DatabaseNetworkConfiguration{
			OutscaleNetID:     "net-id",
			OutscaleAccountID: "owner-id",
		}, nil)
		databaseManager.EXPECT().EnsureDatabaseNetPeering(ctx, "db-123", "pcx-1234").Return(nil)

		requeue, err := reconciler.Reconcile(ctx, databaseManager, resource, DatabaseState{Available: true})

		require.NoError(t, err)
		require.Zero(t, requeue)
	})

	t.Run("requeues while request is not provisioned", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)

		clientStub := &netPeeringResourceClient{
			items: []*unstructured.Unstructured{
				newNetPeeringRequest("net-peering-request", ""),
			},
		}
		databaseManager := databasemock.NewMockManager(ctrl)

		reconciler := NetPeeringReconciler{
			Client: clientStub,
		}
		resource := DatabaseResource{
			Name:       "db-resource",
			Namespace:  "default",
			DatabaseID: "db-123",
			Networking: netPeeringEnabledNetworkingSpec(),
		}

		databaseManager.EXPECT().GetDatabaseNetworkConfiguration(ctx, "db-123").Return(domain.DatabaseNetworkConfiguration{
			OutscaleNetID:     "net-id",
			OutscaleAccountID: "owner-id",
		}, nil)

		requeue, err := reconciler.Reconcile(ctx, databaseManager, resource, DatabaseState{Available: true})

		require.NoError(t, err)
		require.Equal(t, helpers.RequeueLongDelay, requeue)
	})

	t.Run("ensures database net peering from existing active net peering", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)

		clientStub := &netPeeringResourceClient{
			items: []*unstructured.Unstructured{
				newNetPeering("default", "pcx-1234", netPeeringStatusStateActive, "net-id"),
				newNetPeeringRequest("net-peering-request-a", "pcx-5678"),
				newNetPeeringRequest("net-peering-request-b", "pcx-9012"),
			},
		}
		databaseManager := databasemock.NewMockManager(ctrl)

		reconciler := NetPeeringReconciler{
			Client: clientStub,
		}
		resource := DatabaseResource{
			Name:       "db-resource",
			Namespace:  "default",
			DatabaseID: "db-123",
			Networking: netPeeringEnabledNetworkingSpec(),
		}

		databaseManager.EXPECT().GetDatabaseNetworkConfiguration(ctx, "db-123").Return(domain.DatabaseNetworkConfiguration{
			OutscaleNetID:     "net-id",
			OutscaleAccountID: "owner-id",
		}, nil)
		databaseManager.EXPECT().EnsureDatabaseNetPeering(ctx, "db-123", "pcx-1234").Return(nil)

		requeue, err := reconciler.Reconcile(ctx, databaseManager, resource, DatabaseState{Available: true})

		require.NoError(t, err)
		require.Zero(t, requeue)
	})
}

func TestDeleteOKSNetPeerings(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctrl := gomock.NewController(t)

	clientStub := &netPeeringResourceClient{
		items: []*unstructured.Unstructured{
			newNetPeering("default", "net-peering-a", netPeeringStatusStateActive, "net-id"),
			newNetPeering("default", "net-peering-b", netPeeringStatusStateActive, "net-id"),
			newNetPeering("default", "net-peering-inactive", "pending", "net-id"),
			newNetPeering("default", "net-peering-other-network", netPeeringStatusStateActive, "other-net-id"),
			newNetPeering("other", "net-peering-other-namespace", netPeeringStatusStateActive, "net-id"),
		},
	}
	databaseManager := databasemock.NewMockManager(ctrl)

	reconciler := NetPeeringReconciler{
		Client: clientStub,
	}
	resource := DatabaseResource{
		Name:       "db-resource",
		Namespace:  "default",
		DatabaseID: "db-123",
	}

	databaseManager.EXPECT().GetDatabaseNetworkConfiguration(ctx, "db-123").Return(domain.DatabaseNetworkConfiguration{
		OutscaleNetID:     "net-id",
		OutscaleAccountID: "owner-id",
	}, nil)

	err := reconciler.DeleteNetPeerings(ctx, databaseManager, resource)
	require.NoError(t, err)
	require.Len(t, clientStub.items, 3)
	require.ElementsMatch(
		t,
		[]string{"net-peering-inactive", "net-peering-other-network", "net-peering-other-namespace"},
		resourceNames(clientStub.items),
	)
}

type netPeeringResourceClient struct {
	client.Client

	items []*unstructured.Unstructured
}

func (c *netPeeringResourceClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	listOptions := &client.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(listOptions)
	}

	typedList, ok := list.(*unstructured.UnstructuredList)
	if !ok {
		return errors.New(ctx, "expected unstructured list")
	}
	typedList.Items = nil

	for _, item := range c.items {
		if listOptions.Namespace != "" && item.GetNamespace() != listOptions.Namespace {
			continue
		}
		if listOptions.LabelSelector != nil && !listOptions.LabelSelector.Matches(labels.Set(item.GetLabels())) {
			continue
		}

		typedList.Items = append(typedList.Items, *item.DeepCopy())
	}

	return nil
}

func (c *netPeeringResourceClient) Delete(_ context.Context, obj client.Object, _ ...client.DeleteOption) error {
	filteredItems := make([]*unstructured.Unstructured, 0, len(c.items))
	for _, item := range c.items {
		if item.GetNamespace() == obj.GetNamespace() && item.GetName() == obj.GetName() {
			continue
		}

		filteredItems = append(filteredItems, item)
	}

	c.items = filteredItems
	return nil
}

func newNetPeering(namespace, name, state, accepterNetID string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(helpers.OutscaleNetPeeringGVK)
	object.SetNamespace(namespace)
	object.SetName(name)
	object.Object[netPeeringStatusField] = map[string]any{
		netPeeringStatusStateField:     state,
		netPeeringAccepterNetIDField:   accepterNetID,
		netPeeringAccepterOwnerIDField: "owner-id",
	}

	return object
}

func newNetPeeringRequest(name, netPeeringID string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(helpers.OutscaleNetPeeringRequestGVK)
	object.SetNamespace("default")
	object.SetName(name)
	object.SetLabels(map[string]string{netPeeringRequestDatabaseIDLabel: "db-123"})
	object.Object[netPeeringRequestStatusField] = map[string]any{
		netPeeringRequestNetPeeringIDField: netPeeringID,
	}

	return object
}

func netPeeringEnabledNetworkingSpec() apiv1.NetworkingSpec {
	return apiv1.NetworkingSpec{
		Outscale: &apiv1.OutscaleSpec{
			OKS: &apiv1.OutscaleOKSSpec{
				NetPeering: true,
			},
		},
	}
}

func resourceNames(objects []*unstructured.Unstructured) []string {
	names := make([]string, 0, len(objects))
	for _, object := range objects {
		names = append(names, object.GetName())
	}

	return names
}
