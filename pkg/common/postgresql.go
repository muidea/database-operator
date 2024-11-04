package common

const (
	DefaultPostgreSQLImage    = "registry.supos.ai/jenkins/mariadb:10.6.11"
	DefaultPostgreSQLRoot     = "root"
	DefaultPostgreSQLPassword = "rootkit"
)

var PostgreSQLDefaultSpec = Spec{
	CPU:    "2",
	Memory: "4Gi",
}

const PostgreSQLModule = "/module/postgresql"
