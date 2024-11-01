package biz

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/cache"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/task"

	"supos.ai/operator/database/internal/core/base/biz"
	"supos.ai/operator/database/pkg/common"
)

const haSuffix = "-4ha"

func getClusterConfig() (config *rest.Config, err error) {
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		var addrs []string
		addrs, err = net.LookupHost("kubernetes.default.svc")
		if err != nil {
			return
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}

	return rest.InClusterConfig()
}

type K8s struct {
	biz.Base

	serviceCache cache.KVCache

	clientSet    *kubernetes.Clientset
	clientConfig *rest.Config
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *K8s {
	clusterConfig, configErr := getClusterConfig()
	if configErr != nil {
		panic(configErr)
	}

	clusterClient, clientErr := kubernetes.NewForConfig(clusterConfig)
	if clientErr != nil {
		panic(clientErr)
	}

	ptr := &K8s{
		Base:         biz.New(common.K8sModule, eventHub, backgroundRoutine),
		serviceCache: cache.NewKVCache(nil),
		clientConfig: clusterConfig,
		clientSet:    clusterClient,
	}

	ptr.SubscribeFunc(common.ExecuteCommand, ptr.ExecuteCommand)
	ptr.SubscribeFunc(common.StartService, ptr.StartService)
	ptr.SubscribeFunc(common.StopService, ptr.StopService)
	ptr.SubscribeFunc(common.JobService, ptr.JobService)
	ptr.SubscribeFunc(common.ListService, ptr.ListService)
	ptr.SubscribeFunc(common.QueryService, ptr.QueryService)
	ptr.SubscribeFunc(common.CreateService, ptr.CreateService)
	ptr.SubscribeFunc(common.DestroyService, ptr.DestroyService)
	return ptr
}

func (s *K8s) getNamespace() string {
	namespace, found := os.LookupEnv("NAMESPACE")
	if !found {
		namespace = corev1.NamespaceDefault
	}
	return namespace
}

func (s *K8s) Run() {
	s.AsyncTask(func() {
		// 创建一个Watcher来监视Deployment资源变化
		watcher, err := s.clientSet.AppsV1().Deployments(s.getNamespace()).Watch(context.TODO(), metav1.ListOptions{
			LabelSelector: common.GetDefaultLabels(),
		})
		if err != nil {
			log.Criticalf("watch deployment failed, error:%s", err.Error())
			panic(err)
		}

		// 循环监听Watcher的事件
		for event := range watcher.ResultChan() {
			deployment, ok := event.Object.(*appv1.Deployment)
			if !ok {
				log.Errorf("Unexpected object type:%v", event.Object)
				continue
			}

			// 根据事件类型执行相应操作
			switch event.Type {
			case watch.Added, watch.Modified:
				s.addService(deployment)
			case watch.Deleted:
				s.delService(deployment)
			case watch.Error:
				log.Warnf("Error occurred, object type:%v", event.Object)
			}
		}

		// 关闭Watcher
		watcher.Stop()
	})
}

func (s *K8s) addService(deploymentPtr *appv1.Deployment) {
}

func (s *K8s) delService(deploymentPtr *appv1.Deployment) {
	serviceName, serviceCatalog := s.getServiceName(deploymentPtr)
	if serviceCatalog == "" {
		return
	}

	serviceVal := s.serviceCache.Fetch(serviceName)

	values := event.NewValues()
	values.Set(event.Action, event.Del)
	s.BroadCast(common.NotifyService, values, serviceVal.(*common.ServiceInfo))
	s.serviceCache.Remove(serviceName)
}

func (s *K8s) getServiceName(deploymentPtr *appv1.Deployment) (name, catalog string) {
	nameVal := deploymentPtr.ObjectMeta.GetName()
	if strings.Index(nameVal, common.PostgreSQL) != -1 {
		name = nameVal
		catalog = common.PostgreSQL
		return
	}

	return
}

func (s *K8s) Create(serviceName, catalog string) (err *cd.Result) {
	return
}

func (s *K8s) Destroy(serviceName, catalog string) (err *cd.Result) {
	serviceInfo, serviceErr := s.Query(serviceName, catalog)
	if serviceErr != nil {
		err = serviceErr
		return
	}

	err = s.destroyService(serviceInfo)
	return
}

func (s *K8s) Start(serviceName, catalog string) (err *cd.Result) {
	serviceInfo, serviceErr := s.Query(serviceName, catalog)
	if serviceErr != nil {
		err = serviceErr
		return
	}

	err = s.startService(serviceInfo)
	return
}

func (s *K8s) Stop(serviceName, catalog string) (err *cd.Result) {
	serviceInfo, serviceErr := s.Query(serviceName, catalog)
	if serviceErr != nil {
		err = serviceErr
		return
	}

	err = s.stopService(serviceInfo)
	return
}

func (s *K8s) Query(serviceName, catalog string) (ret *common.ServiceInfo, err *cd.Result) {
	serviceVal := s.serviceCache.Fetch(serviceName)
	if serviceVal == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("%s not exist", serviceName))
		return
	}
	servicePtr := serviceVal.(*common.ServiceInfo)
	if servicePtr.Catalog != catalog {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("%s missmatch catalog", serviceName))
		return
	}

	ret = servicePtr
	return
}
