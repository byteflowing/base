package amap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/httpx"
	"github.com/byteflowing/base/pkg/jsonx"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	maxWaypoints           = 16
	maxAvoidPolygons       = 32
	maxPolygonPoints       = 16
	maxCoordLocations      = 40
	urlFormat              = "%s?%s"
	locationFormat         = "%.6f,%.6f"
	errFormat              = "status: %s, code: %s, grpc_message: %s"
	statusOK               = "1"
	statusDailyLimit       = "10003"
	statusIntervalLimit    = "10004"
	getLocationByIpURL     = "https://restapi.amap.com/v3/ip"
	distanceMeasureURL     = "https://restapi.amap.com/v3/distance"
	getDistrictsURL        = "https://restapi.amap.com/v3/config/district"
	locationConvertURL     = "https://restapi.amap.com/v3/assistant/coordinate/convert"
	locationToAddrURL      = "https://restapi.amap.com/v3/geocode/regeo"
	addrToLocationURL      = "https://restapi.amap.com/v3/geocode/geo"
	drivingRoutPlanURL     = "https://restapi.amap.com/v5/direction/driving"
	bicyclingRoutePlanURL  = "https://restapi.amap.com/v5/direction/bicycling"
	eBicyclingRoutePlanURL = "https://restapi.amap.com/v5/direction/electrobike"
	walkingRoutePlanURL    = "https://restapi.amap.com/v5/direction/walking"
	transitPlanURL         = "https://restapi.amap.com/v5/direction/transit/integrated"
)

type MapService struct {
	httpClient *http.Client
}

func NewMapService() *MapService {
	return &MapService{
		httpClient: httpx.NewClient(httpx.GetDefaultConfig()),
	}
}

func (m *MapService) GetDistricts(ctx context.Context, req *mapsv1.AmapGetDistrictsReq) (resp *mapsv1.AmapGetDistrictsResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	if req.KeyWord != nil {
		params.Add("keywords", req.GetKeyWord())
	}
	if req.Subdistrict != nil {
		params.Add("subdistrict", strconv.Itoa(int(req.GetSubdistrict())))
	}
	if req.Page != nil {
		params.Add("page", strconv.Itoa(int(req.GetPage())))
	}
	if req.Offset != nil {
		params.Add("offset", strconv.Itoa(int(*req.Offset)))
	}
	if req.Extensions != nil {
		params.Add("extensions", req.GetExtensions())
	}
	if req.Filter != nil {
		params.Add("filter", req.GetFilter())
	}
	resp = &mapsv1.AmapGetDistrictsResp{}
	if err := m.requestProto(ctx, getDistrictsURL, params, resp); err != nil {
		return nil, err
	}
	if err = m.parseError(resp.Status, resp.Info, resp.Infocode); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) LocationConvert(ctx context.Context, req *mapsv1.AmapLocationConvertReq) (resp *mapsv1.AmapLocationConvertResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	if len(req.GetLocations()) > maxCoordLocations {
		return nil, errors.New("locations no more than 40")
	}
	locations := make([]string, 0, len(req.GetLocations()))
	for _, location := range req.GetLocations() {
		locations = append(locations, fmt.Sprintf(locationFormat, location.Lng, location.Lat))
	}
	params.Add("locations", strings.Join(locations, "|"))
	var coord string
	switch req.Coordsys {
	case enumv1.AmapCoordType_AMAP_COORD_TYPE_GPS:
		coord = "gps"
	case enumv1.AmapCoordType_AMAP_COORD_TYPE_BAIDU:
		coord = "baidu"
	case enumv1.AmapCoordType_AMAP_COORD_TYPE_MAP_BAR:
		coord = "mapbar"
	default:
		return nil, errors.New("unsupported coordinate")
	}
	params.Add("coordinates", coord)
	resp = &mapsv1.AmapLocationConvertResp{}
	if err := m.requestProto(ctx, locationConvertURL, params, resp); err != nil {
		return nil, err
	}
	if err := m.parseError(resp.Status, resp.Info, resp.Infocode); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) LocationToAddr(ctx context.Context, req *mapsv1.AmapLocationToAddrReq) (resp *mapsv1.AmapLocationToAddrResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	params.Add("location", fmt.Sprintf(locationFormat, req.Location.Lng, req.Location.Lat))
	if req.Radius != nil {
		params.Add("radius", strconv.Itoa(int(*req.Radius)))
	}
	if req.GetExtensions() {
		params.Add("extensions", "all")
		if len(req.PoiType) > 0 {
			params.Add("poi_type", strings.Join(req.GetPoiType(), "|"))
		}
		if req.GetAllRoads() {
			params.Add("roadlevel", "0")
		} else {
			params.Add("roadlevel", "1")
		}
		if req.HomeOrCorp != nil {
			params.Add("homeorcorp", strconv.Itoa(int(*req.HomeOrCorp)))
		}
	}
	resp = &mapsv1.AmapLocationToAddrResp{}
	if err := m.requestProto(ctx, locationToAddrURL, params, resp); err != nil {
		return nil, err
	}
	if err := m.parseError(resp.Status, resp.Info, resp.Infocode); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) AddressToLocation(ctx context.Context, req *mapsv1.AmapAddressToLocationReq) (resp *mapsv1.AmapAddressToLocationResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	params.Add("address", req.GetAddress())
	if req.City != nil {
		params.Add("city", req.GetCity())
	}
	resp = &mapsv1.AmapAddressToLocationResp{}
	if err := m.requestProto(ctx, addrToLocationURL, params, resp); err != nil {
		return nil, err
	}
	if err := m.parseError(resp.Status, resp.Info, resp.Infocode); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) RoutePlan(ctx context.Context, req *mapsv1.AmapRoutePlanReq) (resp *mapsv1.AmapRoutePlanResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	params.Add("origin", fmt.Sprintf(locationFormat, req.Origin.Lng, req.Origin.Lat))
	params.Add("destination", fmt.Sprintf(locationFormat, req.Destination.Lng, req.Destination.Lat))

	// --- optional params ---
	if v := req.GetDestinationType(); v != "" {
		params.Add("destination_type", v)
	}
	if v := req.GetOriginId(); v != "" {
		params.Add("origin_id", v)
	}
	if v := req.GetDestinationId(); v != "" {
		params.Add("destination_id", v)
	}

	// --- driving strategy ---
	if req.Strategy != nil {
		params.Add("strategy", amapStrategyToCode(*req.Strategy))
	}

	// --- waypoints ---
	if n := len(req.Waypoints); n > 0 {
		if n > maxWaypoints {
			return nil, fmt.Errorf("no more than %d waypoints", maxWaypoints)
		}
		locs := make([]string, 0, n)
		for _, w := range req.Waypoints {
			locs = append(locs, fmt.Sprintf(locationFormat, w.Lng, w.Lat))
		}
		params.Add("waypoints", strings.Join(locs, ";"))
	}

	// --- avoid polygons ---
	if n := len(req.AvoidPolygons); n > 0 {
		if n > maxAvoidPolygons {
			return nil, fmt.Errorf("no more than %d avoid polygons", maxAvoidPolygons)
		}
		polygons := make([]string, 0, n)
		for _, p := range req.AvoidPolygons {
			if len(p.Locations) > maxPolygonPoints {
				return nil, fmt.Errorf("no more than %d points per polygon", maxPolygonPoints)
			}
			points := make([]string, 0, len(p.Locations))
			for _, loc := range p.Locations {
				points = append(points, fmt.Sprintf(locationFormat, loc.Lng, loc.Lat))
			}
			polygons = append(polygons, strings.Join(points, ";"))
		}
		params.Add("avoidpolygons", strings.Join(polygons, "|"))
	}

	// --- car type ---
	if req.CarType != nil {
		params.Add("cartype", amapCarTypeToCode(*req.CarType))
	}

	// --- misc fields ---
	if v := req.GetPlate(); v != "" {
		params.Add("plate", v)
	}
	if req.GetUseFerry() {
		params.Add("ferry", "0")
	}
	if len(req.ShowFields) > 0 {
		params.Add("show_fields", strings.Join(req.ShowFields, ","))
	}
	if req.AlternativeRoute != nil {
		params.Add("alternative_route", strconv.Itoa(int(*req.AlternativeRoute)))
	}
	if req.GetIsIndoor() {
		params.Add("indoor", "1")
	}

	// --- URL prefix by mode ---
	var addr string
	switch req.Mode {
	case enumv1.PlanMode_PLAN_MODE_DRIVING:
		addr = drivingRoutPlanURL
	case enumv1.PlanMode_PLAN_MODE_WALKING:
		addr = walkingRoutePlanURL
	case enumv1.PlanMode_PLAN_MODE_BICYCLING:
		addr = bicyclingRoutePlanURL
	case enumv1.PlanMode_PLAN_MODE_E_BICYCLING:
		addr = eBicyclingRoutePlanURL
	default:
		return nil, errors.New("unsupported plan mode")
	}

	// --- do request ---
	resp = &mapsv1.AmapRoutePlanResp{}
	if err := m.requestProto(ctx, addr, params, resp); err != nil {
		return nil, err
	}

	// --- parse error ---
	if err := m.parseError(resp.Status, resp.Info, resp.Infocode); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) DistanceMeasure(ctx context.Context, req *mapsv1.AmapDistanceMeasureReq) (resp *mapsv1.AmapDistanceMeasureResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	var origins []string
	for _, o := range req.Origins {
		origins = append(origins, fmt.Sprintf(locationFormat, o.Lng, o.Lat))
	}
	params.Add("origins", strings.Join(origins, "|"))
	params.Add("destination", fmt.Sprintf(locationFormat, req.Destination.Lng, req.Destination.Lat))
	if req.Type != nil {
		switch *req.Type {
		case enumv1.AmapDistanceMeasureType_AMAP_DISTANCE_MEASURE_TYPE_STRAIGHT:
			params.Add("type", "0")
		case enumv1.AmapDistanceMeasureType_AMAP_DISTANCE_MEASURE_TYPE_DRIVING:
			params.Add("type", "1")
		case enumv1.AmapDistanceMeasureType_AMAP_DISTANCE_MEASURE_TYPE_WALKING:
			params.Add("type", "3")
		}
	}
	resp = &mapsv1.AmapDistanceMeasureResp{}
	if err := m.requestProto(ctx, distanceMeasureURL, params, resp); err != nil {
		return nil, err
	}
	// --- parse error ---
	if err := m.parseError(resp.Status, resp.Info, resp.Infocode); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) TransitPlan(ctx context.Context, req *mapsv1.AmapTransitPlanReq) (resp *mapsv1.AmapTransitPlanResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	params.Add("origin", fmt.Sprintf(locationFormat, req.Origin.Lng, req.Origin.Lat))
	params.Add("destination", fmt.Sprintf(locationFormat, req.Destination.Lng, req.Destination.Lat))
	if req.Originpoi != nil && req.Destinationpoi != nil {
		params.Add("originpoi", *req.Originpoi)
		params.Add("destinationpoi", *req.Destinationpoi)
	}
	if req.Ad1 != nil && req.Ad2 != nil {
		params.Add("ad1", *req.Ad1)
		params.Add("ad2", *req.Ad2)
	}
	if req.City1 != nil && req.City2 != nil {
		params.Add("city1", *req.City1)
		params.Add("city2", *req.City2)
	}
	if req.Policy != nil {
		switch *req.Policy {
		case enumv1.AmapTransitPlanPolicy_AMAP_TRANSIT_PLAN_POLICY_RECOMMEND:
			params.Add("strategy", "0")
		case enumv1.AmapTransitPlanPolicy_AMAP_TRANSIT_PLAN_POLICY_LEAST_FEE:
			params.Add("strategy", "1")
		case enumv1.AmapTransitPlanPolicy_AMAP_TRANSIT_PLAN_POLICY_LEAST_TRANSFER:
			params.Add("strategy", "2")
		case enumv1.AmapTransitPlanPolicy_AMAP_TRANSIT_PLAN_POLICY_LEAST_WALKING:
			params.Add("strategy", "3")
		case enumv1.AmapTransitPlanPolicy_AMAP_TRANSIT_PLAN_POLICY_MOST_COMFORTABLE:
			params.Add("strategy", "4")
		case enumv1.AmapTransitPlanPolicy_AMAP_TRANSIT_PLAN_POLICY_NO_SUBWAY:
			params.Add("strategy", "5")
		case enumv1.AmapTransitPlanPolicy_AMAP_TRANSIT_PLAN_POLICY_SUBWAY:
			params.Add("strategy", "6")
		case enumv1.AmapTransitPlanPolicy_AMAP_TRANSIT_PLAN_POLICY_SUBWAY_FIRST:
			params.Add("strategy", "7")
		case enumv1.AmapTransitPlanPolicy_AMAP_TRANSIT_PLAN_POLICY_LEAST_TIME:
			params.Add("strategy", "8")
		}
	}
	if req.AlternativeRoute != nil {
		params.Add("AlternativeRoute", strconv.Itoa(int(*req.AlternativeRoute)))
	}
	if req.Multiexport != nil {
		if *req.Multiexport {
			params.Add("multiexport", "1")
		} else {
			params.Add("multiexport", "0")
		}
	}
	if req.Nightflag != nil {
		if *req.Nightflag {
			params.Add("nightflag", "1")
		} else {
			params.Add("nightflag", "0")
		}
	}
	if req.Date != nil {
		params.Add("date", *req.Date)
	}
	if req.Time != nil {
		params.Add("time", *req.Time)
	}
	var fields []string
	for _, field := range req.ShowFields {
		fields = append(fields, field)
	}
	if len(fields) > 0 {
		params.Add("show_fields", strings.Join(fields, ","))
	}
	resp = &mapsv1.AmapTransitPlanResp{}
	if err := m.requestProto(ctx, transitPlanURL, params, resp); err != nil {
		return nil, err
	}
	// --- parse error ---
	if err := m.parseError(resp.Status, resp.Info, resp.Infocode); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) GetLocationByIp(ctx context.Context, req *mapsv1.AmapGetLocationByIpReq) (resp *mapsv1.AmapGetLocationByIpResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	if req.GetIp() != "" {
		params.Add("ip", req.GetIp())
	}
	resp = &mapsv1.AmapGetLocationByIpResp{}
	if err := m.requestProto(ctx, getLocationByIpURL, params, resp); err != nil {
		return nil, err
	}
	if err := m.parseError(resp.Status, resp.Info, resp.Infocode); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) getParamsURL(addr string, params url.Values) string {
	return fmt.Sprintf(urlFormat, addr, params.Encode())
}

func (m *MapService) parseError(status, info, infoCode string) error {
	if status == statusOK {
		return nil
	}
	switch infoCode {
	case statusIntervalLimit:
		return ecode.ErrMapReachIntervalLimit
	case statusDailyLimit:
		return ecode.ErrMapReachIntervalLimit
	default:
		return fmt.Errorf(errFormat, status, infoCode, info)
	}
}

// 高德地图的接口有问题，对没有的数据返回的是 "[]"，会导致反序列化失败
// requestProto 发起请求然后把响应反序列化到传入的 proto.Message（使用 protojson）
// Caller 必须传入一个已分配的 proto.Message 指针（例如 &mapsv1.AmapGetDistrictsResp{}）
func (m *MapService) requestProto(ctx context.Context, addr string, params url.Values, msg proto.Message) error {
	reqURL := m.getParamsURL(addr, params)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	// 快路径：没有 ":[]" 则直接 protojson.Unmarshal
	if !bytes.Contains(body, []byte(":[]")) {
		if err := protojson.Unmarshal(body, msg); err != nil {
			return fmt.Errorf("protojson unmarshal failed: %w", err)
		}
		return nil
	}

	// 需要清洗，再反序列化到 proto.Message
	var tmp any
	if err := jsonx.Unmarshal(body, &tmp); err != nil {
		return fmt.Errorf("json unmarshal for cleaning failed: %w", err)
	}
	cleanEmptyArrays(tmp)

	cleanedBytes, err := jsonx.Marshal(tmp)
	if err != nil {
		return fmt.Errorf("json marshal after cleaning failed: %w", err)
	}

	if err := protojson.Unmarshal(cleanedBytes, msg); err != nil {
		return fmt.Errorf("protojson unmarshal after cleaning failed: %w", err)
	}
	return nil
}

// cleanEmptyArrays 递归删除 JSON 结构中值为 [] 的字段
// 入参是 json.Unmarshal 后得到的 any（map[string]any / []any / primitive）
func cleanEmptyArrays(obj any) any {
	switch v := obj.(type) {
	case map[string]any:
		for k, val := range v {
			// 如果是空数组则删除这个键
			if arr, ok := val.([]any); ok && len(arr) == 0 {
				delete(v, k)
				continue
			}
			// 递归
			v[k] = cleanEmptyArrays(val)
		}
		return v
	case []any:
		for i := range v {
			v[i] = cleanEmptyArrays(v[i])
		}
		return v
	default:
		return v
	}
}

// --- helper mapping functions ---
func amapStrategyToCode(strategy enumv1.AmapDrivingPlanStrategy) string {
	switch strategy {
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_DEFAULT:
		return "32"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_SPEED_FIRST:
		return "0"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_FARE_FIRST:
		return "1"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_NORMAL_FAST:
		return "2"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_AVOID_CONGESTION:
		return "33"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_HIGHWAY_FIRST:
		return "34"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_NO_HIGHWAY:
		return "35"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_LEAST_FARE:
		return "36"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_HIGHROAD_FIRST:
		return "37"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_FASTEST:
		return "38"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_AVOID_CONGESTION_HIGHWAY_FIRST:
		return "39"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_AVOID_CONGESTION_NO_HIGHWAY:
		return "40"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_AVOID_CONGESTION_LEAST_FARE:
		return "41"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_LEAST_FARE_NO_HIGHWAY:
		return "42"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_AVOID_CONGESTION_LEAST_FARE_NO_HIGHWAY:
		return "43"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_AVOID_CONGESTION_HIGHROAD_FIRST:
		return "44"
	case enumv1.AmapDrivingPlanStrategy_AMAP_DRIVING_PLAN_STRATEGY_AVOID_CONGESTION_FASTEST:
		return "45"
	default:
		return "32"
	}
}

func amapCarTypeToCode(t enumv1.CarType) string {
	switch t {
	case enumv1.CarType_CAR_TYPE_ELECTRIC:
		return "1"
	case enumv1.CarType_CAR_TYPE_HYBRID:
		return "2"
	default:
		return "0"
	}
}
