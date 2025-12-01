package service

import (
	"context"
	_ "embed"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/sdk/huawei/lbs"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
)

type HuaweiImpl struct {
	client    *lbs.MapService
	manager   *MapManager
	maxMatrix int
}

func NewHuaweiImpl(manager *MapManager) *HuaweiImpl {
	return &HuaweiImpl{
		client:    lbs.NewMapService(),
		manager:   manager,
		maxMatrix: 100,
	}
}

func (h *HuaweiImpl) Source() enumsv1.MapSource {
	return enumsv1.MapSource_MAP_SOURCE_HUAWEI
}

func (h *HuaweiImpl) ParseLocationToAddr(ctx context.Context, req *mapsv1.ParseLocationToAddrReq) (*mapsv1.ParseLocationToAddrResp, error) {
	summary, err := h.manager.GetInterfaceSummaryWithLimit(ctx, req.Source, enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_LOCATION_TO_ADDR, 1, req.MapId)
	if err != nil {
		return nil, err
	}
	// Set the key in the request
	huaweiReq := req.GetHuawei()
	huaweiReq.Key = summary.Key

	res, err := h.client.LocationToAddr(ctx, huaweiReq)
	if err != nil {
		return nil, err
	}
	return &mapsv1.ParseLocationToAddrResp{Huawei: res}, nil
}

func (h *HuaweiImpl) ParseAddrToLocation(ctx context.Context, req *mapsv1.ParseAddrToLocationReq) (*mapsv1.ParseAddrToLocationResp, error) {
	summary, err := h.manager.GetInterfaceSummaryWithLimit(ctx, req.Source, enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ADDR_TO_LOCATION, 1, req.MapId)
	if err != nil {
		return nil, err
	}

	// Set the key in the request
	huaweiReq := req.GetHuawei()
	huaweiReq.Key = summary.Key

	res, err := h.client.AddrToLocation(ctx, huaweiReq)
	if err != nil {
		return nil, err
	}
	return &mapsv1.ParseAddrToLocationResp{Huawei: res}, nil
}

func (h *HuaweiImpl) WalkingRoutePlan(ctx context.Context, req *mapsv1.WalkingRoutePlanReq) (*mapsv1.WalkingRoutePlanResp, error) {
	summary, err := h.manager.GetInterfaceSummaryWithLimit(ctx, req.Source, enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_WALKING, 1, req.MapId)
	if err != nil {
		return nil, err
	}

	// Set the key in the request
	huaweiReq := req.GetHuawei()
	huaweiReq.Key = summary.Key

	res, err := h.client.RoutePlan(ctx, huaweiReq)
	if err != nil {
		return nil, err
	}
	return &mapsv1.WalkingRoutePlanResp{Huawei: res}, nil
}

func (h *HuaweiImpl) BicyclingRoutePlan(ctx context.Context, req *mapsv1.BicyclingRoutePlanReq) (*mapsv1.BicyclingRoutePlanResp, error) {
	summary, err := h.manager.GetInterfaceSummaryWithLimit(ctx, req.Source, enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_BICYCLING, 1, req.MapId)
	if err != nil {
		return nil, err
	}

	// Set the key in the request
	huaweiReq := req.GetHuawei()
	huaweiReq.Key = summary.Key

	res, err := h.client.RoutePlan(ctx, huaweiReq)
	if err != nil {
		return nil, err
	}
	return &mapsv1.BicyclingRoutePlanResp{Huawei: res}, nil
}

func (h *HuaweiImpl) DrivingRoutePlan(ctx context.Context, req *mapsv1.DrivingRoutePlanReq) (*mapsv1.DrivingRoutePlanResp, error) {
	summary, err := h.manager.GetInterfaceSummaryWithLimit(ctx, req.Source, enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_DRIVING, 1, req.MapId)
	if err != nil {
		return nil, err
	}

	// Set the key in the request
	huaweiReq := req.GetHuawei()
	huaweiReq.Key = summary.Key

	res, err := h.client.RoutePlan(ctx, huaweiReq)
	if err != nil {
		return nil, err
	}
	return &mapsv1.DrivingRoutePlanResp{Huawei: res}, nil
}

func (h *HuaweiImpl) WalkingDistanceMatrixPlan(ctx context.Context, req *mapsv1.WalkingDistanceMatrixPlanReq) (*mapsv1.WalkingDistanceMatrixPlanResp, error) {
	huaweiReq := req.GetHuawei()
	request, err := h.getMatrixRequest(huaweiReq)
	if err != nil {
		return nil, err
	}
	summary, err := h.manager.GetInterfaceSummaryWithLimit(ctx, req.Source, enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_WALKING, int64(request), req.MapId)
	if err != nil {
		return nil, err
	}

	// Set the key in the request
	huaweiReq.Key = summary.Key

	res, err := h.client.DistanceMatrix(ctx, huaweiReq)
	if err != nil {
		return nil, err
	}
	return &mapsv1.WalkingDistanceMatrixPlanResp{Huawei: res}, nil
}

func (h *HuaweiImpl) BicyclingDistanceMatrixPlan(ctx context.Context, req *mapsv1.BicyclingDistanceMatrixPlanReq) (*mapsv1.BicyclingDistanceMatrixPlanResp, error) {
	huaweiReq := req.GetHuawei()
	request, err := h.getMatrixRequest(huaweiReq)
	if err != nil {
		return nil, err
	}
	summary, err := h.manager.GetInterfaceSummaryWithLimit(ctx, req.Source, enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_BICYCLING, int64(request), req.MapId)
	if err != nil {
		return nil, err
	}
	huaweiReq.Key = summary.Key
	res, err := h.client.DistanceMatrix(ctx, huaweiReq)
	if err != nil {
		return nil, err
	}
	return &mapsv1.BicyclingDistanceMatrixPlanResp{Huawei: res}, nil
}

func (h *HuaweiImpl) DrivingDistanceMatrixPlan(ctx context.Context, req *mapsv1.DrivingDistanceMatrixPlanReq) (*mapsv1.DrivingDistanceMatrixPlanResp, error) {
	huaweiReq := req.GetHuawei()
	request, err := h.getMatrixRequest(huaweiReq)
	if err != nil {
		return nil, err
	}
	summary, err := h.manager.GetInterfaceSummaryWithLimit(ctx, req.Source, enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_DRIVING, int64(request), req.MapId)
	if err != nil {
		return nil, err
	}
	huaweiReq.Key = summary.Key

	res, err := h.client.DistanceMatrix(ctx, huaweiReq)
	if err != nil {
		return nil, err
	}
	return &mapsv1.DrivingDistanceMatrixPlanResp{Huawei: res}, nil
}

func (h *HuaweiImpl) GetLocationByIp(ctx context.Context, req *mapsv1.GetLocationByIpReq) (*mapsv1.GetLocationByIpResp, error) {
	summary, err := h.manager.GetInterfaceSummaryWithLimit(ctx, req.Source, enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_GET_LOCATION_BY_IP, 1, req.MapId)
	if err != nil {
		return nil, err
	}

	// Set the key in the request
	huaweiReq := req.GetHuawei()
	huaweiReq.Key = summary.Key

	res, err := h.client.GetLocationByIp(ctx, huaweiReq)
	if err != nil {
		return nil, err
	}
	return &mapsv1.GetLocationByIpResp{Huawei: res}, nil
}

func (h *HuaweiImpl) GetTimezoneByLocation(ctx context.Context, req *mapsv1.GetTimezoneByLocationReq) (*mapsv1.GetTimezoneByLocationResp, error) {
	summary, err := h.manager.GetInterfaceSummaryWithLimit(ctx, req.Source, enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_GET_TIMEZONE_BY_LOCATION, 1, req.MapId)
	if err != nil {
		return nil, err
	}

	// Set the key in the request
	huaweiReq := req.GetHuawei()
	huaweiReq.Key = summary.Key

	res, err := h.client.GetTimezone(ctx, huaweiReq)
	if err != nil {
		return nil, err
	}
	return &mapsv1.GetTimezoneByLocationResp{Huawei: res}, nil
}

//----------------------------------------------------------------------------------------------------------------------
// unsupported interface

func (h *HuaweiImpl) GetDistricts(ctx context.Context, req *mapsv1.GetDistrictsReq) (*mapsv1.GetDistrictsResp, error) {
	return nil, ecode.ErrUnImplemented
}

func (h *HuaweiImpl) GetDistrictChildren(ctx context.Context, req *mapsv1.GetDistrictChildrenReq) (*mapsv1.GetDistrictChildrenResp, error) {
	return nil, ecode.ErrUnImplemented
}

func (h *HuaweiImpl) SearchDistrict(ctx context.Context, req *mapsv1.SearchDistrictReq) (*mapsv1.SearchDistrictResp, error) {
	return nil, ecode.ErrUnImplemented
}

func (h *HuaweiImpl) ConvertLocation(ctx context.Context, req *mapsv1.ConvertLocationReq) (*mapsv1.ConvertLocationResp, error) {
	return nil, ecode.ErrUnImplemented
}

func (h *HuaweiImpl) EBicyclingRoutePlan(ctx context.Context, req *mapsv1.EBicyclingRoutePlanReq) (*mapsv1.EBicyclingRoutePlanResp, error) {
	return nil, ecode.ErrUnImplemented
}

func (h *HuaweiImpl) TransitRoutePlan(ctx context.Context, req *mapsv1.TransitRoutePlanReq) (*mapsv1.TransitRoutePlanResp, error) {
	return nil, ecode.ErrUnImplemented
}

func (h *HuaweiImpl) DistanceMeasure(ctx context.Context, req *mapsv1.DistanceMeasureReq) (*mapsv1.DistanceMeasureResp, error) {
	return nil, ecode.ErrUnImplemented
}

func (h *HuaweiImpl) getMatrixRequest(req *mapsv1.HuaweiDistanceMatrixPlanReq) (request int, err error) {
	origLen := len(req.Origin)
	destLen := len(req.Destination)
	request = origLen * destLen
	if request > h.maxMatrix {
		return 0, ecode.ErrMapsMatrixTooManyPoints
	}
	return request, nil
}
