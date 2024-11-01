package biz

import (
	"bytes"
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"

	"supos.ai/operator/database/pkg/common"
)

func (s *K8s) execInPod(clientSet *kubernetes.Clientset, clientConfig *rest.Config, namespace, podName, containerName, command string) (stdout []byte, stderr []byte, err *cd.Result) {
	cmd := []string{
		"sh",
		"-c",
		command,
	}

	req := clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).SubResource("exec").Param("container", containerName)
	req.VersionedParams(
		&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		},
		scheme.ParameterCodec,
	)

	var stdoutBuff, stderrBuff bytes.Buffer
	execPtr, execErr := remotecommand.NewSPDYExecutor(clientConfig, "POST", req.URL())
	if execErr != nil {
		err = cd.NewError(cd.UnExpected, execErr.Error())
		log.Errorf("execInPod failed, remotecommand.NewSPDYExecutor error:%s", err.Error())
		return
	}

	execErr = execPtr.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdoutBuff,
		Stderr: &stderrBuff,
	})
	if execErr != nil {
		err = cd.NewError(cd.UnExpected, execErr.Error())
		log.Errorf("execInPod failed, execPtr.Stream error:%s", err.Error())
		return
	}

	stdout = stdoutBuff.Bytes()
	stderr = stderrBuff.Bytes()
	return
}

func (s *K8s) ExecuteCommand(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("ExecuteCommand failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK || cmdInfoPtr == nil {
		log.Warnf("ExecuteCommand failed, illegal param")
		return
	}

	deploymentPtr, deploymentErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Get(context.TODO(), cmdInfoPtr.ServiceInfo.Name, metav1.GetOptions{})
	if deploymentErr != nil {
		log.Errorf("ExecuteCommand failed, s.clientSet.AppsV1().Deployments(s.getNamespace()).Get error:%s", deploymentErr.Error())
		if re != nil {
			re.Set(nil, cd.NewError(cd.UnExpected, deploymentErr.Error()))
		}

		return
	}

	selectorPtr, selectorErr := metav1.LabelSelectorAsSelector(deploymentPtr.Spec.Selector)
	if selectorErr != nil {
		log.Errorf("ExecuteCommand failed, metav1.LabelSelectorAsSelector error:%s", selectorErr.Error())
		if re != nil {
			re.Set(nil, cd.NewError(cd.UnExpected, selectorErr.Error()))
		}

		return
	}
	podList, podsErr := s.clientSet.CoreV1().Pods(s.getNamespace()).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selectorPtr.String(),
	})
	if podsErr != nil || len(podList.Items) == 0 {
		if podsErr == nil {
			podsErr = fmt.Errorf("not exist %s pods", cmdInfoPtr.ServiceInfo.Name)
		}

		log.Errorf("ExecuteCommand failed, s.clientSet.CoreV1().Pods(s.getNamespace()).List error:%s", podsErr.Error())
		if re != nil {
			re.Set(nil, cd.NewError(cd.UnExpected, podsErr.Error()))
		}
		return
	}

	podName := podList.Items[0].Name
	containerName := podList.Items[0].Spec.Containers[0].Name
	commandVal := strings.Join(cmdInfoPtr.Command, " ")
	resultData, errorData, resultErr := s.execInPod(s.clientSet, s.clientConfig, s.getNamespace(), podName, containerName, commandVal)
	if re != nil {
		re.Set(resultData, resultErr)
		re.SetVal("stderr", errorData)
	}
}

func (s *K8s) GetConfig(_ event.Event, re event.Result) {
	if re != nil {
		re.Set(s.clientConfig, nil)
	}
}

func (s *K8s) CreateService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("CreateService failed, nil param")
		return
	}

	catalog := ev.GetData("catalog")
	if catalog == nil {
		log.Warnf("CreateService failed, illegal catalog")
		return
	}

	var err *cd.Result
	switch catalog.(string) {
	case common.PostgreSQL:
	default:
		panic(fmt.Sprintf("illegal catalog:%v", catalog))
	}

	if re != nil {
		re.Set(nil, err)
	}
}

func (s *K8s) DestroyService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("DestroyService failed, nil param")
		return
	}

	serviceName, serviceOK := param.(string)
	if !serviceOK {
		log.Warnf("DestroyService failed, illegal param")
		return
	}
	catalog := ev.GetData("catalog")
	if catalog == nil {
		log.Warnf("DestroyService failed, illegal catalog")
		return
	}

	var err *cd.Result
	switch catalog.(string) {
	case common.PostgreSQL:
		serviceInfo, serviceErr := s.Query(serviceName, common.PostgreSQL)
		if serviceErr == nil {
			err = s.destroyService(serviceInfo)
		} else {
			err = serviceErr
		}
	default:
		panic(fmt.Sprintf("illegal catalog:%v", catalog))
	}

	if re != nil {
		re.Set(nil, err)
	}
}

func (s *K8s) StartService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("StartService failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK {
		log.Warnf("StartService failed, illegal param")
		return
	}

	err := s.startService(cmdInfoPtr.ServiceInfo)
	if re != nil {
		re.Set(nil, err)
	}
}

func (s *K8s) StopService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("StopService failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK {
		log.Warnf("StopService failed, illegal param")
		return
	}

	err := s.stopService(cmdInfoPtr.ServiceInfo)
	if re != nil {
		re.Set(nil, err)
	}
}

func (s *K8s) JobService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("StopService failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK {
		log.Warnf("StopService failed, illegal param")
		return
	}

	jobServiceInfo := *cmdInfoPtr.ServiceInfo
	jobServiceInfo.Name = cmdInfoPtr.Service

	return
}

func (s *K8s) ListService(ev event.Event, re event.Result) {
	catalog2ServiceList := s.enumService()
	if re != nil {
		re.Set(catalog2ServiceList, nil)
	}
}

func (s *K8s) enumService() common.Catalog2ServiceList {
	catalog2ServiceList := common.Catalog2ServiceList{}

	mariadbList := common.ServiceList{}
	serviceList := s.serviceCache.GetAll()
	for _, val := range serviceList {
		servicePtr := val.(*common.ServiceInfo)
		switch servicePtr.Catalog {
		case common.PostgreSQL:
			mariadbList = append(mariadbList, servicePtr.Name)
		}
	}
	if len(mariadbList) > 0 {
		catalog2ServiceList[common.PostgreSQL] = mariadbList
	}

	return catalog2ServiceList
}

func (s *K8s) QueryService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("QueryService failed, nil param")
		return
	}

	serviceName, serviceOK := param.(string)
	if !serviceOK {
		log.Warnf("QueryService failed, illegal param")
		return
	}

	catalog := ev.GetData("catalog")
	if catalog == nil {
		log.Warnf("QueryService failed, illegal catalog")
		return
	}

	serviceInfoPtr, serviceInfoErr := s.Query(serviceName, catalog.(string))
	if re != nil {
		re.Set(serviceInfoPtr, serviceInfoErr)
	}
}

func (s *K8s) createService(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	switch serviceInfo.Catalog {
	case common.PostgreSQL:
		err = s.createPostgreSQL(serviceInfo)
	}

	return
}

func (s *K8s) destroyService(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	switch serviceInfo.Catalog {
	case common.PostgreSQL:
		err = s.destroyPostgreSQL(serviceInfo)
	}

	return
}

func (s *K8s) startService(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	switch serviceInfo.Catalog {
	case common.PostgreSQL:
		err = s.startPostgreSQL(serviceInfo)
	}

	return
}

func (s *K8s) stopService(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	switch serviceInfo.Catalog {
	case common.PostgreSQL:
		err = s.stopPostgreSQL(serviceInfo)
	}

	return
}

func (s *K8s) waitForJobFinished(serviceInfo *common.ServiceInfo, job *batchv1.Job) (err error) {
	finished := false
	for !finished {
		for _, condition := range job.Status.Conditions {
			switch condition.Type {
			case batchv1.JobComplete:
				finished = true
				return s.deleteJob(job)
			case batchv1.JobFailed:
				finished = true
				err = fmt.Errorf("%v job failed: %v", serviceInfo.Name, condition.Reason)
			}
		}
		time.Sleep(time.Second * 5)
	}
	return
}

func (s *K8s) deleteJob(job *batchv1.Job) (err error) {
	err = s.clientSet.
		BatchV1().
		Jobs("default").
		Delete(context.Background(), job.Name, metav1.DeleteOptions{})
	return
}
