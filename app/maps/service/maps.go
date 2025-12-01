package service

import (
	"context"

	"github.com/byteflowing/base/app/maps/dal/model"
	"github.com/byteflowing/base/app/maps/dal/query"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/db"
	"github.com/byteflowing/base/pkg/utils/slicex"
	"github.com/byteflowing/base/pkg/utils/timex"
	"github.com/byteflowing/base/pkg/utils/trans"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IMapService interface {
	Source() enumv1.MapSource
	GetDistricts(ctx context.Context, req *mapsv1.GetDistrictsReq) (*mapsv1.GetDistrictsResp, error)
	GetDistrictChildren(ctx context.Context, req *mapsv1.GetDistrictChildrenReq) (*mapsv1.GetDistrictChildrenResp, error)
	SearchDistrict(ctx context.Context, req *mapsv1.SearchDistrictReq) (*mapsv1.SearchDistrictResp, error)
	ConvertLocation(ctx context.Context, req *mapsv1.ConvertLocationReq) (*mapsv1.ConvertLocationResp, error)
	ParseLocationToAddr(ctx context.Context, req *mapsv1.ParseLocationToAddrReq) (*mapsv1.ParseLocationToAddrResp, error)
	ParseAddrToLocation(ctx context.Context, req *mapsv1.ParseAddrToLocationReq) (*mapsv1.ParseAddrToLocationResp, error)
	WalkingRoutePlan(ctx context.Context, req *mapsv1.WalkingRoutePlanReq) (*mapsv1.WalkingRoutePlanResp, error)
	BicyclingRoutePlan(ctx context.Context, req *mapsv1.BicyclingRoutePlanReq) (*mapsv1.BicyclingRoutePlanResp, error)
	EBicyclingRoutePlan(ctx context.Context, req *mapsv1.EBicyclingRoutePlanReq) (*mapsv1.EBicyclingRoutePlanResp, error)
	DrivingRoutePlan(ctx context.Context, req *mapsv1.DrivingRoutePlanReq) (*mapsv1.DrivingRoutePlanResp, error)
	TransitRoutePlan(ctx context.Context, req *mapsv1.TransitRoutePlanReq) (*mapsv1.TransitRoutePlanResp, error)
	WalkingDistanceMatrixPlan(ctx context.Context, req *mapsv1.WalkingDistanceMatrixPlanReq) (*mapsv1.WalkingDistanceMatrixPlanResp, error)
	BicyclingDistanceMatrixPlan(ctx context.Context, req *mapsv1.BicyclingDistanceMatrixPlanReq) (*mapsv1.BicyclingDistanceMatrixPlanResp, error)
	DrivingDistanceMatrixPlan(ctx context.Context, req *mapsv1.DrivingDistanceMatrixPlanReq) (*mapsv1.DrivingDistanceMatrixPlanResp, error)
	GetLocationByIp(ctx context.Context, req *mapsv1.GetLocationByIpReq) (*mapsv1.GetLocationByIpResp, error)
	DistanceMeasure(ctx context.Context, req *mapsv1.DistanceMeasureReq) (*mapsv1.DistanceMeasureResp, error)
	GetTimezoneByLocation(ctx context.Context, req *mapsv1.GetTimezoneByLocationReq) (*mapsv1.GetTimezoneByLocationResp, error)
}

type MapService struct {
	services map[enumv1.MapSource]IMapService
	db       *query.Query
	manager  *MapManager
	mapsv1.UnimplementedMapServiceServer
}

func (m *MapService) GetDistricts(ctx context.Context, req *mapsv1.GetDistrictsReq) (*mapsv1.GetDistrictsResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_DISTRICTS) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].GetDistricts(ctx, req)
}

func (m *MapService) GetDistrictChildren(ctx context.Context, req *mapsv1.GetDistrictChildrenReq) (*mapsv1.GetDistrictChildrenResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_DISTRICTS_CHILDREN) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].GetDistrictChildren(ctx, req)
}

func (m *MapService) SearchDistrict(ctx context.Context, req *mapsv1.SearchDistrictReq) (*mapsv1.SearchDistrictResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_DISTRICTS_SEARCH) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].SearchDistrict(ctx, req)
}

func (m *MapService) ConvertLocation(ctx context.Context, req *mapsv1.ConvertLocationReq) (*mapsv1.ConvertLocationResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_LOCATION_CONVERT) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].ConvertLocation(ctx, req)
}

func (m *MapService) ParseLocationToAddr(ctx context.Context, req *mapsv1.ParseLocationToAddrReq) (*mapsv1.ParseLocationToAddrResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_LOCATION_TO_ADDR) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].ParseLocationToAddr(ctx, req)
}

func (m *MapService) ParseAddrToLocation(ctx context.Context, req *mapsv1.ParseAddrToLocationReq) (*mapsv1.ParseAddrToLocationResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_ADDR_TO_LOCATION) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].ParseAddrToLocation(ctx, req)
}

func (m *MapService) WalkingRoutePlan(ctx context.Context, req *mapsv1.WalkingRoutePlanReq) (*mapsv1.WalkingRoutePlanResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_WALKING) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].WalkingRoutePlan(ctx, req)
}

func (m *MapService) BicyclingRoutePlan(ctx context.Context, req *mapsv1.BicyclingRoutePlanReq) (*mapsv1.BicyclingRoutePlanResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_BICYCLING) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].BicyclingRoutePlan(ctx, req)
}

func (m *MapService) EBicyclingRoutePlan(ctx context.Context, req *mapsv1.EBicyclingRoutePlanReq) (*mapsv1.EBicyclingRoutePlanResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_E_BICYCLING) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].EBicyclingRoutePlan(ctx, req)
}

func (m *MapService) DrivingRoutePlan(ctx context.Context, req *mapsv1.DrivingRoutePlanReq) (*mapsv1.DrivingRoutePlanResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_DRIVING) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].DrivingRoutePlan(ctx, req)
}

func (m *MapService) TransitRoutePlan(ctx context.Context, req *mapsv1.TransitRoutePlanReq) (*mapsv1.TransitRoutePlanResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_TRANSIT) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].TransitRoutePlan(ctx, req)
}

func (m *MapService) WalkingDistanceMatrixPlan(ctx context.Context, req *mapsv1.WalkingDistanceMatrixPlanReq) (*mapsv1.WalkingDistanceMatrixPlanResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_WALKING) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].WalkingDistanceMatrixPlan(ctx, req)
}

func (m *MapService) BicyclingDistanceMatrixPlan(ctx context.Context, req *mapsv1.BicyclingDistanceMatrixPlanReq) (*mapsv1.BicyclingDistanceMatrixPlanResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_BICYCLING) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].BicyclingDistanceMatrixPlan(ctx, req)
}

func (m *MapService) DrivingDistanceMatrixPlan(ctx context.Context, req *mapsv1.DrivingDistanceMatrixPlanReq) (*mapsv1.DrivingDistanceMatrixPlanResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_DRIVING) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].DrivingDistanceMatrixPlan(ctx, req)
}

func (m *MapService) GetLocationByIp(ctx context.Context, req *mapsv1.GetLocationByIpReq) (*mapsv1.GetLocationByIpResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_GET_LOCATION_BY_IP) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].GetLocationByIp(ctx, req)
}

func (m *MapService) GetTimezoneByLocation(ctx context.Context, req *mapsv1.GetTimezoneByLocationReq) (*mapsv1.GetTimezoneByLocationResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_GET_TIMEZONE_BY_LOCATION) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].GetTimezoneByLocation(ctx, req)
}

func (m *MapService) DistanceMeasure(ctx context.Context, req *mapsv1.DistanceMeasureReq) (*mapsv1.DistanceMeasureResp, error) {
	if !hasInterface(req.Source, enumv1.MapInterfaceType_MAP_INTERFACE_TYPE_DISTANCE_MEASURE) {
		return nil, ecode.ErrMapsInterfaceNotSupported
	}
	return m.services[req.Source].DistanceMeasure(ctx, req)
}

func (m *MapService) GetAvailableInterfaces(ctx context.Context, req *mapsv1.GetAvailableInterfacesReq) (*mapsv1.GetAvailableInterfacesResp, error) {
	interfaces, err := getAvailableInterfaces(req.Source)
	if err != nil {
		return nil, err
	}
	return &mapsv1.GetAvailableInterfacesResp{Types: interfaces}, nil
}

func (m *MapService) AddMap(ctx context.Context, req *mapsv1.AddMapReq) (*mapsv1.AddMapResp, error) {
	_, ok, err := m.checkMapNameExists(ctx, req.ObjectId, req.Name)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, ecode.ErrMapsAlreadyExists
	}
	if err := m.db.Transaction(func(tx *query.Query) error {
		accountM := &model.MapAccount{
			Name:      req.Name,
			MapSource: int16(req.Source),
			MapType:   int16(req.Type),
			Key:       req.Key,
			Status:    int16(enumv1.MapStatus_MAP_STATUS_OK),
			OwnerType: int16(req.OwnerType),
			ObjectID:  req.ObjectId,
			Comment:   req.Comment,
		}
		if err := tx.MapAccount.WithContext(ctx).Create(accountM); err != nil {
			return err
		}
		var interfaces []*model.MapInterface
		for _, i := range req.Interfaces {
			if !hasInterface(req.Source, i.InterfaceType) {
				return ecode.ErrMapsInterfaceNotSupported
			}
			interfaces = append(interfaces, &model.MapInterface{
				MapID:         accountM.ID,
				InterfaceType: int32(i.InterfaceType),
				SecondLimit:   trans.Ref(i.SecondLimit),
				DailyLimit:    trans.Ref(i.DailyLimit),
			})
		}
		if len(req.Interfaces) > 0 {
			if err := tx.MapInterface.WithContext(ctx).Create(interfaces...); err != nil {
				return err
			}
			for _, i := range req.Interfaces {
				if err := m.manager.AddMapInterface(ctx, req.Source, i.InterfaceType, accountM.ID); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &mapsv1.AddMapResp{}, nil
}

func (m *MapService) DeleteMap(ctx context.Context, req *mapsv1.DeleteMapReq) (*mapsv1.DeleteMapResp, error) {
	account, exists, err := m.checkMapExists(ctx, req.MapId)
	if err != nil || !exists {
		return nil, err
	}
	if err := m.db.Transaction(func(tx *query.Query) error {
		is, err := tx.MapInterface.WithContext(ctx).Where(tx.MapInterface.MapID.Eq(req.MapId)).Find()
		if err != nil {
			return err
		}
		if _, err := tx.MapAccount.WithContext(ctx).Where(tx.MapAccount.ID.Eq(req.MapId)).Delete(); err != nil {
			return err
		}
		if _, err := tx.MapInterface.WithContext(ctx).Where(tx.MapInterface.MapID.Eq(req.MapId)).Delete(); err != nil {
			return err
		}
		if _, err := tx.MapInterfaceCount.WithContext(ctx).Where(tx.MapInterfaceCount.MapID.Eq(req.MapId)).Delete(); err != nil {
			return err
		}
		for _, i := range is {
			if err := m.manager.RemoveMapInterface(ctx, enumv1.MapSource(account.MapSource), enumv1.MapInterfaceType(i.InterfaceType), i.MapID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &mapsv1.DeleteMapResp{}, nil
}

func (m *MapService) UpdateMap(ctx context.Context, req *mapsv1.UpdateMapReq) (*mapsv1.UpdateMapResp, error) {
	_, exists, err := m.checkMapExists(ctx, req.MapId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ecode.ErrMapsMapIDNotFound
	}
	account := &model.MapAccount{
		ID:        req.MapId,
		Name:      trans.Deref(req.Name),
		MapSource: int16(trans.Deref(req.Source)),
		MapType:   int16(trans.Deref(req.MapType)),
		Key:       trans.Deref(req.Key),
		Status:    int16(trans.Deref(req.Status)),
		OwnerType: int16(trans.Deref(req.OwnerType)),
		Comment:   req.Comment,
	}
	if err := m.db.Transaction(func(tx *query.Query) error {
		q := tx.MapAccount
		if _, err = q.WithContext(ctx).Where(q.ID.Eq(req.MapId)).Updates(account); err != nil {
			return err
		}
		is, err := tx.MapInterface.WithContext(ctx).Where(tx.MapInterface.MapID.Eq(req.MapId)).Find()
		if err != nil {
			return err
		}
		for _, i := range is {
			if err := m.manager.RemoveMapInterface(ctx, enumv1.MapSource(account.MapSource), enumv1.MapInterfaceType(i.InterfaceType), i.MapID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &mapsv1.UpdateMapResp{}, nil
}

func (m *MapService) GetMapInfoById(ctx context.Context, req *mapsv1.GetMapInfoByIdReq) (*mapsv1.GetMapInfoByIdResp, error) {
	info, err := m.getMapInfoByID(ctx, req.MapId)
	if err != nil {
		return nil, err
	}
	return &mapsv1.GetMapInfoByIdResp{MapInfo: info}, nil
}

func (m *MapService) PagingGetMapInfo(ctx context.Context, req *mapsv1.PagingGetMapInfoReq) (*mapsv1.PagingGetMapInfoResp, error) {
	q := m.db.MapAccount
	tx := q.WithContext(ctx)
	if req.Asc {
		tx = tx.Order(q.ID.Asc())
	} else {
		tx = tx.Order(q.ID.Desc())
	}
	if req.MapId != nil {
		tx = tx.Where(q.ID.Eq(req.GetMapId()))
	}
	if req.MapName != nil {
		tx = tx.Where(q.Name.Eq(req.GetMapName()))
	}
	if req.Source != nil {
		tx = tx.Where(q.MapSource.Eq(int16(req.GetSource())))
	}
	if req.MapType != nil {
		tx = tx.Where(q.MapType.Eq(int16(req.GetMapType())))
	}
	if req.OwnerType != nil {
		tx = tx.Where(q.OwnerType.Eq(int16(req.GetOwnerType())))
	}
	if req.Status != nil {
		tx = tx.Where(q.Status.Eq(int16(req.GetStatus())))
	}
	if req.ObjectId != nil {
		tx = tx.Where(q.ObjectID.Eq(req.GetObjectId()))
	}
	if req.CreatedStart != nil && req.CreatedEnd != nil {
		tx = tx.Where(q.CreatedAt.Between(req.CreatedStart.AsTime(), req.CreatedEnd.AsTime()))
	}
	result, err := db.Paginate[model.MapAccount](tx.UnderlyingDB(), uint32(req.Page), uint32(req.Size))
	if err != nil {
		return nil, err
	}
	var mapInfos = make([]*mapsv1.MapInfo, 0, len(result.List))
	for _, item := range result.List {
		mapInfos = append(mapInfos, accountModelToMapInfo(item))
	}
	resp := &mapsv1.PagingGetMapInfoResp{
		Page:       int32(result.Page),
		Size:       int32(result.PageSize),
		Total:      int64(result.Total),
		TotalPages: int32(result.TotalPages),
		MapInfos:   mapInfos,
	}
	return resp, nil
}

func (m *MapService) AddMapInterfaces(ctx context.Context, req *mapsv1.AddMapInterfacesReq) (*mapsv1.AddMapInterfacesResp, error) {
	account, exists, err := m.checkMapExists(ctx, req.MapId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ecode.ErrMapsMapIDNotFound
	}
	l := len(req.Interfaces)
	var iTypes = make([]enumv1.MapInterfaceType, 0, l)
	var interfaces = make([]*model.MapInterface, 0, l)
	for _, i := range req.Interfaces {
		iTypes = append(iTypes, i.InterfaceType)
		interfaces = append(interfaces, &model.MapInterface{
			MapID:         req.MapId,
			InterfaceType: int32(i.InterfaceType),
			SecondLimit:   trans.Ref(i.SecondLimit),
			DailyLimit:    trans.Ref(i.DailyLimit),
		})
	}
	exists, err = m.checkMapInterfaceExists(ctx, req.MapId, iTypes)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ecode.ErrMapsInterfaceAlreadyExists
	}

	if err := m.db.Transaction(func(tx *query.Query) error {
		q := tx.MapInterface
		if err := q.WithContext(ctx).Where(q.ID.Eq(req.MapId)).Create(interfaces...); err != nil {
			return err
		}
		for _, i := range req.Interfaces {
			if err := m.manager.AddMapInterface(ctx, enumv1.MapSource(account.MapSource), i.InterfaceType, account.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	var ids = make([]int64, 0, l)
	for _, i := range interfaces {
		ids = append(ids, i.ID)

	}
	return &mapsv1.AddMapInterfacesResp{Ids: ids}, nil
}

func (m *MapService) DeleteMapInterface(ctx context.Context, req *mapsv1.DeleteMapInterfaceReq) (*mapsv1.DeleteMapInterfaceResp, error) {
	account, exists, err := m.checkMapExists(ctx, req.MapId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ecode.ErrMapsMapIDNotFound
	}
	if err := m.db.Transaction(func(tx *query.Query) error {
		q := tx.MapInterface
		d := q.WithContext(ctx).Where(q.ID.Eq(req.MapId))
		if req.Type != nil {
			d = d.Where(q.InterfaceType.Eq(int32(req.GetType())))
		}
		if _, err := d.Delete(); err != nil {
			return err
		}
		return m.manager.RemoveMapInterface(ctx, enumv1.MapSource(account.MapSource), req.GetType(), req.MapId)
	}); err != nil {
		return nil, err
	}
	return &mapsv1.DeleteMapInterfaceResp{}, nil
}

func (m *MapService) UpdateMapInterface(ctx context.Context, req *mapsv1.UpdateMapInterfaceReq) (*mapsv1.UpdateMapInterfaceResp, error) {
	account, exists, err := m.checkMapExists(ctx, req.MapId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ecode.ErrMapsMapIDNotFound
	}
	exists, err = m.checkMapInterfaceExists(ctx, req.MapId, []enumv1.MapInterfaceType{req.GetType()})
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ecode.ErrMapsInterfaceNotFound
	}
	interfaceM := model.MapInterface{
		MapID:       req.MapId,
		SecondLimit: req.SecondLimit,
		DailyLimit:  req.DailyLimit,
	}
	if err := m.db.Transaction(func(tx *query.Query) error {
		q := tx.MapInterface
		if _, err := q.WithContext(ctx).Where(q.MapID.Eq(req.MapId), q.InterfaceType.Eq(int32(req.GetType()))).Updates(interfaceM); err != nil {
			return err
		}
		return m.manager.DeleteMapInterfaceCache(ctx, enumv1.MapSource(account.MapSource), req.GetType(), req.MapId)
	}); err != nil {
		return nil, err
	}

	return &mapsv1.UpdateMapInterfaceResp{}, nil
}

func (m *MapService) GetMapInterfaces(ctx context.Context, req *mapsv1.GetMapInterfacesReq) (*mapsv1.GetMapInterfacesResp, error) {
	infos, err := m.getMapInterfacesByMapID(ctx, req.MapId)
	if err != nil {
		return nil, err
	}
	return &mapsv1.GetMapInterfacesResp{InterfaceInfo: infos}, nil
}

func (m *MapService) PagingGetMapInterfaces(ctx context.Context, req *mapsv1.PagingGetMapInterfacesReq) (*mapsv1.PagingGetMapInterfacesResp, error) {
	q := m.db.MapInterface
	tx := q.WithContext(ctx)
	if req.Asc {
		tx = tx.Order(q.ID.Asc())
	} else {
		tx = tx.Order(q.ID.Desc())
	}
	if req.MapId != nil {
		tx = tx.Where(q.MapID.Eq(req.GetMapId()))
	}
	if req.Type != nil {
		tx = tx.Where(q.InterfaceType.Eq(int32(req.GetType())))
	}
	if req.CreatedStart != nil && req.CreatedEnd != nil {
		tx = tx.Where(q.CreatedAt.Between(req.CreatedStart.AsTime(), req.CreatedEnd.AsTime()))
	}
	result, err := db.Paginate[model.MapInterface](tx.UnderlyingDB(), uint32(req.Page), uint32(req.Size))
	if err != nil {
		return nil, err
	}
	resp := &mapsv1.PagingGetMapInterfacesResp{
		Page:       int32(result.Page),
		Size:       int32(result.PageSize),
		Total:      int64(result.Total),
		TotalPages: int32(result.TotalPages),
	}
	l := len(result.List)
	if l == 0 {
		return resp, nil
	}
	var ids = make([]int64, 0, l)
	for _, i := range result.List {
		ids = append(ids, i.MapID)
	}
	ids = slicex.Unique(ids)
	accountQ := m.db.MapAccount
	models, err := accountQ.WithContext(ctx).Where(accountQ.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	if len(models) != len(ids) {
		return nil, ecode.ErrInternal
	}
	accountMapping := make(map[int64]*model.MapAccount, len(ids))
	for _, m := range models {
		accountMapping[m.ID] = m
	}
	var interfaces = make([]*mapsv1.MapInterfaceInfo, 0, l)
	for _, v := range result.List {
		interfaces = append(interfaces, interfaceModelToInfo(v, accountMapping[v.MapID]))
	}
	resp.Interfaces = interfaces
	return resp, nil
}

func (m *MapService) DeleteMapInterfaceCounts(ctx context.Context, req *mapsv1.DeleteMapInterfaceCountsReq) (*mapsv1.DeleteMapInterfaceCountsResp, error) {
	q := m.db.MapInterfaceCount
	affected, err := q.WithContext(ctx).Where(q.CreatedAt.Lte(req.Before.AsTime())).Delete()
	if err != nil {
		return nil, err
	}
	return &mapsv1.DeleteMapInterfaceCountsResp{AffectedRows: affected.RowsAffected}, nil
}

func (m *MapService) GetMapInterfaceCounts(ctx context.Context, req *mapsv1.GetMapInterfaceCountsReq) (*mapsv1.GetMapInterfaceCountsResp, error) {
	if req.Day == nil {
		return nil, ecode.ErrParams
	}
	day := timex.StartOfDay(req.Day.AsTime())
	q := m.db.MapInterfaceCount
	tx := q.WithContext(ctx).Where(q.ID.Eq(req.MapId), q.Day.Eq(day))
	if req.Type != nil {
		tx = tx.Where(q.InterfaceType.Eq(int32(req.GetType())))
	}
	models, err := tx.Find()
	if err != nil {
		return nil, err
	}
	counts, err := m.convertToMapInterfaceCounts(ctx, models)
	if err != nil {
		return nil, err
	}
	return &mapsv1.GetMapInterfaceCountsResp{Counts: counts}, nil
}

func (m *MapService) PagingGetMapInterfaceCountInfos(ctx context.Context, req *mapsv1.PagingGetMapInterfaceCountInfosReq) (*mapsv1.PagingGetMapInterfaceCountInfosResp, error) {
	q := m.db.MapInterfaceCount
	tx := q.WithContext(ctx)
	if req.Asc {
		tx = tx.Order(q.ID.Asc())
	} else {
		tx = tx.Order(q.ID.Desc())
	}
	if req.MapId != nil {
		tx = tx.Where(q.MapID.Eq(req.GetMapId()))
	}
	if req.Type != nil {
		tx = tx.Where(q.InterfaceType.Eq(int32(req.GetType())))
	}
	if req.DayStart != nil && req.DayEnd != nil {
		tx = tx.Where(q.Day.Between(req.DayStart.AsTime(), req.DayEnd.AsTime()))
	}
	results, err := db.Paginate[model.MapInterfaceCount](tx.UnderlyingDB(), uint32(req.Page), uint32(req.Size))
	if err != nil {
		return nil, err
	}
	resp := &mapsv1.PagingGetMapInterfaceCountInfosResp{
		Page:       int32(results.Page),
		Size:       int32(results.PageSize),
		Total:      int64(results.Total),
		TotalPages: int32(results.TotalPages),
	}
	if len(results.List) == 0 {
		return resp, nil
	}
	counts, err := m.convertToMapInterfaceCounts(ctx, results.List)
	if err != nil {
		return nil, err
	}
	resp.Counts = counts
	return resp, nil
}

func (m *MapService) getMapInfoByID(ctx context.Context, mapID int64) (*mapsv1.MapInfo, error) {
	q := m.db.MapAccount
	accountM, err := q.WithContext(ctx).Where(q.ID.Eq(mapID)).First()
	if err != nil {
		if isDBNotFound(err) {
			return nil, ecode.ErrMapsMapIDNotFound
		}
		return nil, err
	}
	mapInfo := accountModelToMapInfo(accountM)
	return mapInfo, nil
}

func (m *MapService) getMapInterfacesByMapID(ctx context.Context, mapID int64) ([]*mapsv1.MapInterfaceInfo, error) {
	q := m.db.MapInterface
	models, err := q.WithContext(ctx).Where(q.MapID.Eq(mapID)).Find()
	if err != nil {
		return nil, err
	}
	if len(models) == 0 {
		return nil, ecode.ErrMapsNoInterface
	}
	accountQ := m.db.MapAccount
	accountM, err := accountQ.WithContext(ctx).Where(q.MapID.Eq(mapID)).Take()
	if err != nil {
		if isDBNotFound(err) {
			return nil, ecode.ErrMapsMapIDNotFound
		}
		return nil, err
	}
	var infos = make([]*mapsv1.MapInterfaceInfo, 0, len(models))
	for _, v := range models {
		infos = append(infos, interfaceModelToInfo(v, accountM))
	}
	return infos, nil
}

func (m *MapService) convertToMapInterfaceCounts(ctx context.Context, models []*model.MapInterfaceCount) ([]*mapsv1.MapInterfaceCount, error) {
	l := len(models)
	if l == 0 {
		return nil, nil
	}
	var counts = make([]*mapsv1.MapInterfaceCount, 0, l)
	var ids = make([]int64, 0, l)
	for _, v := range models {
		ids = append(ids, v.MapID)
	}
	ids = slicex.Unique(ids)
	accountQ := m.db.MapAccount
	accountModels, err := accountQ.WithContext(ctx).Where(accountQ.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	idsLen := len(ids)
	if len(accountModels) != idsLen {
		return nil, ecode.ErrInternal
	}
	var accountMapping = make(map[int64]*model.MapAccount, idsLen)
	for _, v := range accountModels {
		accountMapping[v.ID] = v
	}
	for _, v := range models {
		accountM := accountMapping[v.MapID]
		count := &mapsv1.MapInterfaceCount{
			Id:            v.ID,
			MapId:         v.MapID,
			Source:        enumv1.MapSource(accountM.MapSource),
			MapType:       enumv1.MapType(accountM.MapType),
			InterfaceType: enumv1.MapInterfaceType(v.InterfaceType),
			OwnerType:     enumv1.MapOwnerType(accountM.OwnerType),
			ObjectId:      accountM.ObjectID,
			Count:         v.Count,
			ErrCount:      v.ErrCount,
			Day:           timestamppb.New(*v.Day),
			CreatedAt:     timestamppb.New(*v.CreatedAt),
			UpdatedAt:     timestamppb.New(*v.UpdatedAt),
			DeletedAt:     nil,
		}
		if v.DeletedAt.Valid {
			count.DeletedAt = timestamppb.New(v.DeletedAt.Time)
		}
		counts = append(counts, count)
	}
	return counts, nil
}

func (m *MapService) checkMapNameExists(ctx context.Context, objectId int64, mapName string) (mapID int64, exists bool, err error) {
	q := m.db.MapAccount
	account, err := q.WithContext(ctx).Where(q.ObjectID.Eq(objectId), q.Name.Eq(mapName)).Select(q.ID).First()
	if err != nil {
		if isDBNotFound(err) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return account.ID, true, nil
}

func (m *MapService) checkMapExists(ctx context.Context, mapID int64) (account *model.MapAccount, exists bool, err error) {
	q := m.db.MapAccount
	account, err = q.WithContext(ctx).Where(q.ID.Eq(mapID)).First()
	if err != nil {
		if isDBNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return account, true, nil
}

func (m *MapService) checkMapInterfaceExists(ctx context.Context, mapID int64, iTypes []enumv1.MapInterfaceType) (exists bool, err error) {
	l := len(iTypes)
	if l == 0 {
		return false, nil
	}
	var ts []int32 = make([]int32, 0, l)
	for _, t := range iTypes {
		ts = append(ts, int32(t))
	}
	q := m.db.MapInterface
	_, err = q.WithContext(ctx).Where(q.MapID.Eq(mapID), q.InterfaceType.In(ts...)).Select(q.ID).First()
	if err != nil {
		if isDBNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
