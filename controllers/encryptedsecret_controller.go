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

	"github.com/go-logr/logr"
	"github.com/shubhindia/crypt-core/providers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hackedsecretsv1alpha1 "github.com/shubhindia/cryptctl/apis/secrets/v1alpha1"
	secretsv1alpha1 "github.com/shubhindia/encrypted-secrets/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// EncryptedSecretReconciler reconciles a EncryptedSecret object
type EncryptedSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	log    logr.Logger
}

//+kubebuilder:rbac:groups=secrets.shubhindia.xyz,resources=encryptedsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=secrets.shubhindia.xyz,resources=encryptedsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=secrets.shubhindia.xyz,resources=encryptedsecrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EncryptedSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *EncryptedSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.log = log.FromContext(ctx).WithValues("EncryptedSecret", req.NamespacedName)
	r.log.Info("Started encryptedsecret reconciliation")

	instance := &secretsv1alpha1.EncryptedSecret{}

	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		r.log.Info("Unable to fetch encryptedsecret object")
		return ctrl.Result{}, client.IgnoreNotFound(err)

	}

	// ToDo: below piece of code is a hack to use the same apis.
	// even though they are same, they are declared at two different places so can't really only one in all the places i.e. cryptctl and encrypted-secrets and crypt-core
	// Will cleanup this mess once I have a solid foundation to work upon. For now this will remain a tech debt

	hackedInstance := hackedsecretsv1alpha1.EncryptedSecret{
		TypeMeta:   instance.TypeMeta,
		ObjectMeta: instance.ObjectMeta,
		Data:       instance.Data,
	}
	decryptedData := make(map[string][]byte)

	decryptedObj, err := providers.DecodeAndDecrypt(&hackedInstance)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to decrypt value for %s", err.Error())
	}

	// create a secret to hold the decrypted secrets
	secretInstance := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	for key, value := range decryptedObj.Data {
		decryptedData[key] = []byte(value)
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, &secretInstance, func() error {
		// set Labels and Annotations
		secretInstance.Labels = instance.Labels
		secretInstance.Annotations = instance.Annotations

		// Add the data
		secretInstance.Data = decryptedData

		// set ownerReference
		err := controllerutil.SetOwnerReference(instance, &secretInstance, r.Scheme)
		if err != nil {
			return fmt.Errorf("error setting owner reference %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting secret %s", err.Error())
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EncryptedSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsv1alpha1.EncryptedSecret{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
