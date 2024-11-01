package common

const (
	DefaultMariadbImage      = "registry.supos.ai/jenkins/mariadb:10.6.11"
	DefaultMariadbRoot       = "root"
	DefaultMariadbPassword   = "rootkit"
	DefaultMariadbConfigPath = "/etc/conf.d"
	DefaultMariadbDataPath   = "/var/lib/mysql"
	DefaultMariadbBackPath   = "/backup"
	DefaultMariadbPort       = 3306
)

const MariadbModule = "/module/mariadb"
