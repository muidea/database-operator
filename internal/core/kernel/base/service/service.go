package service

import (
	engine "github.com/muidea/magicEngine/http"

	"supos.ai/operator/database/internal/core/kernel/base/biz"
	"supos.ai/operator/database/pkg/common"
)

// Base BaseService
type Base struct {
	routeRegistry engine.RouteRegistry

	bizPtr *biz.Base

	endpointName string
}

// New create base
func New(endpointName string, bizPtr *biz.Base) *Base {
	ptr := &Base{
		endpointName: endpointName,
		bizPtr:       bizPtr,
	}

	return ptr
}

func (s *Base) BindRegistry(
	routeRegistry engine.RouteRegistry) {

	s.routeRegistry = routeRegistry

	s.routeRegistry.SetApiVersion(common.ApiVersion)
}

// RegisterRoute 注册路由
func (s *Base) RegisterRoute() {

}
