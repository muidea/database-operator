package common

const (
	DefaultPostgreSQLImage    = "registry.supos.ai/jenkins/mariadb:10.6.11"
	DefaultPostgreSQLDataPath = "/var/lib/postgresql/data"
	DefaultPostgreSQLRoot     = "root"
	DefaultPostgreSQLPassword = "rootkit"
	DefaultPostgreSQLPort     = 5432
)

var PostgreSQLDefaultSpec = Spec{
	CPU:    "2",
	Memory: "4Gi",
}

func NewPostgreSQLService(name, namespace string) *ServiceInfo {
	return &ServiceInfo{
		Name:      name,
		Namespace: namespace,
		Catalog:   PostgreSQL,
		Image:     DefaultPostgreSQLImage,
		Labels:    DefaultLabels,
		Spec:      &PostgreSQLDefaultSpec,
		Volumes: &Volumes{
			DataPath: &Path{
				Name:  name,
				Value: DefaultPostgreSQLDataPath,
				Type:  LocalPath,
			},
		},
		Env: &Env{
			Items: []*EnvItem{
				{
					Name:  "POSTGRES_USER",
					Value: DefaultPostgreSQLRoot,
				},
				{
					Name:  "POSTGRES_PASSWORD",
					Value: DefaultPostgreSQLPassword,
				},
			},
		},
		Svc: &Svc{
			Port: DefaultPostgreSQLPort,
		},
		Replicas: 1,
	}
}

const PostgreSQLModule = "/module/postgresql"
