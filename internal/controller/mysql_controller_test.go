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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	databasesv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/controller/helpers"
)

var _ = Describe("MySQL Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName      = "test-resource"
			resourceNamespace = "default"
		)

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}
		mysql := &databasesv1.MySQL{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind MySQL")
			err := k8sClient.Get(ctx, typeNamespacedName, mysql)
			if err != nil && errors.IsNotFound(err) {
				resource := &databasesv1.MySQL{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: resourceNamespace,
					},
					Spec: databasesv1.MySQLSpec{
						AuthSecret: databasesv1.AuthSecretSpec{
							Name: "scalingo-auth-secret",
							Key:  "api_token",
						},
						ConnInfoSecretTarget: databasesv1.SecretTargetSpec{
							Name: "mysql-conn-info",
						},
						Networking: databasesv1.NetworkingSpec{
							InternetAccess: databasesv1.InternetAccessSpec{Enabled: true},
						},
						Plan:   "mysql-dr-starter-8192",
						Region: "osc-fr1",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &databasesv1.MySQL{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			controllerutil.RemoveFinalizer(resource, helpers.MySQLFinalizerName)
			err = k8sClient.Update(ctx, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance MySQL")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("initializes the MySQL resource before provisioning", func() {
			controllerReconciler := &MySQLReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			request := reconcile.Request{NamespacedName: typeNamespacedName}

			By("adding the MySQL finalizer")
			result, err := controllerReconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(helpers.RequeueShortDelay))

			resource := &databasesv1.MySQL{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(controllerutil.ContainsFinalizer(resource, helpers.MySQLFinalizerName)).To(BeTrue())

			By("initializing the database status")
			result, err = controllerReconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(helpers.RequeueShortDelay))

			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(helpers.IsDatabaseInitialized(resource.Status.Conditions)).To(BeTrue())

			By("initializing the running annotation")
			result, err = controllerReconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(helpers.RequeueShortDelay))

			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(metav1.HasAnnotation(resource.ObjectMeta, helpers.DatabaseAnnotationIsRunning)).To(BeTrue())

			By("stopping before provisioning when the auth secret is missing")
			_, err = controllerReconciler.Reconcile(ctx, request)
			Expect(err).To(MatchError(ContainSubstring("get auth secret")))
		})
	})
})
