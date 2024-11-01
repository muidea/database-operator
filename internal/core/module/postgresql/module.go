package postgresql

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"

	engine "github.com/muidea/magicEngine/http"

	"supos.ai/operator/database/internal/core/module/postgresql/biz"
	"supos.ai/operator/database/pkg/common"
)

func init() {
	module.Register(New())
}

type PostgreSQL struct {
	routeRegistry engine.RouteRegistry

	biz *biz.PostgreSQL
}

func New() *PostgreSQL {
	return &PostgreSQL{}
}

func (s *PostgreSQL) ID() string {
	return common.PostgreSQLModule
}

func (s *PostgreSQL) BindRegistry(routeRegistry engine.RouteRegistry) {

	s.routeRegistry = routeRegistry
}

func (s *PostgreSQL) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	s.biz = biz.New(eventHub, backgroundRoutine)
}
