package networking

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Scalingo/scalingo-operator/internal/controller/helpers"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

const (
	netPeeringStatusField          = "status"
	netPeeringStatusStateField     = "netPeeringState"
	netPeeringAccepterNetIDField   = "accepterNetId"
	netPeeringAccepterOwnerIDField = "accepterOwnerId"
	netPeeringStatusStateActive    = "active"
)

type NetPeering struct {
	object *unstructured.Unstructured
}

type NetPeeringList struct {
	objects *unstructured.UnstructuredList
}

func NewNetPeeringList() NetPeeringList {
	objects := &unstructured.UnstructuredList{}
	objects.SetGroupVersionKind(helpers.OutscaleNetPeeringListGVK)

	return NetPeeringList{objects: objects}
}

func NetPeeringFromUnstructured(object unstructured.Unstructured) NetPeering {
	return NetPeering{object: &object}
}

func (l NetPeeringList) ObjectList() client.ObjectList {
	return l.objects
}

func (l NetPeeringList) ItemsMatching(databaseNetworkConfig domain.DatabaseNetworkConfiguration) []NetPeering {
	netPeerings := make([]NetPeering, 0, len(l.objects.Items))
	for _, object := range l.objects.Items {
		netPeering := NetPeeringFromUnstructured(object)
		if netPeering.IsActiveForDatabase(databaseNetworkConfig) {
			netPeerings = append(netPeerings, netPeering)
		}
	}

	return netPeerings
}

func (p NetPeering) Object() client.Object {
	return p.object
}

func (p NetPeering) Name() string {
	return p.object.GetName()
}

func (p NetPeering) IsActiveForDatabase(databaseNetworkConfig domain.DatabaseNetworkConfiguration) bool {
	state, _, _ := unstructured.NestedString(p.object.Object, netPeeringStatusField, netPeeringStatusStateField)
	accepterNetID, _, _ := unstructured.NestedString(p.object.Object, netPeeringStatusField, netPeeringAccepterNetIDField)
	accepterOwnerID, _, _ := unstructured.NestedString(p.object.Object, netPeeringStatusField, netPeeringAccepterOwnerIDField)

	return state == netPeeringStatusStateActive && accepterNetID == databaseNetworkConfig.OutscaleNetID && accepterOwnerID == databaseNetworkConfig.OutscaleAccountID
}
