package biz

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/cache"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/task"

	"supos.ai/operator/database/internal/core/base/biz"
	"supos.ai/operator/database/pkg/common"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getClusterConfig() (config *rest.Config, err error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	return
	/*
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
	*/

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
	ptr.SubscribeFunc(common.GetK8sConfig, ptr.GetConfig)
	ptr.SubscribeFunc(common.StartService, ptr.StartService)
	ptr.SubscribeFunc(common.StopService, ptr.StopService)
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
	serviceInfo, serviceErr := s.getServiceInfoFromDeployment(deploymentPtr, s.clientSet)
	if serviceErr != nil {
		log.Errorf("addService failed, s.getServiceInfoFromDeployment %v error:%s", deploymentPtr.ObjectMeta.GetName(), serviceErr.Error())
		return
	}
	// 如果返回空，则表示是不需要处理的服务，直接调过
	if serviceInfo == nil {
		return
	}

	values := event.NewValues()
	values.Set(event.Action, event.Add)
	s.BroadCast(common.NotifyService, values, serviceInfo)

	s.serviceCache.Put(serviceInfo.Name, serviceInfo, cache.ForeverAgeValue)
}

func (s *K8s) delService(deploymentPtr *appv1.Deployment) {
	serviceName := s.getServiceName(deploymentPtr)
	serviceVal := s.serviceCache.Fetch(serviceName)

	values := event.NewValues()
	values.Set(event.Action, event.Del)
	s.BroadCast(common.NotifyService, values, serviceVal.(*common.ServiceInfo))

	s.serviceCache.Remove(serviceName)
}

func (s *K8s) getServiceName(deploymentPtr *appv1.Deployment) (name string) {
	return deploymentPtr.ObjectMeta.GetName()
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

func (s *K8s) getServiceInfoFromDeployment(deploymentPtr *appv1.Deployment, clientSet *kubernetes.Clientset) (ret *common.ServiceInfo, err *cd.Result) {
	ptr := &common.ServiceInfo{
		Name:      deploymentPtr.ObjectMeta.GetName(),
		Namespace: deploymentPtr.ObjectMeta.GetNamespace(),
		Image:     deploymentPtr.Spec.Template.Spec.Containers[0].Image,
		Labels:    deploymentPtr.ObjectMeta.Labels,
		Spec: &common.Spec{
			CPU:    deploymentPtr.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String(),
			Memory: deploymentPtr.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String(),
		},
		Volumes:  &common.Volumes{},
		Replicas: *deploymentPtr.Spec.Replicas,
	}

	dataPath, dataErr := s.getServiceDataPath(deploymentPtr, clientSet)
	if dataErr != nil {
		err = dataErr
		log.Errorf("getServiceInfoFromDeployment failed, s.getServiceDataPath %v error:%v", deploymentPtr.ObjectMeta.GetName(), dataErr.Error())
		return
	}
	ptr.Volumes.DataPath = dataPath

	//TODO
	ptr.Catalog = common.PostgreSQL

	ret = ptr
	return
}

func (s *K8s) getServiceDataPath(deploymentPtr *appv1.Deployment, clientSet *kubernetes.Clientset) (ret *common.Path, err *cd.Result) {
	var dataVolumes *corev1.Volume
	for _, val := range deploymentPtr.Spec.Template.Spec.Volumes {
		if val.Name == deploymentPtr.ObjectMeta.GetName() {
			dataVolumes = &val
			break
		}
	}
	if dataVolumes == nil || dataVolumes.PersistentVolumeClaim == nil {
		return
	}

	pvcInfo, pvcErr := clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Get(
		context.TODO(),
		dataVolumes.PersistentVolumeClaim.ClaimName,
		metav1.GetOptions{})
	if pvcErr != nil {
		err = cd.NewError(cd.UnExpected, pvcErr.Error())
		log.Errorf("getServiceDataPath failed, clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Get %v error:%v",
			dataVolumes.PersistentVolumeClaim.ClaimName,
			pvcErr.Error())
		return
	}

	pvInfo, pvErr := clientSet.CoreV1().PersistentVolumes().Get(context.TODO(), pvcInfo.Spec.VolumeName, metav1.GetOptions{})
	if pvErr != nil {
		err = cd.NewError(cd.UnExpected, pvErr.Error())
		log.Errorf("getServiceDataPath failed, clientSet.CoreV1().PersistentVolumes().Get %v error:%v",
			pvcInfo.Spec.VolumeName,
			pvErr.Error())
		return
	}

	ret = &common.Path{
		Name:  dataVolumes.Name,
		Value: pvInfo.Spec.HostPath.Path,
		Type:  common.LocalPath,
	}
	return
}
