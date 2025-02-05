package helper

import (
	"context"
	"fmt"

	updatev1alpha1 "github.com/wellcom-rocks/update-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func CreateImageVersionForDeployment(ctx context.Context, client client.Client, deploy interface{}) error {
	logger := log.FromContext(ctx)

	template, err := GetSpecTemplateFromObject(deploy)
	if err != nil {
		logger.Error(err, "unable to get spec template from object")
		return err
	}

	deployName := GetNameFromObject(deploy)
	if deployName == "" {
		err := fmt.Errorf("cannot get name")
		return err
	}

	deployNamespace := GetNamespaceFromObject(deploy)
	if deployNamespace == "" {
		err := fmt.Errorf("cannot get namespace")
		return err
	}

	containers := template.Spec.Containers
	for _, container := range containers {
		logger.Info("Found container")
		imageVersion := &updatev1alpha1.ImageVersion{}
		imageVersion.ObjectMeta.Name = deployName + "-" + container.Name
		imageVersion.Name = deployName
		imageVersion.Namespace = deployNamespace
		// imageVersion.DeploymentType = template.
		imageVersion.ContainerName = container.Name
		imageVersion.InstalledVersion = container.Image

		foundImageVersion := &updatev1alpha1.ImageVersion{}
		findImageVersion := types.NamespacedName{
			Name:      deployName + "-" + container.Name,
			Namespace: deployNamespace,
		}

		err := client.Get(ctx, findImageVersion, foundImageVersion)
		if err != nil && errors.IsNotFound(err) {
			err := client.Create(ctx, imageVersion)
			if err != nil {
				logger.Error(err, "unable to create ImageVersion")
				return err
			}
			logger.Info("Created ImageVersion")
		} else {
			imageVersion.SetResourceVersion(foundImageVersion.GetResourceVersion())
			err = client.Update(ctx, imageVersion)
			if err != nil {
				logger.Error(err, "unable to update ImageVersion")
				return err
			}
			logger.Info("Updated ImageVersion")
		}
	}

	return nil
}

func GetSpecTemplateFromObject(deploy interface{}) (v1.PodTemplateSpec, error) {
	switch deployObj := deploy.(type) {
	case *appsv1.Deployment:
		template := deployObj.Spec.Template
		return template, nil
	case *appsv1.DaemonSet:
		template := deployObj.Spec.Template
		return template, nil
	default:
		return v1.PodTemplateSpec{}, fmt.Errorf("unknown type")
	}
}

func GetNameFromObject(deploy interface{}) string {
	switch deployObj := deploy.(type) {
	case *appsv1.Deployment:
		return deployObj.Name
	case *appsv1.DaemonSet:
		return deployObj.Name
	default:
		return ""
	}
}

func GetNamespaceFromObject(deploy interface{}) string {
	switch deployObj := deploy.(type) {
	case *appsv1.Deployment:
		return deployObj.Namespace
	case *appsv1.DaemonSet:
		return deployObj.Namespace
	default:
		return ""
	}
}
