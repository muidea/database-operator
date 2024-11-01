package crds

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const Postgresql = "postgresqls"

type Spec struct {
	Image string `json:"image"`
}

type Status struct {
}

type PostgreSQL struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Spec   `json:"spec"`
	Status Status `json:"status,omitempty"`
}

type PostgreSQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []PostgreSQL `json:"items"`
}
