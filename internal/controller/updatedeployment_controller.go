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

	updatev1alpha1 "github.com/wellcom-rocks/update-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// UpdateDeploymentReconciler reconciles a Update object
type UpdateDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Update object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *UpdateDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		logger.Error(err, "unable to fetch Deployment")
		return ctrl.Result{}, err
	}

	logger.Info("Getting Deployment")

	containers := deployment.Spec.Template.Spec.Containers
	for _, container := range containers {
		logger.Info("Found container")

		imageVersion := &updatev1alpha1.ImageVersion{}
		imageVersion.ObjectMeta.Name = deployment.Name + "-" + container.Name
		imageVersion.Name = deployment.Name
		imageVersion.Namespace = req.Namespace
		imageVersion.DeploymentType = "deployment"
		imageVersion.ContainerName = container.Name
		imageVersion.InstalledVersion = container.Image

		foundImageVersion := &updatev1alpha1.ImageVersion{}
		findImageVersion := types.NamespacedName{
			Name:      deployment.Name + "-" + container.Name,
			Namespace: req.Namespace,
		}

		err := r.Client.Get(ctx, findImageVersion, foundImageVersion)
		if err != nil && errors.IsNotFound(err) {
			err := r.Create(ctx, imageVersion)
			if err != nil {
				logger.Error(err, "unable to create ImageVersion")
				return ctrl.Result{}, err
			}
		}

		imageVersion.SetResourceVersion(foundImageVersion.GetResourceVersion())
		err = r.Update(ctx, imageVersion)
		if err != nil {
			logger.Error(err, "unable to update ImageVersion")
			return ctrl.Result{}, err
		}

		logger.Info("Created ImageVersion", "ImageVersion", *imageVersion)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UpdateDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
