package biz

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"supos.ai/operator/database/internal/core/module/k8s/pkg/database"
	"supos.ai/operator/database/pkg/common"
)

func (s *K8s) createDatabase(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	_, curErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Get(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if curErr == nil || !errors.IsNotFound(curErr) {
		return
	}

	// 1、Create pvc
	_, pvcErr := s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Create(context.TODO(),
		database.GetPersistentVolumeClaims(serviceInfo),
		metav1.CreateOptions{})
	if pvcErr != nil {
		err = cd.NewError(cd.UnExpected, pvcErr.Error())
		log.Errorf("createDatabase %v pvc failed, s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Create error:%s",
			serviceInfo, pvcErr.Error())
		return
	}

	// 2、Create Deployment
	_, deploymentErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Create(context.TODO(),
		database.GetDeployment(serviceInfo),
		metav1.CreateOptions{})
	if deploymentErr != nil {
		err = cd.NewError(cd.UnExpected, deploymentErr.Error())
		log.Errorf("createDatabase %v deployment failed, s.clientSet.AppsV1().Deployments(s.getNamespace()).Create error:%s",
			serviceInfo, deploymentErr.Error())

		s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
		return
	}

	// 3、Create Service
	_, serviceErr := s.clientSet.CoreV1().Services(s.getNamespace()).Create(context.TODO(),
		database.GetService(serviceInfo),
		metav1.CreateOptions{})
	if serviceErr != nil {
		err = cd.NewError(cd.UnExpected, serviceErr.Error())
		log.Errorf("createDatabase %v service failed, s.clientSet.CoreV1().Services(s.getNamespace()).Create error:%s",
			serviceInfo, serviceErr.Error())

		s.clientSet.AppsV1().Deployments(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
		s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
		return
	}

	return
}

func (s *K8s) destroyDatabase(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	_ = s.clientSet.CoreV1().Services(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
	_ = s.clientSet.AppsV1().Deployments(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
	_ = s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})

	return
}

func (s *K8s) startDatabase(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	scalePtr, scaleErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).GetScale(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if scaleErr != nil {
		err = cd.NewError(cd.UnExpected, scaleErr.Error())
		log.Errorf("startDatabase %v failed, get service scale error:%s", serviceInfo, scaleErr.Error())
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
		log.Errorf("startDatabase %v failed, set service scale error:%s", serviceInfo, scaleErr.Error())
		return
	}

	return
}

func (s *K8s) stopDatabase(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	scalePtr, scaleErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).GetScale(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if scaleErr != nil {
		err = cd.NewError(cd.UnExpected, scaleErr.Error())
		log.Errorf("stopDatabase %v failed, get service scale error:%s", serviceInfo, scaleErr.Error())
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
		log.Errorf("stopDatabase %v failed, set service scale error:%s", serviceInfo, scaleErr.Error())
		return
	}

	return
}
