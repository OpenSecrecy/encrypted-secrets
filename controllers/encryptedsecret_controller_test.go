/*
Copyright 2023.

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

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"

	secretsv1alpha1 "github.com/opensecrecy/encrypted-secrets/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("EncryptedSecrets", func() {

	Context("Verify controller", func() {
		ctx := context.Background()
		namespacedName := types.NamespacedName{
			Namespace: "default",
		}
		It("Reconcile non-existent object", func() {
			namespacedName.Name = "test-encrypted-secret-non-existent"
			res, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
			// expect no error and no requeue since the object doesn't exist
			Expect(err).To(BeNil())
			Expect(res.Requeue).To(BeFalse())
		})
		It("Check for failure since decryption key doesn't exist", func() {

			namespacedName.Name = "test-encrypted-secret"
			instance := &secretsv1alpha1.EncryptedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-encrypted-secret",
					Namespace: "default",
					Annotations: map[string]string{
						"secrets.opensecrecy.org/provider": "k8s",
					},
				},
				Data: map[string]string{
					"secret1": "4GeYrfHDZrGN+QsZO76LnrMTc1zb6sbwIpSRR+SuSbDY+yNjH7K8",
					"secret2": "uuGfVjV9k9al/N92VX9zqk8UN3HvNl77XBgCGSsJqeE=",
				},
			}
			// Create the EncryptedSecret object and expect the Reconcile
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			// reconcile
			res, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})

			Expect(err).To(BeNil())
			Expect(res.Requeue).To(BeFalse())
			// get the instance again to check if the status is updated
			_ = k8sClient.Get(ctx, namespacedName, instance)
			Expect(instance.Status.Status).To(Equal(secretsv1alpha1.EncryptedSecretStatusError))
			Expect(instance.Status.Message).To(ContainSubstring("failed to decrypt"))

			// check for failure since the secret is not created
			secret := &corev1.Secret{}
			err = k8sClient.Get(ctx, namespacedName, secret)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("not found"))

		})
		It("Make sure secret with valid values is created", func() {
			namespacedName.Name = "test-encrypted-secret-success"
			decryptionKey := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cryptctl-key",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"tls.crt": []byte("justRandomEncryptionKey"),
				},
			}
			err := k8sClient.Create(ctx, decryptionKey)
			Expect(err).To(BeNil())

			instance := &secretsv1alpha1.EncryptedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-encrypted-secret-success",
					Namespace: "default",
					Annotations: map[string]string{
						"secrets.opensecrecy.org/provider": "k8s",
					},
				},
				Data: map[string]string{
					"secret": "VdnNsF55TFX9kRiorzy0XPJQRK0FlICFntVqgEMeGOqq+IZfpHmr",
				},
			}
			// Create the EncryptedSecret object and expect the Reconcile
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			// reconcile
			res, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
			Expect(err).To(BeNil())
			Expect(res.Requeue).To(BeFalse())

			// get the instance again to check if the status is updated
			_ = k8sClient.Get(ctx, namespacedName, instance)
			Expect(instance.Status.Status).To(Equal(secretsv1alpha1.EncryptedSecretStatusReady))
			Expect(instance.Status.Message).To(ContainSubstring("ready to be used"))

			// check if the secret is created and has the correct values
			secret := &corev1.Secret{}
			err = k8sClient.Get(ctx, namespacedName, secret)
			Expect(err).To(BeNil())
			Expect(secret.Data["secret"]).To(Equal([]byte("hello-world")))

		})

	})
})
