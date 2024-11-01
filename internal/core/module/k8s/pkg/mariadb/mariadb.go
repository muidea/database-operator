package mariadb

import (
	"fmt"
	"path"

	appv1 "k8s.io/api/apps/v1"
	bachV1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"supos.ai/operator/database/pkg/common"
)

func GetContainerPorts(_ *common.ServiceInfo) (ret []corev1.ContainerPort) {
	ret = []corev1.ContainerPort{
		{
			Name:          "default",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: common.DefaultPostgreSQLPort,
		},
	}

	return
}

func GetEnv(serviceInfo *common.ServiceInfo) (ret []corev1.EnvVar) {
	ret = []corev1.EnvVar{
		{
			Name:  "GROUP",
			Value: "internal",
		},
		{
			Name:  "MALLOC_ARENA_MAX",
			Value: "1",
		},
		{
			Name:  "MANAGE_PORT",
			Value: "8080",
		},
		{
			Name:  "MYSQL_DATABASES",
			Value: "object_internal,cas_service",
		},
		{
			Name:  "MYSQL_ROOT_PASSWORD",
			Value: serviceInfo.Env.Password,
		},
		{
			Name:  "NAME",
			Value: serviceInfo.Name,
		},
		{
			Name:  "PORT",
			Value: "3306",
		},
		{
			Name:  "RUNTIME_METRICS",
			Value: "true",
		},
		{
			Name:  "RUNTIME_METRICS_TTL",
			Value: "1",
		},
		{
			Name:  "RUNTIME_METRICS_URL",
			Value: "lake-pushgateway:9091",
		},
	}
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
		{
			Name:      serviceInfo.Volumes.BackPath.Name,
			MountPath: "/backup",
		},
	}
	return
}

func GetLiveness(_ *common.ServiceInfo) (ret *corev1.Probe) {
	ret = &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"bash", "-ec", "mysqladmin status -uroot -p\"${MYSQL_ROOT_PASSWORD}\" --connect-timeout=2"},
			},
		},
		FailureThreshold:    3,
		InitialDelaySeconds: 30,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		TimeoutSeconds:      5,
	}
	return
}

func GetReadiness(serviceInfo *common.ServiceInfo) (ret *corev1.Probe) {
	return GetLiveness(serviceInfo)
}

func GetStartupProbe(serviceInfo *common.ServiceInfo) (ret *corev1.Probe) {
	return GetLiveness(serviceInfo)
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
			LivenessProbe:   GetLiveness(serviceInfo),
			ReadinessProbe:  GetReadiness(serviceInfo),
			StartupProbe:    GetStartupProbe(serviceInfo),
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
			Port:     common.DefaultPostgreSQLPort,
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
			Resources: corev1.ResourceRequirements{
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

func GetJobInitContainer(serviceInfo *common.ServiceInfo, command []string) (ret []corev1.Container) {
	ret = []corev1.Container{}
	for idx, val := range command {
		ret = append(ret, corev1.Container{
			Name:            fmt.Sprintf("%s%02d", serviceInfo.Name, idx),
			Image:           serviceInfo.Image,
			Command:         []string{"/usr/bin/sh", "-c", val},
			SecurityContext: &corev1.SecurityContext{Privileged: func() *bool { b := true; return &b }()},
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports:           GetContainerPorts(serviceInfo),
			Env:             GetEnv(serviceInfo),
			Resources:       GetResources(serviceInfo),
			VolumeMounts:    GetVolumeMounts(serviceInfo),
		})
	}

	return
}

func GetJobContainer(serviceInfo *common.ServiceInfo, _ []string) (ret []corev1.Container) {
	ret = []corev1.Container{
		{
			Name:            serviceInfo.Name,
			Image:           serviceInfo.Image,
			Command:         []string{"/usr/bin/sh", "-c", "echo 'execute finish!'"},
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports:           GetContainerPorts(serviceInfo),
			Env:             GetEnv(serviceInfo),
			Resources:       GetResources(serviceInfo),
			VolumeMounts:    GetVolumeMounts(serviceInfo),
		},
	}

	return
}

func GetJobVolumes(serviceInfo *common.ServiceInfo) (ret []corev1.Volume) {
	ret = []corev1.Volume{
		{
			Name: serviceInfo.Volumes.DataPath.Name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: serviceInfo.Volumes.DataPath.Value,
				},
			},
		},
		{
			Name: serviceInfo.Volumes.BackPath.Name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: path.Join(serviceInfo.Volumes.BackPath.Value, "data"),
				},
			},
		},
	}
	return
}

func GetJobPodTemplate(serviceInfo *common.ServiceInfo, command []string) (ret corev1.PodTemplateSpec) {
	ret = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: serviceInfo.Labels,
		},
		Spec: corev1.PodSpec{
			InitContainers: GetJobInitContainer(serviceInfo, command),
			Containers:     GetJobContainer(serviceInfo, command),
			Volumes:        GetJobVolumes(serviceInfo),
			RestartPolicy:  corev1.RestartPolicyNever,
		},
	}
	return
}

func GetJob(serviceInfo *common.ServiceInfo, command []string) (ret *bachV1.Job) {
	ret = &bachV1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceInfo.Name,
			Namespace: serviceInfo.Namespace,
			Labels:    serviceInfo.Labels,
		},
		Spec: bachV1.JobSpec{
			Template: GetJobPodTemplate(serviceInfo, command),
		},
	}

	return
}
