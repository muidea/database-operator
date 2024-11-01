package biz

import (
	"context"
	appv1 "k8s.io/api/apps/v1"
	"strings"
	"supos.ai/operator/database/internal/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"supos.ai/operator/database/internal/core/module/k8s/pkg/mariadb"
	"supos.ai/operator/database/pkg/common"
)

func (s *K8s) checkIsPostgreSQL(deploymentPtr *appv1.Deployment) bool {
	return strings.Index(deploymentPtr.ObjectMeta.GetName(), common.PostgreSQL) != -1
}

func (s *K8s) getDefaultPostgreSQLServiceInfo(serviceName string) (ret *common.ServiceInfo) {
	ret = &common.ServiceInfo{
		Name:      serviceName,
		Namespace: s.getNamespace(),
		Catalog:   common.PostgreSQL,
		Image:     common.DefaultPostgreSQLImage,
		Labels:    common.DefaultLabels,
		Spec:      &common.PostgreSQLDefaultSpec,
		Volumes: &common.Volumes{
			ConfPath: &common.Path{
				Name:  "config",
				Value: common.DefaultPostgreSQLConfigPath,
				Type:  common.InnerPath,
			},
			BackPath: &common.Path{
				Name:  "back-path",
				Value: common.DefaultPostgreSQLBackPath,
				Type:  common.HostPath,
			},
		},
		Env: &common.Env{
			Root:     common.DefaultPostgreSQLRoot,
			Password: common.DefaultPostgreSQLPassword,
		},
		Svc: &common.Svc{
			Host: config.GetLocalHost(),
			Port: common.DefaultPostgreSQLPort,
		},
		Replicas: 1,
	}

	ret.Labels["app"] = serviceName
	ret.Labels["catalog"] = common.PostgreSQL
	return
}

func (s *K8s) createPostgreSQL(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	// 1、Create pvc
	_, pvcErr := s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Create(context.TODO(),
		mariadb.GetPersistentVolumeClaims(serviceInfo),
		metav1.CreateOptions{})
	if pvcErr != nil {
		err = cd.NewError(cd.UnExpected, pvcErr.Error())
		log.Errorf("createPostgreSQL %v pvc failed, s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Create error:%s",
			serviceInfo, pvcErr.Error())
		return
	}

	// 2、Create Deployment
	_, deploymentErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Create(context.TODO(),
		mariadb.GetDeployment(serviceInfo),
		metav1.CreateOptions{})
	if deploymentErr != nil {
		err = cd.NewError(cd.UnExpected, deploymentErr.Error())
		log.Errorf("createPostgreSQL %v deployment failed, s.clientSet.AppsV1().Deployments(s.getNamespace()).Create error:%s",
			serviceInfo, deploymentErr.Error())

		s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
		return
	}

	// 3、Create Service
	_, serviceErr := s.clientSet.CoreV1().Services(s.getNamespace()).Create(context.TODO(),
		mariadb.GetService(serviceInfo),
		metav1.CreateOptions{})
	if serviceErr != nil {
		err = cd.NewError(cd.UnExpected, serviceErr.Error())
		log.Errorf("createPostgreSQL %v service failed, s.clientSet.CoreV1().Services(s.getNamespace()).Create error:%s",
			serviceInfo, serviceErr.Error())

		s.clientSet.AppsV1().Deployments(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
		s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
		return
	}

	return
}

func (s *K8s) destroyPostgreSQL(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	_ = s.clientSet.CoreV1().Services(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
	_ = s.clientSet.AppsV1().Deployments(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
	_ = s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})

	return
}

func (s *K8s) startPostgreSQL(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	scalePtr, scaleErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).GetScale(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if scaleErr != nil {
		err = cd.NewError(cd.UnExpected, scaleErr.Error())
		log.Errorf("startPostgreSQL %v failed, get service scale error:%s", serviceInfo, scaleErr.Error())
		return
	}

	scalePtr.Status.Replicas = serviceInfo.Replicas
	_, scaleErr = s.clientSet.AppsV1().Deployments(s.getNamespace()).UpdateScale(
		context.TODO(),
		serviceInfo.Name,
		scalePtr,
		metav1.UpdateOptions{},
	)
	if scaleErr != nil {
		err = cd.NewError(cd.UnExpected, scaleErr.Error())
		log.Errorf("startPostgreSQL %v failed, set service scale error:%s", serviceInfo, scaleErr.Error())
		return
	}

	return
}

func (s *K8s) stopPostgreSQL(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	scalePtr, scaleErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).GetScale(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if scaleErr != nil {
		err = cd.NewError(cd.UnExpected, scaleErr.Error())
		log.Errorf("stopPostgreSQL %v failed, get service scale error:%s", serviceInfo, scaleErr.Error())
		return
	}

	scalePtr.Status.Replicas = 0
	_, scaleErr = s.clientSet.AppsV1().Deployments(s.getNamespace()).UpdateScale(
		context.TODO(),
		serviceInfo.Name,
		scalePtr,
		metav1.UpdateOptions{},
	)
	if scaleErr != nil {
		err = cd.NewError(cd.UnExpected, scaleErr.Error())
		log.Errorf("stopPostgreSQL %v failed, set service scale error:%s", serviceInfo, scaleErr.Error())
		return
	}

	return
}

func (s *K8s) jobPostgreSQL(serviceInfo *common.ServiceInfo, command []string) (err *cd.Result) {
	job, jobErr := s.clientSet.BatchV1().Jobs(s.getNamespace()).Create(context.TODO(),
		mariadb.GetJob(serviceInfo, command),
		metav1.CreateOptions{})
	if jobErr != nil {
		err = cd.NewError(cd.UnExpected, jobErr.Error())
		log.Errorf("jobPostgreSQL %v service failed, s.clientSet.BatchV1().Jobs(s.getNamespace()).Create error:%s",
			serviceInfo, jobErr.Error())
		return
	}
	errWait := s.waitForJobFinished(serviceInfo, job)
	if errWait != nil {
		log.Errorf("PostgreSQL job %+v failed: %v", serviceInfo, errWait)
		err = cd.NewError(cd.Failed, errWait.Error())
		return
	}
	return
}
