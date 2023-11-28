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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	secretsv1alpha1 "github.com/opensecrecy/encrypted-secrets/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("EncryptedSecrets", func() {

	Context("Verify controller", func() {
		ctx := context.Background()
		namespacedName := types.NamespacedName{
			Namespace: "default",
		}
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
					"tls.crt": []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMrakNDQWVLZ0F3SUJBZ0lRWFBSM2xHV1VSL2hOUTQyTTBFQWNJekFOQmdrcWhraUc5dzBCQVFzRkFEQVgKTVJVd0V3WURWUVFERXd4amNubHdkR04wYkMxclpYa3dIaGNOTWpNd09ESXhNVEUxT1RVMldoY05Nek13T0RFNApNVEUxT1RVMldqQVhNUlV3RXdZRFZRUURFd3hqY25sd2RHTjBiQzFyWlhrd2dnRWlNQTBHQ1NxR1NJYjNEUUVCCkFRVUFBNElCRHdBd2dnRUtBb0lCQVFEVDNreFZCTHdiTHEvUzNxUzlNVU9id2tiV05CZGZUVkFHekdOUEpyVXcKWmE1c1dhNExUNGtZcHhLWDBIVW1aNmIrckF1c1ZHT0oxL2haNi9yYU1DOHI5VVUwSVVPMit1d3B4QlE5eDVYZwo0aWZFRnFBOTJCNFRRRStXSVkzUFFZdnlsZWVCc2RIQWVHYUo2Tlc2bC93enUxZFh4eHVsYmgzU2R3QUk0TFJoCmwrRjlCNGovSWgwUlBjOWlBc21CY0xLYjVwcDZwMlJZdFVObWwwTWNPd1lhWXIyRWhhZnkxNC84UVRvRW11TzUKV0dleUJidHI0MytzOXloZmdNYmRyaEpuYzhLeCtrQTZISWw1N0RtTS9XR2hKcVNLRmxrN3E0UHhPLzZWeVR1awpOZUtEZ3dJQ0dyRDYzOEpqRTcvZi94YnY2azlyY3paV2dBWWppTUdVRUx1dEFnTUJBQUdqUWpCQU1BNEdBMVVkCkR3RUIvd1FFQXdJQUFUQVBCZ05WSFJNQkFmOEVCVEFEQVFIL01CMEdBMVVkRGdRV0JCUzErODBMMTRKaDVqdnUKT2xSTndhVEtORnBhanpBTkJna3Foa2lHOXcwQkFRc0ZBQU9DQVFFQUlzVWh1TTVKZWp1WjNLdXpYNWVJdkxGUQpXM2NPVHBuWTB6KzhoclVNVEg1U3RycnJXRzJoblZ2ZXMvL0NXb2ZwdmxyUHhMYWsxWm5meGRhNmV4ZW9LckdvCnJRd3pNTytvQ055eFNCTGFWSHJ6cC91YThiTFo4RlZlTHRXcmtSZWJhVXMwWnNZalpydXFmQ3BQb2NSbjM0QzMKK0hKbjVwQTJwSjZaUFZrNnlGWDRlWGNqY3Y0TWdpRWt2NWd6R2pMNEFrK2JQelNRbkpQVnpiMysxS1Q1Q1NRegpJRmRsL3VGQ2taMitRVHJJZnBKYUo2QlVKTkhMWENBZmc0YTZsRERyS3NkZGUzTmlVck83MDJvREk2eTg2dVJECmxqVWlJL0xseEsxTGRDVHlURlRUSUY1ZTA5d2VwaHNsR2p0SUlJZm5HSk43bzZ3NlNSVHI3UnQ3eWN6dGZBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="),
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
			fmt.Printf("%+v", instance)
			Expect(instance.Status.Status).To(Equal(secretsv1alpha1.EncryptedSecretStatusReady))
			Expect(instance.Status.Message).To(ContainSubstring("failed to decrypt"))

		})

	})
})
