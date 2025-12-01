package lbs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/byteflowing/base/pkg/httpx"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	urlFormat                  = "%s?%s"
	errFormat                  = "code:%s, msg:%s"
	walkingRoutePlanUrl        = "https://mapapi.cloud.huawei.com/mapApi/v1/routeService/walking"
	bicyclingRoutePlanUrl      = "https://mapapi.cloud.huawei.com/mapApi/v1/routeService/bicycling"
	drivingRoutePlanUrl        = "https://mapapi.cloud.huawei.com/mapApi/v1/routeService/driving"
	distanceMatrixWalkingUrl   = "https://mapapi.cloud.huawei.com/mapApi/v1/routeService/walkingMatrix"
	distanceMatrixBicyclingUrl = "https://mapapi.cloud.huawei.com/mapApi/v1/routeService/bicyclingMatrix"
	distanceMatrixDrivingUrl   = "https://mapapi.cloud.huawei.com/mapApi/v1/routeService/drivingMatrix"
	getIpByLocationUrl         = "https://openlocation-drcn.platform.dbankcloud.com/networklocation/v1/ipLocation"
	addressToLocationUrl       = "https://siteapi.cloud.huawei.com/mapApi/v1/siteService/geocode"
	locationToAddrUrl          = "https://siteapi.cloud.huawei.com/mapApi/v1/siteService/reverseGeocode"
	getTimezoneUrl             = "https://siteapi.cloud.huawei.com/mapApi/v1/timezoneService/getTimezone"
)

type MapService struct {
	httpClient *http.Client
}

func NewMapService() *MapService {
	return &MapService{
		httpClient: httpx.NewClient(httpx.GetDefaultConfig()),
	}
}

// RoutePlan
// 文档：https://developer.huawei.com/consumer/cn/doc/HMSCore-Guides/web-diretions-api-introduction-0000001231972971
func (m *MapService) RoutePlan(_ context.Context, req *mapsv1.HuaweiRoutePlanReq) (resp *mapsv1.HuaweiRoutePlanResp, err error) {
	var addr string
	if req.Mode == enumv1.PlanMode_PLAN_MODE_DRIVING {
		addr = drivingRoutePlanUrl
	} else if req.Mode == enumv1.PlanMode_PLAN_MODE_WALKING {
		addr = walkingRoutePlanUrl
	} else if req.Mode == enumv1.PlanMode_PLAN_MODE_BICYCLING {
		addr = bicyclingRoutePlanUrl
	} else {
		return nil, fmt.Errorf("invalid mode: %s", req.Mode)
	}
	resp = &mapsv1.HuaweiRoutePlanResp{}
	if err := m.requestPost(req.Key, addr, req, resp); err != nil {
		return nil, err
	}
	if err := m.parseError(resp.ReturnCode, resp.ReturnDesc); err != nil {
		return nil, err
	}
	return resp, nil
}

// DistanceMatrix
// 文档： https://developer.huawei.com/consumer/cn/doc/HMSCore-Guides/web-matrix-api-introduction-0000001232294527
func (m *MapService) DistanceMatrix(_ context.Context, req *mapsv1.HuaweiDistanceMatrixPlanReq) (resp *mapsv1.HuaweiDistanceMatrixPlanResp, err error) {
	var addr string
	if req.Mode == enumv1.PlanMode_PLAN_MODE_DRIVING {
		addr = distanceMatrixDrivingUrl
	} else if req.Mode == enumv1.PlanMode_PLAN_MODE_WALKING {
		addr = distanceMatrixWalkingUrl
	} else if req.Mode == enumv1.PlanMode_PLAN_MODE_BICYCLING {
		addr = distanceMatrixBicyclingUrl
	} else {
		return nil, fmt.Errorf("invalid mode: %s", req.Mode)
	}
	resp = &mapsv1.HuaweiDistanceMatrixPlanResp{}
	if err := m.requestPost(req.Key, addr, req, resp); err != nil {
		return nil, err
	}
	if err := m.parseError(resp.ReturnCode, resp.ReturnDesc); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) GetLocationByIp(_ context.Context, req *mapsv1.HuaweiGetLocationByIpReq) (resp *mapsv1.HuaweiGetLocationByIpResp, err error) {
	h := http.Header{}
	h.Add("Content-Type", "application/json")
	h.Add("Authorization", "Bearer "+req.GetKey())
	h.Add("x-forwarded-for", req.GetIp())
	requestBody, err := protojson.Marshal(req)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodPost, getIpByLocationUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	request.Header = h
	res, err := m.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	resp = &mapsv1.HuaweiGetLocationByIpResp{}
	if err := protojson.Unmarshal(body, resp); err != nil {
		return nil, err
	}
	if resp.ErrCode != "0" {
		return nil, fmt.Errorf(errFormat, resp.ErrCode, resp.ErrorMsg)
	}
	return resp, nil
}

func (m *MapService) AddrToLocation(_ context.Context, req *mapsv1.HuaweiAddressToLocationReq) (resp *mapsv1.HuaweiAddressToLocationResp, err error) {
	resp = &mapsv1.HuaweiAddressToLocationResp{}
	if err := m.requestPost(req.Key, addressToLocationUrl, req, resp); err != nil {
		return nil, err
	}
	if err := m.parseError(resp.ReturnCode, resp.ReturnDesc); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) LocationToAddr(_ context.Context, req *mapsv1.HuaweiLocationToAddrReq) (resp *mapsv1.HuaweiLocationToAddrResp, err error) {
	resp = &mapsv1.HuaweiLocationToAddrResp{}
	if err := m.requestPost(req.Key, locationToAddrUrl, req, resp); err != nil {
		return nil, err
	}
	if err := m.parseError(resp.ReturnCode, resp.ReturnDesc); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) GetTimezone(_ context.Context, req *mapsv1.HuaweiGetTimezoneByLocationReq) (resp *mapsv1.HuaweiGetTimezoneByLocationResp, err error) {
	resp = &mapsv1.HuaweiGetTimezoneByLocationResp{}
	if err := m.requestPost(req.Key, getTimezoneUrl, req, resp); err != nil {
		return nil, err
	}
	if err := m.parseError(resp.ReturnCode, resp.ReturnDesc); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MapService) requestPost(key, addr string, req, resp proto.Message) (err error) {
	params := url.Values{}
	params.Add("key", key)
	requestBody, err := protojson.Marshal(req)
	if err != nil {
		return err
	}
	requestUrl := fmt.Sprintf(urlFormat, addr, params.Encode())
	res, err := m.httpClient.Post(requestUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return protojson.Unmarshal(body, resp)
}

func (m *MapService) parseError(returnCode, returnDesc string) error {
	if returnCode == "0" && returnDesc == "OK" {
		return nil
	}
	return fmt.Errorf(errFormat, returnCode, returnDesc)
}
