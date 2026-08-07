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
	"fmt"

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
	object *apiv1.PostgreSQL
}

func (r *postgresqlResource) Object() client.Object {
	return r.object
}

func (r *postgresqlResource) Meta() *metav1.ObjectMeta {
	return &r.object.ObjectMeta
}

func (r *postgresqlResource) AuthSecret() apiv1.AuthSecretSpec {
	return r.object.Spec.AuthSecret
}

func (r *postgresqlResource) ConnInfoSecretTarget() apiv1.SecretTargetSpec {
	return r.object.Spec.ConnInfoSecretTarget
}

func (r *postgresqlResource) Networking() apiv1.NetworkingSpec {
	return r.object.Spec.Networking
}

func (r *postgresqlResource) Region() string {
	return r.object.Spec.Region
}

func (r *postgresqlResource) DatabaseID() string {
	return r.object.Status.ScalingoDatabaseID
}

func (r *postgresqlResource) SetDatabaseID(id string) {
	r.object.Status.ScalingoDatabaseID = id
}

func (r *postgresqlResource) Conditions() *[]metav1.Condition {
	return &r.object.Status.Conditions
}

func newPostgreSQLDedicatedDatabaseReconciler(k8sClient client.Client, scheme *runtime.Scheme) *dedicatedDatabaseReconciler {
	return &dedicatedDatabaseReconciler{
		Client: k8sClient,
		Scheme: scheme,
		config: dedicatedDatabaseConfig{
			newResource: func() dedicatedDatabaseResource {
				return &postgresqlResource{object: &apiv1.PostgreSQL{}}
			},
			finalizerName: helpers.PostgreSQLFinalizerName,
			databaseType:  domain.DatabaseTypePostgreSQL,
			toDatabase: func(ctx context.Context, resource dedicatedDatabaseResource) (domain.Database, error) {
				postgresqlResource, ok := resource.(*postgresqlResource)
				if !ok {
					return domain.Database{}, fmt.Errorf("expected PostgreSQL resource, got %T", resource)
				}
				return adapters.PostgreSQLToDatabase(ctx, *postgresqlResource.object)
			},
		},
	}
}

func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return newPostgreSQLDedicatedDatabaseReconciler(r.Client, r.Scheme).Reconcile(ctx, req)
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
