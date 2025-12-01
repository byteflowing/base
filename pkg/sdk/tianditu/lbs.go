package tianditu

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/byteflowing/base/pkg/httpx"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	urlFormat                = "%s?%s"
	locationFormat           = "%.6f,%.6f"
	coderKeywordFormat       = `{"keyWord":"%s"}`
	coderLocationFormat      = `{'lon':%.6f,'lat':%.6f}`
	drivingPlanFormat        = `{"orig":"%s","dest":"%s","style":"%d"}`
	drivingPlanWithMidFormat = `{"orig":"%s","dest":"%s","mid":"%s","style":"%d"}`
	transitPlanFormat        = `{"startposition":"%.6f,%.6f","endposition":"%.6f,%.6f","linetype":"%d"}`
	getLineInfoFormat        = `{"uuid":"%s"}`
	errFormat                = "status: %v, grpc_message: %v"
	getDistrictURL           = "https://api.tianditu.gov.cn/v2/administrative"
	geoCoderURL              = "https://api.tianditu.gov.cn/geocoder"
	drivingPlanURL           = "https://api.tianditu.gov.cn/drive"
	transitPlanURL           = "https://api.tianditu.gov.cn/transit"
)

const (
	LineTypeFastest     = 1 << 0 // 第0位：较快捷
	LineTypeLeastChange = 1 << 1 // 第1位：少换乘
	LineTypeLeastWalk   = 1 << 2 // 第2位：少步行
	LineTypeNoSubway    = 1 << 3 // 第3位：不坐地铁
)

type MapService struct {
	httpClient *http.Client
}

func NewMapService() *MapService {
	return &MapService{
		httpClient: httpx.NewClient(httpx.GetDefaultConfig()),
	}
}

func (m *MapService) GetDistricts(ctx context.Context, req *mapsv1.TianDiTuGetDistrictsReq) (resp *mapsv1.TianDiTuGetDistrictsResp, err error) {
	params := url.Values{}
	params.Add("tk", req.GetKey())
	params.Add("keyword", req.GetKeyword())
	if req.ChildLevel != nil {
		params.Add("childLevel", req.GetChildLevel())
	}
	if req.Extensions != nil {
		params.Add("extensions", getBoolString(req.GetExtensions()))
	}
	resp = &mapsv1.TianDiTuGetDistrictsResp{}
	if err := m.requestProto(ctx, getDistrictURL, params, resp); err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf(errFormat, resp.Status, resp.Message)
	}
	return resp, nil
}

func (m *MapService) AddressToLocation(ctx context.Context, req *mapsv1.TianDiTuAddressToLocationReq) (resp *mapsv1.TianDiTuAddressToLocationResp, err error) {
	params := url.Values{}
	params.Add("tk", req.GetKey())
	params.Add("ds", fmt.Sprintf(coderKeywordFormat, req.GetKeyWord()))
	resp = &mapsv1.TianDiTuAddressToLocationResp{}
	if err := m.requestProto(ctx, geoCoderURL, params, resp); err != nil {
		return nil, err
	}
	if resp.Status != "0" {
		return nil, fmt.Errorf(errFormat, resp.Status, resp.Msg)
	}
	return resp, nil
}

func (m *MapService) LocationToAddr(ctx context.Context, req *mapsv1.TianDiTuLocationToAddrReq) (resp *mapsv1.TianDiTuLocationToAddrResp, err error) {
	params := url.Values{}
	params.Add("tk", req.GetKey())
	params.Add("postStr", fmt.Sprintf(coderLocationFormat, req.Lng, req.Lat))
	resp = &mapsv1.TianDiTuLocationToAddrResp{}
	if err := m.requestProto(ctx, geoCoderURL, params, resp); err != nil {
		return nil, err
	}
	if resp.Status != "0" {
		return nil, fmt.Errorf(errFormat, resp.Status, resp.Msg)
	}
	return resp, nil
}

func (m *MapService) RoutePlan(ctx context.Context, req *mapsv1.TianDiTuRoutePlanReq) (resp *mapsv1.TianDiTuRoutePlanResp, err error) {
	params := url.Values{}
	params.Add("tk", req.GetKey())
	origin := fmt.Sprintf(locationFormat, req.Orig.Lng, req.Orig.Lat)
	dest := fmt.Sprintf(locationFormat, req.Dest.Lng, req.Dest.Lat)
	mid := make([]string, 0, len(req.Mid))
	for _, s := range req.Mid {
		mid = append(mid, fmt.Sprintf(locationFormat, s.Lng, s.Lat))
	}
	style := getStyle(req.GetStyle())
	if len(mid) > 0 {
		params.Add("postStr", fmt.Sprintf(drivingPlanWithMidFormat, origin, dest, strings.Join(mid, ";"), style))
	} else {
		params.Add("postStr", fmt.Sprintf(drivingPlanFormat, origin, dest, style))
	}
	res := &RoutePlanResp{}
	if err := m.requestXML(ctx, drivingPlanURL, params, res); err != nil {
		return nil, err
	}
	resp = routePlanRespToProtoResp(res)
	return resp, nil
}

func (m *MapService) TransitPlan(ctx context.Context, req *mapsv1.TianDiTuTransitPlanReq) (resp *mapsv1.TianDiTuTransitPlanResp, err error) {
	resp = &mapsv1.TianDiTuTransitPlanResp{}
	switch req.Type {
	case enumv1.TianDiTuTransitType_TIAN_DI_TU_TRANSIT_TYPE_PLAN:
		resp.Plan, err = m.transitPlan(ctx, req.GetPlan())
	case enumv1.TianDiTuTransitType_TIAN_DI_TU_TRANSIT_TYPE_LINE:
		resp.LineInfo, err = m.getTransitLineInfo(ctx, req.GetLineInfo())
	case enumv1.TianDiTuTransitType_TIAN_DI_TU_TRANSIT_TYPE_STATION:
		resp.StationInfo, err = m.getTransitStationInfo(ctx, req.GetStationInfo())
	}
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (m *MapService) transitPlan(ctx context.Context, req *mapsv1.TianDiTuTransitRoutePlanReq) (resp *mapsv1.TianDiTuTransitRoutePlanResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	params.Add("type", "busline")
	lineType := getLineType(req)
	params.Add("postStr", fmt.Sprintf(transitPlanFormat, req.StartPosition.Lng, req.StartPosition.Lat, req.EndPosition.Lng, req.EndPosition.Lat, lineType))
	resp = &mapsv1.TianDiTuTransitRoutePlanResp{}
	if err := m.requestProto(ctx, transitPlanURL, params, resp); err != nil {
		return nil, err
	}
	if resp.ResultCode != 0 {
		return nil, fmt.Errorf("code: %d", resp.ResultCode)
	}
	return resp, nil
}

func (m *MapService) getTransitLineInfo(ctx context.Context, req *mapsv1.TianDiTuGetTransitLineInfoReq) (resp *mapsv1.TianDiTuGetTransitLineInfoResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	params.Add("postStr", fmt.Sprintf(getLineInfoFormat, req.GetUuid()))
	resp = &mapsv1.TianDiTuGetTransitLineInfoResp{}
	if err := m.requestProto(ctx, transitPlanURL, params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) getTransitStationInfo(ctx context.Context, req *mapsv1.TianDiTuGetStationInfoReq) (resp *mapsv1.TianDiTuGetStationInfoResp, err error) {
	params := url.Values{}
	params.Add("key", req.GetKey())
	params.Add("postStr", fmt.Sprintf(getLineInfoFormat, req.GetUuid()))
	resp = &mapsv1.TianDiTuGetStationInfoResp{}
	if err := m.requestProto(ctx, transitPlanURL, params, resp); err != nil {
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

func (m *MapService) requestXML(ctx context.Context, addr string, params url.Values, msg interface{}) (err error) {
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
	err = xml.Unmarshal(body, msg)
	log.Println(err)
	return err
}

func getBoolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func getStyle(style enumv1.TianDiTuRoutePlanStyle) int {
	switch style {
	case enumv1.TianDiTuRoutePlanStyle_TIAN_DI_TU_ROUTE_PLAN_STYLE_FASTEST:
		return 0
	case enumv1.TianDiTuRoutePlanStyle_TIAN_DI_TU_ROUTE_PLAN_STYLE_LEAST_DISTANCE:
		return 1
	case enumv1.TianDiTuRoutePlanStyle_TIAN_DI_TU_ROUTE_PLAN_STYLE_AVOID_HIGHWAY:
		return 2
	case enumv1.TianDiTuRoutePlanStyle_TIAN_DI_TU_ROUTE_PLAN_STYLE_WALKING:
		return 3
	default:
		return 0
	}
}

func getLineType(req *mapsv1.TianDiTuTransitRoutePlanReq) int {
	var lineType int
	// 仅在用户有传该字段时才处理
	if req.IsLeastTime != nil && *req.IsLeastTime {
		lineType |= LineTypeFastest
	}
	if req.IsLeastTransfer != nil && *req.IsLeastTransfer {
		lineType |= LineTypeLeastChange
	}
	if req.IsLeastWalking != nil && *req.IsLeastWalking {
		lineType |= LineTypeLeastWalk
	}
	if req.IsNoSubway != nil && *req.IsNoSubway {
		lineType |= LineTypeNoSubway
	}
	return lineType
}
