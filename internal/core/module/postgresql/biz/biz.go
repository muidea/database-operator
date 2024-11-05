package biz

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/cache"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/task"

	"supos.ai/operator/database/internal/core/base/biz"
	"supos.ai/operator/database/pkg/common"

	pgv1 "supos.ai/operator/database/pkg/crds/v1"
)

type serviceInfo struct {
	serviceInfo   *common.ServiceInfo
	postgreSQLPtr *pgv1.PostgreSQL
}

type PostgreSQL struct {
	biz.Base

	postgresqlCache cache.KVCache
	client          dynamic.Interface
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *PostgreSQL {
	ptr := &PostgreSQL{
		Base:            biz.New(common.PostgreSQLModule, eventHub, backgroundRoutine),
		postgresqlCache: cache.NewKVCache(nil),
	}

	ptr.SubscribeFunc(common.NotifyTimer, ptr.timerCheck)
	ptr.SubscribeFunc(common.NotifyService, ptr.serviceNotify)
	return ptr
}

func (s *PostgreSQL) timerCheck(_ event.Event, _ event.Result) {
	s.List("default")

	s.serviceVerify()
}

func (s *PostgreSQL) serviceNotify(ev event.Event, _ event.Result) {
	serviceInfoPtr, serviceInfoOK := ev.Data().(*common.ServiceInfo)
	if !serviceInfoOK {
		return
	}

	curPtr := s.postgresqlCache.Fetch(serviceInfoPtr.Name)
	if curPtr == nil {
		infoPtr := &serviceInfo{
			serviceInfo: serviceInfoPtr,
		}

		s.postgresqlCache.Put(serviceInfoPtr.Name, infoPtr, cache.ForeverAgeValue)
		return
	}

	infoPtr := curPtr.(*serviceInfo)
	infoPtr.serviceInfo = serviceInfoPtr
	s.postgresqlCache.Put(serviceInfoPtr.Name, infoPtr, cache.ForeverAgeValue)
}

func (s *PostgreSQL) serviceVerify() {
	postgresqlList := s.postgresqlCache.GetAll()
	for _, val := range postgresqlList {
		serviceInfoPtr, serviceInfoOK := val.(*serviceInfo)
		if !serviceInfoOK {
			continue
		}
		if serviceInfoPtr.serviceInfo == nil {
			// create postgresql k8s deployment...
			s.createK8sDeployment(serviceInfoPtr.postgreSQLPtr)
			continue
		}
		if serviceInfoPtr.postgreSQLPtr == nil {
			// create postgresql crd instance...
			continue
		}
	}
}

func (s *PostgreSQL) createK8sDeployment(pgPtr *pgv1.PostgreSQL) {
	pgName := pgPtr.GetName()
	pgNamespace := pgPtr.GetNamespace()
	pgServicePtr := common.NewPostgreSQLService(pgName, pgNamespace)

	createEvent := event.NewEvent(common.CreateService, s.ID(), common.K8sModule, nil, pgServicePtr)
	s.PostEvent(createEvent)
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
	for _, val := range pgList.Items {

		curPtr := s.postgresqlCache.Fetch(val.Name)
		if curPtr == nil {
			infoPtr := &serviceInfo{
				postgreSQLPtr: &val,
			}

			s.postgresqlCache.Put(val.Name, infoPtr, cache.ForeverAgeValue)
			continue
		}

		infoPtr := curPtr.(*serviceInfo)
		infoPtr.postgreSQLPtr = &val
		s.postgresqlCache.Put(val.Name, infoPtr, cache.ForeverAgeValue)
	}
}

func (s *PostgreSQL) Get(namespace, name string) (ret *pgv1.PostgreSQL, err *cd.Result) {
	val := s.postgresqlCache.Fetch(name)
	if val != nil {
		ret = val.(*pgv1.PostgreSQL)
		return
	}

	res := s.getResource()
	client := s.getK8sClient()
	resVal, resErr := client.Resource(res).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if resErr != nil {
		log.Errorf("s.client.Resource(res).Namespace(namespace).Get Postgresql failed, namespace:%s, name:%s, error:%s", namespace, name, resErr.Error())
		return
	}

	pgVal := &pgv1.PostgreSQL{}
	convertErr := runtime.DefaultUnstructuredConverter.FromUnstructured(resVal.UnstructuredContent(), pgVal)
	if convertErr != nil {
		log.Errorf("runtime.DefaultUnstructuredConverter.FromUnstructured failed, error:%s", convertErr.Error())
		return
	}

	s.postgresqlCache.Put(pgVal.Name, pgVal, cache.ForeverAgeValue)
	ret = pgVal
	return
}

func (s *PostgreSQL) Create(namespace string, pgPtr *pgv1.PostgreSQL) (ret *pgv1.PostgreSQL, err *cd.Result) {
	res := s.getResource()
	client := s.getK8sClient()

	unstructuredPtr, unstructuredErr := runtime.DefaultUnstructuredConverter.ToUnstructured(pgPtr)
	if unstructuredErr != nil {
		log.Errorf("json.Unmarshal failed, error:%s", err.Error())
		return
	}

	resVal, resErr := client.Resource(res).Create(context.TODO(), &unstructured.Unstructured{
		Object: unstructuredPtr,
	}, metav1.CreateOptions{})
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

	return
}
