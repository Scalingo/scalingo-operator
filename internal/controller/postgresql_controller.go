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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/controller/adapters"
	"github.com/Scalingo/scalingo-operator/internal/controller/helpers"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// PostgreSQLReconciler reconciles a PostgreSQL object.
type PostgreSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type postgresqlResource struct {
	resource *apiv1.PostgreSQL
}

func (r *postgresqlResource) object() client.Object {
	return r.resource
}

func (r *postgresqlResource) meta() *metav1.ObjectMeta {
	return &r.resource.ObjectMeta
}

func (r *postgresqlResource) toDatabase(ctx context.Context) (domain.Database, error) {
	return adapters.PostgreSQLToDatabase(ctx, *r.resource)
}

func (r *postgresqlResource) authSecret() apiv1.AuthSecretSpec {
	return r.resource.Spec.AuthSecret
}

func (r *postgresqlResource) connInfoSecretTarget() apiv1.SecretTargetSpec {
	return r.resource.Spec.ConnInfoSecretTarget
}

func (r *postgresqlResource) networking() apiv1.NetworkingSpec {
	return r.resource.Spec.Networking
}

func (r *postgresqlResource) region() string {
	return r.resource.Spec.Region
}

func (r *postgresqlResource) databaseID() string {
	return r.resource.Status.ScalingoDatabaseID
}

func (r *postgresqlResource) setDatabaseID(id string) {
	r.resource.Status.ScalingoDatabaseID = id
}

func (r *postgresqlResource) conditions() *[]metav1.Condition {
	return &r.resource.Status.Conditions
}

func newPostgreSQLDedicatedDatabaseReconciler(k8sClient client.Client, scheme *runtime.Scheme) *dedicatedDatabaseReconciler {
	return &dedicatedDatabaseReconciler{
		Client: k8sClient,
		Scheme: scheme,
		config: dedicatedDatabaseConfig{
			newResource: func() dedicatedDatabaseResource {
				return &postgresqlResource{resource: &apiv1.PostgreSQL{}}
			},
			finalizerName: helpers.PostgreSQLFinalizerName,
			databaseType:  domain.DatabaseTypePostgreSQL,
		},
	}
}

func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return newPostgreSQLDedicatedDatabaseReconciler(r.Client, r.Scheme).reconcile(ctx, req)
}

// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.scalingo.com,resources=postgresqls/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oks.dev,resources=netpeeringrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oks.dev,resources=netpeerings,verbs=get;list;delete

// SetupWithManager sets up the controller.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.PostgreSQL{}).
		Named("postgresql").
		Complete(r)
}
