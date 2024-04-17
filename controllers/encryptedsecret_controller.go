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
	"github.com/opensecrecy/encrypted-secrets/pkg/providers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	secretsv1alpha1 "github.com/opensecrecy/encrypted-secrets/api/v1alpha1"
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

//+kubebuilder:rbac:groups=secrets.opensecrecy.org,resources=encryptedsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=secrets.opensecrecy.org,resources=encryptedsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=secrets.opensecrecy.org,resources=encryptedsecrets/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *EncryptedSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.log = log.FromContext(ctx).WithValues("EncryptedSecret", req.NamespacedName)
	r.log.Info("Started encryptedsecret reconciliation")

	instance := &secretsv1alpha1.EncryptedSecret{}

	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		r.log.Info("Unable to fetch encryptedsecret object")
		return ctrl.Result{}, client.IgnoreNotFound(err)

	}

	// check if injectAnnotation exists
	if addInitcontainer, ok := instance.Annotations["secrets.opensecrecy.org/inject-encrypted-secrets"]; ok && addInitcontainer == "true" {
		r.log.Info(fmt.Sprintf("Skipping reconciliation for %s in %s since the secret is supposed to be injected", instance.Name, instance.Namespace))
		return ctrl.Result{Requeue: false}, nil
	}

	fmt.Printf("I should not run\n")

	decryptedObj, err := providers.DecodeAndDecrypt(instance)
	if err != nil {
		r.log.Error(err, "Failed to decrypt")
		instance.Status.Status = secretsv1alpha1.EncryptedSecretStatusError
		instance.Status.Message = fmt.Sprintf("failed to decrypt value for %s", err.Error())
		return r.ensureStatus(ctx, instance, ctrl.Result{})
	}

	// create a secret to hold the decrypted secrets
	secretInstance := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	// map to hold decryptedData in map[string][]byte format
	// ToDo: figure out optimal way to do this. There is absolutely no need to increase space complexity here
	decryptedData := make(map[string][]byte)
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
			instance.Status.Status = secretsv1alpha1.EncryptedSecretStatusError
			instance.Status.Message = fmt.Sprintf("error setting owner reference %s", err.Error())
			return r.Status().Update(ctx, instance)
		}
		return nil
	})
	if err != nil {
		instance.Status.Status = secretsv1alpha1.EncryptedSecretStatusError
		instance.Status.Message = fmt.Sprintf("error getting secret %s", err.Error())
		return r.ensureStatus(ctx, instance, ctrl.Result{})
	}

	instance.Status.Status = secretsv1alpha1.EncryptedSecretStatusReady
	instance.Status.Message = fmt.Sprintf("encrypted secrets %s is ready to be used", instance.Name)
	return r.ensureStatus(ctx, instance, ctrl.Result{})
}

// SetupWithManager sets up the controller with the Manager.
func (r *EncryptedSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsv1alpha1.EncryptedSecret{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

// ensureStatus makes sure that proper status is applied to the EncryptedSecret instance
func (r *EncryptedSecretReconciler) ensureStatus(ctx context.Context, instance *secretsv1alpha1.EncryptedSecret, result ctrl.Result) (ctrl.Result, error) {

	err := r.Status().Update(ctx, instance)
	if err != nil {
		r.log.Error(err, "Failed to update status")
		return ctrl.Result{Requeue: true}, nil
	}

	return result, nil
}
