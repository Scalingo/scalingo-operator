/*
Copyright 2026.

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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/controller/adapters"
	"github.com/Scalingo/scalingo-operator/internal/controller/helpers"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// ValkeyReconciler reconciles a Valkey object.
type ValkeyReconciler struct {
	client.Client

	Scheme *runtime.Scheme
}

type valkeyResource struct {
	resource *apiv1.Valkey
}

func (r *valkeyResource) object() client.Object {
	return r.resource
}

func (r *valkeyResource) meta() *metav1.ObjectMeta {
	return &r.resource.ObjectMeta
}

func (r *valkeyResource) toDatabase(ctx context.Context) (domain.Database, error) {
	return adapters.ValkeyToDatabase(ctx, *r.resource)
}

func (r *valkeyResource) authSecret() apiv1.AuthSecretSpec {
	return r.resource.Spec.AuthSecret
}

func (r *valkeyResource) connInfoSecretTarget() apiv1.SecretTargetSpec {
	return r.resource.Spec.ConnInfoSecretTarget
}

func (r *valkeyResource) networking() apiv1.NetworkingSpec {
	return r.resource.Spec.Networking
}

func (r *valkeyResource) region() string {
	return r.resource.Spec.Region
}

func (r *valkeyResource) databaseID() string {
	return r.resource.Status.ScalingoDatabaseID
}

func (r *valkeyResource) setDatabaseID(id string) {
	r.resource.Status.ScalingoDatabaseID = id
}

func (r *valkeyResource) conditions() *[]metav1.Condition {
	return &r.resource.Status.Conditions
}

func newValkeyDedicatedDatabaseReconciler(k8sClient client.Client, scheme *runtime.Scheme) *dedicatedDatabaseReconciler {
	return &dedicatedDatabaseReconciler{
		Client: k8sClient,
		Scheme: scheme,
		config: dedicatedDatabaseConfig{
			newResource: func() dedicatedDatabaseResource {
				return &valkeyResource{resource: &apiv1.Valkey{}}
			},
			finalizerName: helpers.ValkeyFinalizerName,
			databaseType:  domain.DatabaseTypeValkey,
		},
	}
}

func (r *ValkeyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return newValkeyDedicatedDatabaseReconciler(r.Client, r.Scheme).reconcile(ctx, req)
}

// +kubebuilder:rbac:groups=databases.scalingo.com,resources=valkeys,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.scalingo.com,resources=valkeys/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.scalingo.com,resources=valkeys/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oks.dev,resources=netpeeringrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oks.dev,resources=netpeerings,verbs=get;list;delete

// SetupWithManager sets up the controller.
func (r *ValkeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Valkey{}).
		Named("valkey").
		Complete(r)
}
