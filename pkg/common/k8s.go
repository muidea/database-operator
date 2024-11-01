package common

const (
	ExecuteCommand = "/command/execute"
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

const (
	HostPath  = "host-path"
	LocalPath = "local-path"
	InnerPath = "inner-path"
)

var PostgreSQLDefaultSpec = Spec{
	CPU:    "2",
	Memory: "4Gi",
}

const PostgreSQLBackupBinlogFile = "data/xtrabackup_binlog_info"

type Endpoint Svc
