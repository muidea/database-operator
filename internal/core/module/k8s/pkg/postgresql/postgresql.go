package postgresql

import (
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"supos.ai/operator/database/pkg/common"
)

func GetContainerPorts(_ *common.ServiceInfo) (ret []corev1.ContainerPort) {
	ret = []corev1.ContainerPort{
		{
			Name:     "default",
			Protocol: corev1.ProtocolTCP,
		},
	}

	return
}

func GetEnv(serviceInfo *common.ServiceInfo) (ret []corev1.EnvVar) {
	ret = []corev1.EnvVar{}
	return
}

func GetResources(_ *common.ServiceInfo) (ret corev1.ResourceRequirements) {
	resourceQuantity := func(quantity string) resourcev1.Quantity {
		r, _ := resourcev1.ParseQuantity(quantity)
		return r
	}

	ret = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resourceQuantity(common.PostgreSQLDefaultSpec.CPU),
			corev1.ResourceMemory: resourceQuantity(common.PostgreSQLDefaultSpec.Memory),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resourceQuantity("100m"),
			corev1.ResourceMemory: resourceQuantity("64Mi"),
		},
	}

	return
}

func GetVolumeMounts(serviceInfo *common.ServiceInfo) (ret []corev1.VolumeMount) {
	ret = []corev1.VolumeMount{
		{
			Name:      serviceInfo.Volumes.DataPath.Name,
			MountPath: "/var/lib/mysql",
		},
	}
	return
}

func GetContainer(serviceInfo *common.ServiceInfo) (ret []corev1.Container) {
	ret = []corev1.Container{
		{
			Name:            serviceInfo.Name,
			Image:           serviceInfo.Image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports:           GetContainerPorts(serviceInfo),
			Env:             GetEnv(serviceInfo),
			Resources:       GetResources(serviceInfo),
			VolumeMounts:    GetVolumeMounts(serviceInfo),
		},
	}
	return
}

func GetVolumes(serviceInfo *common.ServiceInfo) (ret []corev1.Volume) {
	ret = []corev1.Volume{
		{
			Name: serviceInfo.Volumes.DataPath.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: serviceInfo.Volumes.DataPath.Value,
				},
			},
		},
		{
			Name: serviceInfo.Volumes.BackPath.Name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: serviceInfo.Volumes.BackPath.Value,
				},
			},
		},
	}
	return
}

func GetPodTemplate(serviceInfo *common.ServiceInfo) (ret corev1.PodTemplateSpec) {
	ret = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: serviceInfo.Labels,
		},
		Spec: corev1.PodSpec{
			Containers: GetContainer(serviceInfo),
			Volumes:    GetVolumes(serviceInfo),
		},
	}
	return
}

func GetDeployment(serviceInfo *common.ServiceInfo) (ret *appv1.Deployment) {
	ret = &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceInfo.Name,
			Namespace: serviceInfo.Namespace,
			Labels:    serviceInfo.Labels,
		},
		Spec: appv1.DeploymentSpec{
			Replicas: &serviceInfo.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: serviceInfo.Labels,
			},
			Template: GetPodTemplate(serviceInfo),
			Strategy: appv1.DeploymentStrategy{
				Type: appv1.RecreateDeploymentStrategyType,
			},
		},
	}

	return
}

func GetServicePorts(serviceInfo *common.ServiceInfo) (ret []corev1.ServicePort) {
	ret = []corev1.ServicePort{
		{
			Name:     "default",
			Protocol: corev1.ProtocolTCP,
			TargetPort: intstr.IntOrString{
				Type:   0,
				IntVal: serviceInfo.Svc.Port,
			},
		},
	}

	return
}

func GetService(serviceInfo *common.ServiceInfo) (ret *corev1.Service) {
	ret = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceInfo.Name,
			Namespace: serviceInfo.Namespace,
			Labels:    serviceInfo.Labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: GetServicePorts(serviceInfo),
			Type:  corev1.ServiceTypeClusterIP,
		},
	}
	return
}

func GetPersistentVolumeClaims(serviceInfo *common.ServiceInfo) (ret *corev1.PersistentVolumeClaim) {
	resourceQuantity := func(quantity string) resourcev1.Quantity {
		r, _ := resourcev1.ParseQuantity(quantity)
		return r
	}
	storageClassName := func(className string) *string {
		ret := className
		return &ret
	}
	volumeModeFileSystem := func() *corev1.PersistentVolumeMode {
		ret := corev1.PersistentVolumeFilesystem
		return &ret
	}

	ret = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceInfo.Name,
			Namespace: serviceInfo.Namespace,
			Labels:    serviceInfo.Labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resourceQuantity("10Gi"),
				},
			},
			StorageClassName: storageClassName(serviceInfo.Volumes.DataPath.Type),
			VolumeName:       serviceInfo.Name,
			VolumeMode:       volumeModeFileSystem(),
		},
	}

	return
}
