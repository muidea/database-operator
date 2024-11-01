package mariadb

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"

	engine "github.com/muidea/magicEngine/http"

	"supos.ai/operator/database/internal/core/module/mariadb/biz"
	"supos.ai/operator/database/pkg/common"
)

func init() {
	module.Register(New())
}

type Mariadb struct {
	routeRegistry engine.RouteRegistry

	biz *biz.Mariadb
}

func New() *Mariadb {
	return &Mariadb{}
}

func (s *Mariadb) ID() string {
	return common.MariadbModule
}

func (s *Mariadb) BindRegistry(routeRegistry engine.RouteRegistry) {

	s.routeRegistry = routeRegistry
}

func (s *Mariadb) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	s.biz = biz.New(eventHub, backgroundRoutine)
}
