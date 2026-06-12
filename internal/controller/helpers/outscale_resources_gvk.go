package helpers

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	OutscaleNetPeeringRequestGVK = schema.GroupVersionKind{
		Group:   "oks.dev",
		Version: "v1beta",
		Kind:    "NetPeeringRequest",
	}

	OutscaleNetPeeringRequestListGVK = schema.GroupVersionKind{
		Group:   "oks.dev",
		Version: "v1beta",
		Kind:    "NetPeeringRequestList",
	}

	OutscaleNetPeeringGVK = schema.GroupVersionKind{
		Group:   "oks.dev",
		Version: "v1beta",
		Kind:    "NetPeering",
	}

	OutscaleNetPeeringListGVK = schema.GroupVersionKind{
		Group:   "oks.dev",
		Version: "v1beta",
		Kind:    "NetPeeringList",
	}
)
