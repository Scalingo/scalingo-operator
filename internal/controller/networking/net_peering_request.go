package networking

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/Scalingo/scalingo-operator/internal/controller/helpers"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

const (
	netPeeringRequestSuffix = "net-peering-request"

	netPeeringRequestAPIVersionField = "apiVersion"
	netPeeringRequestKindField       = "kind"
	netPeeringRequestMetadataField   = "metadata"
	netPeeringRequestNameField       = "name"
	netPeeringRequestNamespaceField  = "namespace"
	netPeeringRequestLabelsField     = "labels"
	netPeeringRequestSpecField       = "spec"
	netPeeringRequestStatusField     = "status"

	netPeeringRequestManagedByLabel  = "app.kubernetes.io/managed-by"
	netPeeringRequestDatabaseIDLabel = "scalingo.com/database-id"
	netPeeringRequestManagerName     = "scalingo-operator"

	netPeeringRequestNetPeeringIDField    = "netPeeringId"
	netPeeringRequestAccepterNetIDField   = "accepterNetId"
	netPeeringRequestAccepterOwnerIDField = "accepterOwnerId"
)

type NetPeeringRequest struct {
	object *unstructured.Unstructured
}

type NetPeeringRequestList struct {
	objects *unstructured.UnstructuredList
}

func NewNetPeeringRequestList() NetPeeringRequestList {
	objects := &unstructured.UnstructuredList{}
	objects.SetGroupVersionKind(helpers.OutscaleNetPeeringRequestListGVK)

	return NetPeeringRequestList{objects: objects}
}

func NewNetPeeringRequest(resource DatabaseResource, databaseNetworkConfig domain.DatabaseNetworkConfiguration) NetPeeringRequest {
	now := time.Now().UnixMilli()
	resourceName := fmt.Sprintf("%s-%s-%d", resource.Name, netPeeringRequestSuffix, now)
	apiVersion, kind := helpers.OutscaleNetPeeringRequestGVK.ToAPIVersionAndKind()

	object := &unstructured.Unstructured{
		Object: map[string]any{
			netPeeringRequestAPIVersionField: apiVersion,
			netPeeringRequestKindField:       kind,
			netPeeringRequestMetadataField: map[string]any{
				netPeeringRequestNameField:      resourceName,
				netPeeringRequestNamespaceField: resource.Namespace,
				netPeeringRequestLabelsField: map[string]any{
					netPeeringRequestManagedByLabel:  netPeeringRequestManagerName,
					netPeeringRequestDatabaseIDLabel: resource.DatabaseID,
				},
			},
			netPeeringRequestSpecField: map[string]any{
				netPeeringRequestAccepterNetIDField:   databaseNetworkConfig.OutscaleNetID,
				netPeeringRequestAccepterOwnerIDField: databaseNetworkConfig.OutscaleAccountID,
			},
		},
	}

	return NetPeeringRequest{object: object}
}

func NetPeeringRequestFromUnstructured(object unstructured.Unstructured) NetPeeringRequest {
	return NetPeeringRequest{object: &object}
}

func NetPeeringRequestMatchingLabels(databaseID string) client.MatchingLabels {
	return client.MatchingLabels{netPeeringRequestDatabaseIDLabel: databaseID}
}

func (l NetPeeringRequestList) ObjectList() client.ObjectList {
	return l.objects
}

func (l NetPeeringRequestList) Items() []NetPeeringRequest {
	requests := make([]NetPeeringRequest, 0, len(l.objects.Items))
	for _, object := range l.objects.Items {
		requests = append(requests, NetPeeringRequestFromUnstructured(object))
	}

	return requests
}

func (r NetPeeringRequest) Object() client.Object {
	return r.object
}

func (r NetPeeringRequest) SetControllerReference(owner client.Object, scheme *runtime.Scheme) error {
	return controllerutil.SetControllerReference(owner, r.object, scheme)
}

func (r NetPeeringRequest) NetPeeringID() string {
	value, found, _ := unstructured.NestedString(r.object.Object, netPeeringRequestStatusField, netPeeringRequestNetPeeringIDField)
	if found && value != "" {
		return value
	}
	return ""
}
