package common

const (
	ExecuteCommand = "/command/execute"
	GetK8sConfig   = "/config/get"
	CreateService  = "/service/create"
	DestroyService = "/service/destroy"
	StartService   = "/service/start"
	StopService    = "/service/stop"
	JobService     = "/service/job"
	ListService    = "/service/list"
	QueryService   = "/service/query"
	NotifyService  = "/service/notify"
)

const K8sModule = "/module/k8s"

var PostgreSQLDefaultSpec = Spec{
	CPU:    "2",
	Memory: "4Gi",
}

type Endpoint Svc
