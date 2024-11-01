package biz

import (
	"context"
	"encoding/json"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/task"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"supos.ai/operator/database/internal/core/base/biz"
	"supos.ai/operator/database/pkg/common"

	pgv1 "supos.ai/operator/database/pkg/crds/v1"
)

type PostgreSQL struct {
	biz.Base

	client dynamic.Interface
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *PostgreSQL {
	ptr := &PostgreSQL{
		Base: biz.New(common.PostgreSQLModule, eventHub, backgroundRoutine),
	}

	ptr.SubscribeFunc(common.NotifyTimer, ptr.timerCheck)
	return ptr
}

func (s *PostgreSQL) timerCheck(_ event.Event, _ event.Result) {
	s.List("default")
}

func (s *PostgreSQL) Run() {
	s.AsyncTask(func() {

	})
}

func (s *PostgreSQL) getResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: pgv1.Group, Version: pgv1.Version, Resource: pgv1.Postgresql}
}

func (s *PostgreSQL) getK8sClient() (ret dynamic.Interface) {
	if s.client != nil {
		ret = s.client
		return
	}

	ev := event.NewEvent(common.GetK8sConfig, s.ID(), common.K8sModule, nil, nil)
	result := s.SendEvent(ev)
	cfgVal, cfgErr := result.Get()
	if cfgErr != nil {
		log.Errorf("getK8sClient failed, error:%s", cfgErr.Error())
		return
	}

	dynamicClient, dynamicErr := dynamic.NewForConfig(cfgVal.(*rest.Config))
	if dynamicErr != nil {
		log.Errorf("getK8sClient failed, dynamic.NewForConfig error:%s", dynamicErr.Error())
		return
	}

	s.client = dynamicClient
	ret = s.client
	return
}

func (s *PostgreSQL) List(namespace string) {
	res := s.getResource()
	client := s.getK8sClient()
	resList, resErr := client.Resource(res).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if resErr != nil {
		log.Errorf("s.client.Resource(res).Namespace(namespace).List Postgresql failed, namespace:%s, error:%s", namespace, resErr.Error())
		return
	}

	var pgList pgv1.PostgreSQLList
	convertErr := runtime.DefaultUnstructuredConverter.FromUnstructured(resList.UnstructuredContent(), &pgList)
	if convertErr != nil {
		log.Errorf("runtime.DefaultUnstructuredConverter.FromUnstructured failed, error:%s", convertErr.Error())
		return
	}

	log.Infof("%s, count:%v", pgList.Kind, len(pgList.Items))
	if len(pgList.Items) > 0 {
		log.Infof("%s", pgList.Items[0].GetName())
	}
}

func (s *PostgreSQL) Get(namespace, name string) {
	res := s.getResource()
	client := s.getK8sClient()
	resVal, resErr := client.Resource(res).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if resErr != nil {
		log.Errorf("s.client.Resource(res).Namespace(namespace).Get Postgresql failed, namespace:%s, name:%s, error:%s", namespace, name, resErr.Error())
		return
	}

	var pgVal pgv1.PostgreSQL
	convertErr := runtime.DefaultUnstructuredConverter.FromUnstructured(resVal.UnstructuredContent(), &pgVal)
	if convertErr != nil {
		log.Errorf("runtime.DefaultUnstructuredConverter.FromUnstructured failed, error:%s", convertErr.Error())
		return
	}
}

func (s *PostgreSQL) Create(namespace string, pgPtr *pgv1.PostgreSQL) {
	res := s.getResource()
	client := s.getK8sClient()
	byteData, byteErr := json.Marshal(pgPtr)
	if byteErr != nil {
		log.Errorf("json.Marshal failed, error:%s", byteErr.Error())
		return
	}

	unstructuredPtr := &unstructured.Unstructured{}
	err := json.Unmarshal(byteData, &unstructuredPtr.Object)
	if err != nil {
		log.Errorf("json.Unmarshal failed, error:%s", err.Error())
		return
	}

	resVal, resErr := client.Resource(res).Create(context.TODO(), unstructuredPtr, metav1.CreateOptions{})
	if resErr != nil {
		log.Errorf("s.client.Resource(res).Namespace(namespace).Create Postgresql failed, namespace:%s, error:%s", namespace, resErr.Error())
		return
	}

	var pgVal pgv1.PostgreSQL
	convertErr := runtime.DefaultUnstructuredConverter.FromUnstructured(resVal.UnstructuredContent(), &pgVal)
	if convertErr != nil {
		log.Errorf("runtime.DefaultUnstructuredConverter.FromUnstructured failed, error:%s", convertErr.Error())
		return
	}
}
