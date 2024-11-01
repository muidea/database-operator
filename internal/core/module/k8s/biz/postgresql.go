package biz

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"supos.ai/operator/database/internal/core/module/k8s/pkg/postgresql"
	"supos.ai/operator/database/pkg/common"
)

func (s *K8s) createPostgreSQL(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	// 1、Create pvc
	_, pvcErr := s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Create(context.TODO(),
		postgresql.GetPersistentVolumeClaims(serviceInfo),
		metav1.CreateOptions{})
	if pvcErr != nil {
		err = cd.NewError(cd.UnExpected, pvcErr.Error())
		log.Errorf("createPostgreSQL %v pvc failed, s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Create error:%s",
			serviceInfo, pvcErr.Error())
		return
	}

	// 2、Create Deployment
	_, deploymentErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Create(context.TODO(),
		postgresql.GetDeployment(serviceInfo),
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
		postgresql.GetService(serviceInfo),
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
