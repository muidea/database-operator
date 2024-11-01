package base

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"
	engine "github.com/muidea/magicEngine/http"

	"supos.ai/operator/database/internal/core/kernel/base/biz"
	"supos.ai/operator/database/internal/core/kernel/base/service"
	"supos.ai/operator/database/pkg/common"
)

func init() {
	module.Register(New())
}

type Base struct {
	routeRegistry engine.RouteRegistry

	service *service.Base
	biz     *biz.Base
}

func New() *Base {
	return &Base{}
}

func (s *Base) ID() string {
	return common.BaseModule
}

func (s *Base) BindRegistry(routeRegistry engine.RouteRegistry) {

	s.routeRegistry = routeRegistry
}

func (s *Base) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	s.biz = biz.New(eventHub, backgroundRoutine)

	s.service = service.New(endpointName, s.biz)
	s.service.BindRegistry(s.routeRegistry)
	s.service.RegisterRoute()
}
