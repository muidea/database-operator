package service

import (
	"context"
	"net/http"

	engine "github.com/muidea/magicEngine/http"

	cd "github.com/muidea/magicCommon/def"
	fn "github.com/muidea/magicCommon/foundation/net"

	"supos.ai/operator/database/internal/core/module/k8s/biz"
	"supos.ai/operator/database/pkg/common"
)

// K8s BaseService
type K8s struct {
	routeRegistry engine.RouteRegistry

	bizPtr *biz.K8s

	endpointName string
}

// New create base
func New(endpointName string, bizPtr *biz.K8s) *K8s {
	ptr := &K8s{
		endpointName: endpointName,
		bizPtr:       bizPtr,
	}

	return ptr
}

func (s *K8s) BindRegistry(
	routeRegistry engine.RouteRegistry) {

	s.routeRegistry = routeRegistry

	s.routeRegistry.SetApiVersion(common.ApiVersion)
}

// RegisterRoute 注册路由
func (s *K8s) RegisterRoute() {
	createRoute := engine.CreateRoute(common.CreateService, engine.POST, s.CreateHandle)
	s.routeRegistry.AddRoute(createRoute)

	destroyRoute := engine.CreateRoute(common.DestroyService, engine.POST, s.DestroyHandle)
	s.routeRegistry.AddRoute(destroyRoute)

	startRoute := engine.CreateRoute(common.StartService, engine.POST, s.StartHandle)
	s.routeRegistry.AddRoute(startRoute)

	stopRoute := engine.CreateRoute(common.StopService, engine.POST, s.StopHandle)
	s.routeRegistry.AddRoute(stopRoute)

	queryRoute := engine.CreateRoute(common.QueryService, engine.POST, s.QueryHandle)
	s.routeRegistry.AddRoute(queryRoute)
}

func (s *K8s) CreateHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.CreateServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		createErr := s.bizPtr.Create(param.Name, param.Catalog)
		if createErr != nil {
			result.Result = *createErr
			break
		}

		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *K8s) DestroyHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.DestroyServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		destroyErr := s.bizPtr.Destroy(param.Name, param.Catalog)
		if destroyErr != nil {
			result.Result = *destroyErr
			break
		}
		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *K8s) StartHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.StartServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		startErr := s.bizPtr.Start(param.Name, param.Catalog)
		if startErr != nil {
			result.Result = *startErr
			break
		}

		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *K8s) StopHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.StopServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		stopErr := s.bizPtr.Stop(param.Name, param.Catalog)
		if stopErr != nil {
			result.Result = *stopErr
			break
		}
		break
	}

	fn.PackageHTTPResponse(res, result)
}

func (s *K8s) QueryHandle(_ context.Context, res http.ResponseWriter, req *http.Request) {
	result := &common.QueryServiceResult{}
	for {
		param := &common.ServiceParam{}
		err := fn.ParseJSONBody(req, nil, param)
		if err != nil {
			result.ErrorCode = cd.IllegalParam
			result.Reason = "非法参数"
			break
		}
		serviceInfo, serviceErr := s.bizPtr.Query(param.Name, param.Catalog)
		if serviceErr != nil {
			result.Result = *serviceErr
			break
		}

		result.ServiceInfo = serviceInfo
		break
	}

	fn.PackageHTTPResponse(res, result)
}
