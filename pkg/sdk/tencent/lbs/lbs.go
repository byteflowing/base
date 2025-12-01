package lbs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/httpx"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
)

const (
	maxWaypoints           = 30
	maxAvoidPolygons       = 32
	maxPolygonPoints       = 16
	urlFormat              = "%s?%s"
	locationFormat         = "%.6f,%.6f"
	trackFormat            = "%.6f,%.6f,%f,%d,%f,%f,%d"
	errFormat              = "status: %d, grpc_message: %s, request_id: %s"
	statusOK               = 0
	statusSecondLimit      = 120
	statusDayLimit         = 121
	getDistrictsURL        = "https://apis.map.qq.com/ws/district/v1/list"
	getDistrictChildrenURL = "https://apis.map.qq.com/ws/district/v1/getchildren"
	searchDistrictURL      = "https://apis.map.qq.com/ws/district/v1/search"
	locationConvertURL     = "https://apis.map.qq.com/ws/coord/v1/translate"
	geocoderURL            = "https://apis.map.qq.com/ws/geocoder/v1/"
	drivingPlanURL         = "https://apis.map.qq.com/ws/direction/v1/driving/"
	walkingPlanURL         = "https://apis.map.qq.com/ws/direction/v1/walking/"
	bicyclingPlanURL       = "https://apis.map.qq.com/ws/direction/v1/bicycling/"
	eBicyclingPlanURL      = "https://apis.map.qq.com/ws/direction/v1/ebicycling/"
	transitPlanURL         = "https://apis.map.qq.com/ws/direction/v1/transit/"
	distanceMatrixURL      = "https://apis.map.qq.com/ws/distance/v1/matrix"
	getLocationByIpURL     = "https://apis.map.qq.com/ws/location/v1/ip"
)

type MapService struct {
	httpClient *http.Client
}

func NewMapService() *MapService {
	return &MapService{
		httpClient: httpx.NewClient(httpx.GetDefaultConfig()),
	}
}

func (m *MapService) GetDistricts(ctx context.Context, req *mapsv1.TencentGetDistrictsReq) (resp *mapsv1.TencentDistrictResp, err error) {
	params := url.Values{}
	params.Add("key", req.Key)
	resp = &mapsv1.TencentDistrictResp{}
	if err := m.requestProto(ctx, getDistrictsURL, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Message, resp.RequestId); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) GetDistrictChildren(ctx context.Context, req *mapsv1.TencentGetDistrictChildrenReq) (resp *mapsv1.TencentDistrictResp, err error) {
	params := url.Values{}
	params.Add("key", req.Key)
	if req.GetGetPolygon() {
		params.Add("polygon", "1")
	}
	resp = &mapsv1.TencentDistrictResp{}
	if err := m.requestProto(ctx, getDistrictChildrenURL, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Message, resp.RequestId); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) SearchDistrict(ctx context.Context, req *mapsv1.TencentSearchDistrictReq) (resp *mapsv1.TencentDistrictResp, err error) {
	params := url.Values{}
	params.Add("key", req.Key)
	if req.GetGetPolygon() {
		params.Add("polygon", "1")
	}
	resp = &mapsv1.TencentDistrictResp{}
	if err := m.requestProto(ctx, searchDistrictURL, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Message, resp.RequestId); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) LocationConvert(ctx context.Context, req *mapsv1.TencentLocationConvertReq) (resp *mapsv1.TencentLocationConvertResp, err error) {
	var sb strings.Builder
	for idx, l := range req.Locations {
		if idx > 0 {
			sb.WriteString(";")
		}
		_, _ = fmt.Fprintf(&sb, locationFormat, l.Lat, l.Lng)
	}
	params := url.Values{}
	params.Add("key", req.Key)
	params.Add("locations", sb.String())
	params.Add("type", strconv.Itoa(int(req.Type)))
	resp = &mapsv1.TencentLocationConvertResp{}
	if err := m.requestProto(ctx, locationConvertURL, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Message, resp.RequestId); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) LocationToAddr(ctx context.Context, req *mapsv1.TencentLocationToAddrReq) (resp *mapsv1.TencentLocationToAddrResp, err error) {
	params := url.Values{}
	params.Add("key", req.Key)
	params.Add("location", fmt.Sprintf(locationFormat, req.Location.Lat, req.Location.Lng))
	if req.Radius != nil {
		params.Add("radius", strconv.Itoa(int(*req.Radius)))
	}
	if req.GetPoi != nil && *req.GetPoi {
		params.Add("get_poid", "1")
	}
	var poiOptions []string
	if req.PoiOptionAddressFormatShort != nil {
		poiOptions = append(poiOptions, "address_format=short")
	}
	if req.PoiOptionRadius != nil {
		poiOptions = append(poiOptions, fmt.Sprintf("radius=%d", *req.PoiOptionRadius))
	}
	if req.PoiOptionPolicy != nil {
		poiOptions = append(poiOptions, fmt.Sprintf("policy=%d", *req.PoiOptionPolicy))
	}
	if req.GetPoiOptionOrderByDistance() {
		poiOptions = append(poiOptions, "order_by=_distance")
	}
	if len(poiOptions) > 0 {
		params.Add("poi_options", strings.Join(poiOptions, ";"))
	}
	resp = &mapsv1.TencentLocationToAddrResp{}
	if err := m.requestProto(ctx, geocoderURL, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Message, resp.RequestId); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) AddressToLocation(ctx context.Context, req *mapsv1.TencentAddressToLocationReq) (resp *mapsv1.TencentAddressToLocationResp, err error) {
	params := url.Values{}
	params.Add("key", req.Key)
	params.Add("address", req.Address)
	resp = &mapsv1.TencentAddressToLocationResp{}
	if err := m.requestProto(ctx, geocoderURL, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Message, resp.RequestId); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) RoutePlan(ctx context.Context, req *mapsv1.TencentRoutePlanReq) (resp *mapsv1.TencentRoutePlanResp, err error) {
	params := url.Values{}
	params.Add("key", req.Key)
	params.Add("from", fmt.Sprintf(locationFormat, req.From.Lat, req.From.Lng))
	params.Add("to", fmt.Sprintf(locationFormat, req.To.Lat, req.To.Lng))

	// 途经点
	waypoints := len(req.Waypoints)
	if waypoints > 0 {
		if waypoints > maxWaypoints {
			return nil, errors.New("no more than 30 waypoints")
		}
		waypoints := make([]string, 0, waypoints)
		for _, wp := range req.Waypoints {
			waypoints = append(waypoints, fmt.Sprintf(locationFormat, wp.Lat, wp.Lng))
		}
		params.Add("waypoints", strings.Join(waypoints, ";"))
		if req.WaypointOrder != nil && *req.WaypointOrder {
			params.Add("waypoint_order", "1")
		}
	}

	// 基础参数
	if req.FromPoi != nil {
		params.Add("from_poi", *req.FromPoi)
	}
	if req.ToPoi != nil {
		params.Add("to_poi", *req.ToPoi)
	}
	if req.ToPoiName != nil {
		params.Add("to_poiname", *req.ToPoiName)
	}
	if req.Heading != nil {
		params.Add("heading", strconv.Itoa(int(*req.Heading)))
	}
	if req.Speed != nil {
		params.Add("speed", strconv.FormatFloat(*req.Speed, 'f', -1, 64))
	}
	if req.Accuracy != nil {
		params.Add("accuracy", strconv.Itoa(int(*req.Accuracy)))
	}
	if req.RoadType != nil {
		params.Add("road_type", strconv.Itoa(int(*req.RoadType)))
	}

	// 行驶轨迹
	if len(req.Tracks) > 0 {
		tracks := make([]string, 0, len(req.Tracks))
		for _, tr := range req.Tracks {
			var (
				speed           float64 = -1
				accuracy        int64   = -1
				motionDirection float64 = -1
				deviceDirection float64 = -1
				timestamp       int64   = 0
			)
			if tr.Speed != nil {
				speed = *tr.Speed
			}
			if tr.Accuracy != nil {
				accuracy = *tr.Accuracy
			}
			if tr.MotionDirection != nil {
				motionDirection = *tr.MotionDirection
			}
			if tr.DeviceDirection != nil {
				deviceDirection = *tr.DeviceDirection
			}
			if tr.Timestamp != nil {
				timestamp = *tr.Timestamp
			}

			// 使用 strings.Builder 避免重复分配
			var sb strings.Builder
			_, _ = fmt.Fprintf(&sb, trackFormat, tr.Lat, tr.Lng, speed, accuracy, motionDirection, deviceDirection, timestamp)
			tracks = append(tracks, sb.String())
		}
		params.Add("from_track", strings.Join(tracks, ";"))
	}

	// 其他标志
	if req.GetWithDest() {
		params.Add("with_dest", "1")
	}
	if req.DepartureTime != nil {
		params.Add("departure_time", strconv.FormatInt(*req.DepartureTime, 10))
	}
	if req.PlateNumber != nil {
		params.Add("plate_number", *req.PlateNumber)
	}

	// 策略参数
	policies := make([]string, 0, 2)
	if req.PolicyParam != nil {
		switch *req.PolicyParam {
		case enumsv1.TencentDrivePlanPolicyParam_TENCENT_DRIVE_PLAN_POLICY_PARAM_LEASE_TIME:
			policies = append(policies, "LEASE_TIME")
		case enumsv1.TencentDrivePlanPolicyParam_TENCENT_DRIVE_PLAN_POLICY_PARAM_PICKUP:
			policies = append(policies, "PICKUP")
		case enumsv1.TencentDrivePlanPolicyParam_TENCENT_DRIVE_PLAN_POLICY_PARAM_TRIP:
			policies = append(policies, "TRIP")
		case enumsv1.TencentDrivePlanPolicyParam_TENCENT_DRIVE_PLAN_POLICY_PARAM_SHORT_DISTANCE:
			policies = append(policies, "SHORT_DISTANCE")
		}
	}
	if req.Policy != nil {
		switch *req.Policy {
		case enumsv1.TencentDrivingPlanPolicy_TENCENT_DRIVING_PLAN_POLICY_REAL_TRAFFIC:
			policies = append(policies, "REAL_TRAFFIC")
		case enumsv1.TencentDrivingPlanPolicy_TENCENT_DRIVING_PLAN_POLICY_LEAST_FEE:
			policies = append(policies, "LEAST_FEE")
		case enumsv1.TencentDrivingPlanPolicy_TENCENT_DRIVING_PLAN_POLICY_HIGHWAY_FIRST:
			policies = append(policies, "HIGHWAY_FIRST")
		case enumsv1.TencentDrivingPlanPolicy_TENCENT_DRIVING_PLAN_POLICY_AVOID_HIGHWAY:
			policies = append(policies, "AVOID_HIGHWAY")
		case enumsv1.TencentDrivingPlanPolicy_TENCENT_DRIVING_PLAN_POLICY_HIGHROAD_FIRST:
			policies = append(policies, "HIGHROAD_FIRST")
		case enumsv1.TencentDrivingPlanPolicy_TENCENT_DRIVING_PLAN_POLICY_NAV_POINT_FIRST:
			policies = append(policies, "NAV_POINT_FIRST")
		case enumsv1.TencentDrivingPlanPolicy_TENCENT_DRIVING_PLAN_POLICY_PICKUP_ADSORB:
			policies = append(policies, "PICKUP_ADSORB")
		case enumsv1.TencentDrivingPlanPolicy_TENCENT_DRIVING_PLAN_POLICY_TRIP_ADSORB:
			policies = append(policies, "TRIP_ADSORB")
		}
	}
	if len(policies) > 0 {
		params.Add("policy", strings.Join(policies, ","))
	}

	// 避让区域
	if len(req.AvoidPolygons) > 0 {
		if len(req.AvoidPolygons) > maxAvoidPolygons {
			return nil, fmt.Errorf("no more than %d avoid polygons", maxAvoidPolygons)
		}
		avoidPolygons := make([]string, 0, len(req.AvoidPolygons))
		for _, polygon := range req.AvoidPolygons {
			if l := len(polygon.Locations); l > 0 {
				if l > maxPolygonPoints {
					return nil, fmt.Errorf("no more than %d avoid points per polygon", maxPolygonPoints)
				}
				points := make([]string, 0, l)
				for _, loc := range polygon.Locations {
					points = append(points, fmt.Sprintf(locationFormat, loc.Lat, loc.Lng))
				}
				avoidPolygons = append(avoidPolygons, strings.Join(points, ";"))
			}
		}
		if len(avoidPolygons) > 0 {
			params.Add("avoid_polygons", strings.Join(avoidPolygons, "|"))
		}
	}

	// 其他标志字段
	if req.GetGetMp() {
		params.Add("get_mp", "1")
	}
	if req.GetGetSpeed_() {
		params.Add("get_speed", "1")
	}
	if len(req.AddedFields) > 0 {
		params.Add("added_fields", strings.Join(req.AddedFields, ","))
	}
	if req.GetNoStep() {
		params.Add("no_step", "1")
	}
	if req.ServiceLevel != nil {
		params.Add("service_level", strconv.Itoa(int(*req.ServiceLevel)))
	}

	// URL 拼接
	var addr string
	switch req.Mode {
	case enumsv1.PlanMode_PLAN_MODE_DRIVING:
		addr = drivingPlanURL
	case enumsv1.PlanMode_PLAN_MODE_WALKING:
		addr = walkingPlanURL
	case enumsv1.PlanMode_PLAN_MODE_BICYCLING:
		addr = bicyclingPlanURL
	case enumsv1.PlanMode_PLAN_MODE_E_BICYCLING:
		addr = eBicyclingPlanURL
	default:
		return nil, errors.New("invalid plan mode")
	}

	// 请求执行
	resp = &mapsv1.TencentRoutePlanResp{}
	if err := m.requestProto(ctx, addr, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Message, resp.RequestId); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) TransitPlan(ctx context.Context, req *mapsv1.TencentTransitPlanReq) (resp *mapsv1.TencentTransitPlanResp, err error) {
	params := url.Values{}
	params.Add("key", req.Key)
	params.Add("from", fmt.Sprintf(locationFormat, req.From.Lat, req.From.Lng))
	params.Add("to", fmt.Sprintf(locationFormat, req.To.Lat, req.To.Lng))
	if req.FromPoi != nil {
		params.Add("from_poi", *req.FromPoi)
	}
	if req.ToPoi != nil {
		params.Add("to_poi", *req.ToPoi)
	}
	if req.DepartureTime != nil {
		params.Add("departure_time", strconv.FormatInt(*req.DepartureTime, 10))
	}
	var policy []string
	if req.Policy != nil {
		switch *req.Policy {
		case enumsv1.TencentTransitPlanPolicy_TENCENT_TRANSIT_PLAN_POLICY_RECOMMEND:
			policy = append(policy, "RECOMMEND")
		case enumsv1.TencentTransitPlanPolicy_TENCENT_TRANSIT_PLAN_POLICY_LEAST_TIME:
			policy = append(policy, "LEAST_TIME")
		case enumsv1.TencentTransitPlanPolicy_TENCENT_TRANSIT_PLAN_POLICY_LEAST_TRANSFER:
			policy = append(policy, "LEAST_TRANSFER")
		case enumsv1.TencentTransitPlanPolicy_TENCENT_TRANSIT_PLAN_POLICY_LEAST_WALKING:
			policy = append(policy, "LEAST_WALKING")
		}
	}
	if req.Limit != nil {
		switch *req.Limit {
		case enumsv1.TencentTransitPlanLimit_TENCENT_TRANSIT_PLAN_LIMIT_NO_SUBWAY:
			policy = append(policy, "NO_SUBWAY")
		case enumsv1.TencentTransitPlanLimit_TENCENT_TRANSIT_PLAN_LIMIT_ONLY_SUBWAY:
			policy = append(policy, "ONLY_SUBWAY")
		case enumsv1.TencentTransitPlanLimit_TENCENT_TRANSIT_PLAN_LIMIT_SUBWAY_FIRST:
			policy = append(policy, "SUBWAY_FIRST")
		}
	}
	if len(policy) > 0 {
		params.Add("policy", strings.Join(policy, ","))
	}
	// 请求执行
	resp = &mapsv1.TencentTransitPlanResp{}
	if err := m.requestProto(ctx, transitPlanURL, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Message, resp.RequestId); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) DistanceMatrix(ctx context.Context, req *mapsv1.TencentDistanceMatrixReq) (resp *mapsv1.TencentDistanceMatrixResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	switch req.GetMode() {
	case enumsv1.PlanMode_PLAN_MODE_DRIVING:
		params.Add("mode", "driving")
	case enumsv1.PlanMode_PLAN_MODE_BICYCLING:
		params.Add("mode", "bicycling")
	case enumsv1.PlanMode_PLAN_MODE_WALKING:
		params.Add("mode", "walking")
	default:
		return nil, errors.New("invalid plan mode")
	}
	from := make([]string, 0, len(req.GetFrom()))
	for _, f := range req.From {
		from = append(from, fmt.Sprintf(locationFormat, f.Lat, f.Lng))
	}
	params.Add("from", strings.Join(from, ";"))
	to := make([]string, 0, len(req.GetTo()))
	for _, t := range req.GetTo() {
		to = append(to, fmt.Sprintf(locationFormat, t.Lat, t.Lng))
	}
	params.Add("to", strings.Join(to, ";"))
	resp = &mapsv1.TencentDistanceMatrixResp{}
	if err := m.requestProto(ctx, distanceMatrixURL, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Message, resp.RequestId); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) GetLocationByIpReq(ctx context.Context, req *mapsv1.TencentGetLocationByIpReq) (resp *mapsv1.TencentGetLocationByIpResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	if req.GetIp() != "" {
		params.Add("ip", req.GetIp())
	}
	resp = &mapsv1.TencentGetLocationByIpResp{}
	if err := m.requestProto(ctx, getLocationByIpURL, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Message, resp.RequestId); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) getParamsURL(addr string, params url.Values) string {
	return fmt.Sprintf(urlFormat, addr, params.Encode())
}

func (m *MapService) requestProto(ctx context.Context, addr string, params url.Values, msg proto.Message) (err error) {
	reqURL := m.getParamsURL(addr, params)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	response, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return protojson.Unmarshal(body, msg)
}

func (m *MapService) parseError(st int32, message, requestID string) (err error) {
	if st == statusOK {
		return nil
	}
	switch st {
	case statusSecondLimit:
		return ecode.ErrMapReachIntervalLimit
	case statusDayLimit:
		return ecode.ErrMapReachDailyLimit
	default:
		return fmt.Errorf(errFormat, st, message, requestID)
	}
}
