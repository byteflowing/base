package service

import (
	"context"

	"github.com/byteflowing/base/pkg/sdk/alibaba/amap"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
)

type AmapImpl struct {
	client *amap.MapService

	manager *MapManager
}

func NewAmapImpl(manager *MapManager) *AmapImpl {
	return &AmapImpl{
		client:  amap.NewMapService(),
		manager: manager,
	}
}

func (a *AmapImpl) Source() enumsv1.MapSource {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) GetDistricts(ctx context.Context, req *mapsv1.GetDistrictsReq) (*mapsv1.GetDistrictsResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) GetDistrictChildren(ctx context.Context, req *mapsv1.GetDistrictChildrenReq) (*mapsv1.GetDistrictChildrenResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) SearchDistrict(ctx context.Context, req *mapsv1.SearchDistrictReq) (*mapsv1.SearchDistrictResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) ConvertLocation(ctx context.Context, req *mapsv1.ConvertLocationReq) (*mapsv1.ConvertLocationResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) ParseLocationToAddr(ctx context.Context, req *mapsv1.ParseLocationToAddrReq) (*mapsv1.ParseLocationToAddrResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) ParseAddrToLocation(ctx context.Context, req *mapsv1.ParseAddrToLocationReq) (*mapsv1.ParseAddrToLocationResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) WalkingRoutePlan(ctx context.Context, req *mapsv1.WalkingRoutePlanReq) (*mapsv1.WalkingRoutePlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) BicyclingRoutePlan(ctx context.Context, req *mapsv1.BicyclingRoutePlanReq) (*mapsv1.BicyclingRoutePlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) EBicyclingRoutePlan(ctx context.Context, req *mapsv1.EBicyclingRoutePlanReq) (*mapsv1.EBicyclingRoutePlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) DrivingRoutePlan(ctx context.Context, req *mapsv1.DrivingRoutePlanReq) (*mapsv1.DrivingRoutePlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) TransitRoutePlan(ctx context.Context, req *mapsv1.TransitRoutePlanReq) (*mapsv1.TransitRoutePlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) WalkingDistanceMatrixPlan(ctx context.Context, req *mapsv1.WalkingDistanceMatrixPlanReq) (*mapsv1.WalkingDistanceMatrixPlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) BicyclingDistanceMatrixPlan(ctx context.Context, req *mapsv1.BicyclingDistanceMatrixPlanReq) (*mapsv1.BicyclingDistanceMatrixPlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) DrivingDistanceMatrixPlan(ctx context.Context, req *mapsv1.DrivingDistanceMatrixPlanReq) (*mapsv1.DrivingDistanceMatrixPlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) GetLocationByIp(ctx context.Context, req *mapsv1.GetLocationByIpReq) (*mapsv1.GetLocationByIpResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) DistanceMeasure(ctx context.Context, req *mapsv1.DistanceMeasureReq) (*mapsv1.DistanceMeasureResp, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AmapImpl) GetTimezoneByLocation(ctx context.Context, req *mapsv1.GetTimezoneByLocationReq) (*mapsv1.GetTimezoneByLocationResp, error) {
	//TODO implement me
	panic("implement me")
}
