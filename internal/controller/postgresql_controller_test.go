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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	databasesv1alpha1 "github.com/Scalingo/scalingo-operator/api/v1alpha1"
)

var _ = Describe("PostgreSQL Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		const namespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace,
		}
		postgresql := &databasesv1alpha1.PostgreSQL{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind PostgreSQL")
			err := k8sClient.Get(ctx, typeNamespacedName, postgresql)
			if err != nil && errors.IsNotFound(err) {
				resource := &databasesv1alpha1.PostgreSQL{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace,
					},
					Spec: databasesv1alpha1.PostgreSQLSpec{
						AuthSecret: databasesv1alpha1.AuthSecretSpec{
							Name: "scalingo-auth-secret",
							Key:  "api_token",
						},
						ConnInfoSecretTarget: databasesv1alpha1.SecretTargetSpec{
							Name: "postgresql-conn-info",
						},
						Name:   "my-postgresql-db",
						Plan:   "postgresql-ng-enterprise-4096",
						Region: "osc-st-fr1",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("creating Scalingo auth secret")
			authSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "scalingo-auth-secret",
					Namespace: namespace,
				},
				Type: corev1.SecretTypeOpaque,
				StringData: map[string]string{
					"api_token": "s3cr3t",
				},
			}
			Expect(k8sClient.Create(ctx, authSecret)).To(Succeed())
		})

		AfterEach(func() {
			// Cleanup logic after each test.
			resource := &databasesv1alpha1.PostgreSQL{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance PostgreSQL")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		// By now, we only test the first API call to go-scalingo client using an invalid token.
		// This ensures the spec fields validity, and the secret is read properly.
		It("fails due to invalid token", func() {
			By("Reconciling the created resource")
			controllerReconciler := &PostgreSQLReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})

			errUnauthorizedFullMessage := "create database manager: new scalingo client: fail to get region informations: fail to list regions: fail to call GET /regions: fail to fill client with default values: fail to get the access token for this request: fail to get access token: fail to make request POST /v1/tokens/exchange: unauthorized - you are not authorized to do this operation"
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(errUnauthorizedFullMessage))
		})
	})
})
