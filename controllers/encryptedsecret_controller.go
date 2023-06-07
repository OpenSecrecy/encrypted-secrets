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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	secretsv1alpha1 "github.com/shubhindia/encrypted-secrets/api/v1alpha1"
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

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EncryptedSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsv1alpha1.EncryptedSecret{}).
		Complete(r)
}
