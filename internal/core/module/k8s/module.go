package k8s

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"

	engine "github.com/muidea/magicEngine/http"

	"supos.ai/operator/database/internal/core/module/k8s/biz"
	"supos.ai/operator/database/internal/core/module/k8s/service"
	"supos.ai/operator/database/pkg/common"
)

func init() {
	module.Register(New())
}

type K8s struct {
	routeRegistry engine.RouteRegistry

	service *service.K8s
	biz     *biz.K8s
}

func New() *K8s {
	return &K8s{}
}

func (s *K8s) ID() string {
	return common.K8sModule
}

func (s *K8s) BindRegistry(routeRegistry engine.RouteRegistry) {

	s.routeRegistry = routeRegistry
}

func (s *K8s) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	s.biz = biz.New(eventHub, backgroundRoutine)

	s.service = service.New(endpointName, s.biz)
	s.service.BindRegistry(s.routeRegistry)
	s.service.RegisterRoute()
}

func (s *K8s) Run() {
	if s.biz != nil {
		s.biz.Run()
	}
}
