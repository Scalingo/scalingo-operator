package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
	"github.com/Scalingo/scalingo-operator/internal/usecases/database/databasemock"
)

func TestDeleteOKSNetPeerings(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctrl := gomock.NewController(t)

	clientStub := &netPeeringResourceClient{
		items: []*unstructured.Unstructured{
			newNetPeering("default", "net-peering-a", "active", "net-id", "owner-id"),
			newNetPeering("default", "net-peering-b", "active", "net-id", "owner-id"),
			newNetPeering("default", "net-peering-inactive", "pending", "net-id", "owner-id"),
			newNetPeering("default", "net-peering-other-network", "active", "other-net-id", "owner-id"),
			newNetPeering("other", "net-peering-other-namespace", "active", "net-id", "owner-id"),
		},
	}
	databaseManager := databasemock.NewMockManager(ctrl)

	reconciler := &PostgreSQLReconciler{
		Client: clientStub,
	}

	postgresql := apiv1.PostgreSQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-resource",
			Namespace: "default",
		},
		Status: apiv1.PostgreSQLStatus{
			ScalingoDatabaseID: "db-123",
		},
	}

	databaseManager.EXPECT().GetDatabaseNetworkConfiguration(ctx, "db-123").Return(domain.DatabaseNetworkConfiguration{
		OutscaleNetID:     "net-id",
		OutscaleAccountID: "owner-id",
	}, nil)

	err := reconciler.deleteNetPeerings(ctx, databaseManager, postgresql)
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

func (c *netPeeringResourceClient) List(_ context.Context, list client.ObjectList, opts ...client.ListOption) error {
	listOptions := &client.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(listOptions)
	}

	typedList := list.(*unstructured.UnstructuredList)
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

func newNetPeering(namespace, name, state, accepterNetID, accepterOwnerID string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(netPeeringGVK)
	object.SetNamespace(namespace)
	object.SetName(name)
	object.Object["status"] = map[string]any{
		"netPeeringState": state,
		"accepterNetId":   accepterNetID,
		"accepterOwnerId": accepterOwnerID,
	}

	return object
}

func resourceNames(objects []*unstructured.Unstructured) []string {
	names := make([]string, 0, len(objects))
	for _, object := range objects {
		names = append(names, object.GetName())
	}

	return names
}
