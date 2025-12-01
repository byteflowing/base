package service

import (
	"context"

	"github.com/byteflowing/base/pkg/sdk/tianditu"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
)

type TiandituImpl struct {
	client  *tianditu.MapService
	manager *MapManager
}

func NewTiandituImpl(manager *MapManager) *TiandituImpl {
	return &TiandituImpl{
		client:  tianditu.NewMapService(),
		manager: manager,
	}
}

func (t *TiandituImpl) Source() enumsv1.MapSource {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) GetDistricts(ctx context.Context, req *mapsv1.GetDistrictsReq) (*mapsv1.GetDistrictsResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) GetDistrictChildren(ctx context.Context, req *mapsv1.GetDistrictChildrenReq) (*mapsv1.GetDistrictChildrenResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) SearchDistrict(ctx context.Context, req *mapsv1.SearchDistrictReq) (*mapsv1.SearchDistrictResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) ConvertLocation(ctx context.Context, req *mapsv1.ConvertLocationReq) (*mapsv1.ConvertLocationResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) ParseLocationToAddr(ctx context.Context, req *mapsv1.ParseLocationToAddrReq) (*mapsv1.ParseLocationToAddrResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) ParseAddrToLocation(ctx context.Context, req *mapsv1.ParseAddrToLocationReq) (*mapsv1.ParseAddrToLocationResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) WalkingRoutePlan(ctx context.Context, req *mapsv1.WalkingRoutePlanReq) (*mapsv1.WalkingRoutePlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) BicyclingRoutePlan(ctx context.Context, req *mapsv1.BicyclingRoutePlanReq) (*mapsv1.BicyclingRoutePlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) EBicyclingRoutePlan(ctx context.Context, req *mapsv1.EBicyclingRoutePlanReq) (*mapsv1.EBicyclingRoutePlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) DrivingRoutePlan(ctx context.Context, req *mapsv1.DrivingRoutePlanReq) (*mapsv1.DrivingRoutePlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) TransitRoutePlan(ctx context.Context, req *mapsv1.TransitRoutePlanReq) (*mapsv1.TransitRoutePlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) WalkingDistanceMatrixPlan(ctx context.Context, req *mapsv1.WalkingDistanceMatrixPlanReq) (*mapsv1.WalkingDistanceMatrixPlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) BicyclingDistanceMatrixPlan(ctx context.Context, req *mapsv1.BicyclingDistanceMatrixPlanReq) (*mapsv1.BicyclingDistanceMatrixPlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) DrivingDistanceMatrixPlan(ctx context.Context, req *mapsv1.DrivingDistanceMatrixPlanReq) (*mapsv1.DrivingDistanceMatrixPlanResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) GetLocationByIp(ctx context.Context, req *mapsv1.GetLocationByIpReq) (*mapsv1.GetLocationByIpResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) DistanceMeasure(ctx context.Context, req *mapsv1.DistanceMeasureReq) (*mapsv1.DistanceMeasureResp, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TiandituImpl) GetTimezoneByLocation(ctx context.Context, req *mapsv1.GetTimezoneByLocationReq) (*mapsv1.GetTimezoneByLocationResp, error) {
	//TODO implement me
	panic("implement me")
}
