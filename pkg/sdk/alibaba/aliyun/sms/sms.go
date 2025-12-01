// Package sms
// 文档：https://help.aliyun.com/zh/sms/developer-reference/api-dysmsapi-2017-05-25-sendsms?spm=a2c4g.11186623.0.0.31ba614c4SMnC0
package sms

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	smsCli "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/byteflowing/base/pkg/utils"
	"github.com/byteflowing/go-common/jsonx"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	dateFormat = "20060102"
)

type Sms struct {
	accessKeyId     string
	accessKeySecret string
	securityToken   *string
	cli             *smsCli.Client
}

func New(opts *msgv1.SmsProvider) (s *Sms, err error) {
	smsClient, err := smsCli.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(opts.AccessKey),
		AccessKeySecret: tea.String(opts.SecretKey),
		SecurityToken:   opts.SecurityToken,
	})
	return &Sms{
		accessKeyId:     opts.AccessKey,
		accessKeySecret: opts.SecretKey,
		securityToken:   opts.SecurityToken,
		cli:             smsClient,
	}, nil
}

func (s *Sms) SendSms(req *msgv1.SendSmsReq) (resp *msgv1.SendSmsResp, err error) {
	var params string
	if len(req.TemplateParams) > 0 {
		params, err = jsonx.MarshalToString(req.TemplateParams)
		if err != nil {
			return nil, err
		}
	}
	request := &smsCli.SendSmsRequest{
		PhoneNumbers: tea.String(req.PhoneNumber.Number),
		SignName:     tea.String(req.SignName),
		TemplateCode: tea.String(req.TemplateCode),
	}
	if len(params) > 0 {
		request.TemplateParam = tea.String(params)
	}
	res, err := s.cli.SendSms(request)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("response is nil")
	}
	err = s.parseErr(res.Body.RequestId, res.Body.Code, res.Body.Message)
	return
}

func (s *Sms) QuerySendStatistics(req *msgv1.QuerySmsStatisticsReq) (resp *msgv1.QuerySmsStatisticsResp, err error) {
	if req.StartDate == nil || req.EndDate == nil {
		return nil, errors.New("start date or end date is nil")
	}
	request := &smsCli.QuerySendStatisticsRequest{
		EndDate:   tea.String(req.EndDate.AsTime().Format(dateFormat)),
		PageIndex: tea.Int32(int32(req.Page)),
		PageSize:  tea.Int32(int32(req.Size)),
		SignName:  req.SignName,
		StartDate: tea.String(req.StartDate.AsTime().Format(dateFormat)),
	}
	if req.TemplateType != nil {
		request.TemplateType = tea.Int32(s.parseTemplateType(*req.TemplateType))
	}
	if req.IsGlobal {
		request.IsGlobe = tea.Int32(2)
	} else {
		request.IsGlobe = tea.Int32(1)
	}
	res, err := s.cli.QuerySendStatistics(request)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("response is nil")
	}
	if err := s.parseErr(res.Body.RequestId, res.Body.Code, res.Body.Message); err != nil {
		return nil, err
	}
	total := tea.Int64Value(res.Body.Data.TotalSize)
	statistics := make([]*msgv1.SmsSendStatistic, 0, len(res.Body.Data.TargetList))
	for _, v := range res.Body.Data.TargetList {
		t, err := time.Parse(dateFormat, tea.StringValue(v.SendDate))
		if err != nil {
			return nil, err
		}
		statistics = append(statistics, &msgv1.SmsSendStatistic{
			Total:      tea.Int64Value(v.TotalCount),
			Success:    tea.Int64Value(v.RespondedSuccessCount),
			Failed:     tea.Int64Value(v.RespondedFailCount),
			NoResponse: tea.Int64Value(v.NoRespondedCount),
			Date:       timestamppb.New(t),
		})
	}
	resp = &msgv1.QuerySmsStatisticsResp{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: utils.CalcTotalPage(total, req.Size),
		Vendor:     req.Vendor,
		Account:    req.Account,
		Statistics: statistics,
	}
	return resp, nil
}

func (s *Sms) QuerySendDetail(req *msgv1.QuerySmsSendDetailReq) (resp *msgv1.QuerySmsSendDetailResp, err error) {
	request := &smsCli.QuerySendDetailsRequest{
		CurrentPage: tea.Int64(int64(req.Page)),
		PageSize:    tea.Int64(int64(req.Size)),
		PhoneNumber: tea.String(req.PhoneNumber.Number),
		SendDate:    tea.String(req.SendDate.AsTime().Format(dateFormat)),
	}
	res, err := s.cli.QuerySendDetails(request)
	if err != nil {
		return nil, err
	}
	if err := s.parseErr(res.Body.RequestId, res.Body.Code, res.Body.Message); err != nil {
		return nil, err
	}
	details := make([]*msgv1.SmsSendDetail, 0, len(res.Body.SmsSendDetailDTOs.SmsSendDetailDTO))
	for _, d := range res.Body.SmsSendDetailDTOs.SmsSendDetailDTO {
		sendDate, err := time.Parse(dateFormat, tea.StringValue(d.SendDate))
		if err != nil {
			return nil, err
		}
		receiveDate, err := time.Parse(dateFormat, tea.StringValue(d.ReceiveDate))
		if err != nil {
			return nil, err
		}
		details = append(details, &msgv1.SmsSendDetail{
			PhoneNumber:  tea.StringValue(d.PhoneNum),
			Content:      tea.StringValue(d.Content),
			Status:       s.aliSendStatusToStatus(d.SendStatus),
			TemplateCode: tea.StringValue(d.TemplateCode),
			OutId:        tea.StringValue(d.OutId),
			ErrCode:      tea.StringValue(d.ErrCode),
			SendDate:     timestamppb.New(sendDate),
			ReceiveDate:  timestamppb.New(receiveDate),
		})
	}
	total, err := strconv.ParseInt(tea.StringValue(res.Body.TotalCount), 10, 64)
	if err != nil {
		return nil, err
	}
	resp = &msgv1.QuerySmsSendDetailResp{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: utils.CalcTotalPage(total, req.Size),
		Vendor:     req.Vendor,
		Account:    req.Account,
		Details:    details,
	}
	return resp, nil
}

func (s *Sms) parseErr(requestID, errCode, errMsg *string) (err error) {
	_requestID := tea.StringValue(requestID)
	_errCode := tea.StringValue(errCode)
	_errMsg := tea.StringValue(errMsg)
	if _errCode != "OK" {
		return fmt.Errorf("[requestID:%s, code:%s] errMsg:%s", _requestID, _errCode, _errMsg)
	}
	return nil
}

func (s *Sms) parseTemplateType(templateType enumv1.SmsTemplateType) int32 {
	switch templateType {
	case enumv1.SmsTemplateType_SMS_TEMPLATE_TYPE_CAPTCHA:
		return 0
	case enumv1.SmsTemplateType_SMS_TEMPLATE_TYPE_NOTIFICATION:
		return 1
	case enumv1.SmsTemplateType_SMS_TEMPLATE_TYPE_AD:
		return 2
	case enumv1.SmsTemplateType_SMS_TEMPLATE_TYPE_INTERNATIONAL:
		return 3
	case enumv1.SmsTemplateType_SMS_TEMPLATE_TYPE_DIGITAL:
		return 7
	}
	return 0
}

func (s *Sms) aliSendStatusToStatus(st *int64) enumv1.SmsSendStatus {
	if st == nil {
		return enumv1.SmsSendStatus_SMS_SEND_STATUS_UNSPECIFIED
	}
	switch *st {
	case 1:
		return enumv1.SmsSendStatus_SMS_SEND_STATUS_WAIT_RESPONSE
	case 2:
		return enumv1.SmsSendStatus_SMS_SEND_STATUS_FAILED
	case 3:
		return enumv1.SmsSendStatus_SMS_SEND_STATUS_SUCCESS
	}
	return enumv1.SmsSendStatus_SMS_SEND_STATUS_UNSPECIFIED
}
