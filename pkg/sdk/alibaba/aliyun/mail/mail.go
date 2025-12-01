package mail

import (
	"errors"
	"strconv"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	mailCli "github.com/alibabacloud-go/dm-20151123/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	dateFormat     = "2006-01-02"
	timeFormat     = "2006-01-02 15:04"
	fullTimeFormat = "2021-04-28T17:11Z"
)

type Mail struct {
	accessKeyId     string
	accessKeySecret string
	securityToken   *string
	cli             *mailCli.Client
}

func NewMail(opts *msgv1.MailApiConfig) (*Mail, error) {
	cli, err := mailCli.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(opts.AccessKey),
		AccessKeySecret: tea.String(opts.SecretKey),
		SecurityToken:   opts.SecurityToken,
	})
	if err != nil {
		return nil, err
	}
	return &Mail{
		accessKeyId:     opts.AccessKey,
		accessKeySecret: opts.SecretKey,
		securityToken:   opts.SecurityToken,
		cli:             cli,
	}, nil
}

func (m *Mail) QueryMailSendStatistics(req *msgv1.QueryMailSendStatisticsReq) (resp *msgv1.QueryMailSendStatisticsResp, err error) {
	request := &mailCli.SenderStatisticsByTagNameAndBatchIDRequest{
		AccountName:       req.FromAddress,
		DedicatedIp:       req.DedicatedIp,
		DedicatedIpPoolId: req.DedicatedIpPoolId,
		EndTime:           tea.String(req.EndTime.AsTime().Format(dateFormat)),
		Esp:               req.Esp,
		StartTime:         tea.String(req.StartTime.AsTime().Format(dateFormat)),
		TagName:           req.TagName,
	}
	res, err := m.cli.SenderStatisticsByTagNameAndBatchID(request)
	if err != nil {
		if res != nil {
			err = m.joinRequestID(err, tea.StringValue(res.Body.RequestId))
		}
		return nil, err
	}
	statistics := make([]*msgv1.MailSendStatistic, 0, len(res.Body.Data.Stat))
	for _, v := range res.Body.Data.Stat {
		total, err := m.parseStringToInt64(tea.StringValue(v.RequestCount))
		if err != nil {
			return nil, err
		}
		success, err := m.parseStringToInt64(tea.StringValue(v.SuccessCount))
		if err != nil {
			return nil, err
		}
		fail, err := m.parseStringToInt64(tea.StringValue(v.FaildCount))
		if err != nil {
			return nil, err
		}
		unavailable, err := m.parseStringToInt64(tea.StringValue(v.UnavailableCount))
		if err != nil {
			return nil, err
		}
		t, err := time.Parse(dateFormat, tea.StringValue(v.CreateTime))
		if err != nil {
			return nil, err
		}
		statistics = append(statistics, &msgv1.MailSendStatistic{
			Total:       total,
			Success:     success,
			Failed:      fail,
			Unavailable: unavailable,
			Date:        timestamppb.New(t),
		})
	}
	resp = &msgv1.QueryMailSendStatisticsResp{Statistics: statistics}
	return resp, nil
}

func (m *Mail) QueryMailSendDetail(req *msgv1.QueryMailSendDetailsReq) (resp *msgv1.QueryMailSendDetailsResp, err error) {
	request := &mailCli.SenderStatisticsDetailByParamRequest{
		AccountName: req.FromAddress,
		ConfigSetId: req.ConfigSetId,
		IpPoolId:    req.IpPool_Id,
		Length:      tea.Int32(int32(req.PageSize)),
		NextStart:   req.NextStart,
		TagName:     req.TagName,
		ToAddress:   req.ToAddress,
	}
	if req.Status != nil {
		request.Status = tea.Int32(m.statusToSendMailStatus(*req.Status))
	}
	if req.StartTime != nil {
		request.StartTime = tea.String(req.StartTime.AsTime().Format(timeFormat))
	}
	if req.EndTime != nil {
		request.EndTime = tea.String(req.EndTime.AsTime().Format(timeFormat))
	}
	res, err := m.cli.SenderStatisticsDetailByParam(request)
	if err != nil {
		if res != nil {
			err = m.joinRequestID(err, tea.StringValue(res.Body.RequestId))
		}
		return nil, err
	}
	details := make([]*msgv1.MailSendDetail, 0, len(res.Body.Data.MailDetail))
	for _, v := range res.Body.Data.MailDetail {
		updateTime, err := time.Parse(fullTimeFormat, tea.StringValue(v.LastUpdateTime))
		if err != nil {
			return nil, err
		}
		details = append(details, &msgv1.MailSendDetail{
			Status:              m.sendMailStatusToStatus(tea.Int32Value(v.Status)),
			LastUpdateTime:      timestamppb.New(updateTime),
			Message:             tea.StringValue(v.Message),
			ToAddress:           tea.StringValue(v.ToAddress),
			FromAddress:         tea.StringValue(v.AccountName),
			Subject:             tea.StringValue(v.Subject),
			ErrorClassification: tea.StringValue(v.ErrorClassification),
			IpPoolId:            tea.StringValue(v.IpPoolId),
			IpPoolName:          tea.StringValue(v.IpPoolName),
			ConfigSetId:         tea.StringValue(v.ConfigSetId),
			ConfigSetName:       tea.StringValue(v.ConfigSetName),
		})
	}
	resp = &msgv1.QueryMailSendDetailsResp{Details: details}
	return resp, nil
}

func (m *Mail) QueryMailTracks(req *msgv1.QueryMailTracksReq) (resp *msgv1.QueryMailTracksResp, err error) {
	request := &mailCli.GetTrackListRequest{
		AccountName:       req.FromAddress,
		ConfigSetId:       req.ConfigSetId,
		DedicatedIp:       req.DedicatedIp,
		DedicatedIpPoolId: req.DedicatedIp,
		Esp:               req.Esp,
		TagName:           req.TagName,
	}
	if req.StartTime != nil {
		request.StartTime = tea.String(req.StartTime.AsTime().Format(dateFormat))
	}
	if req.EndTime != nil {
		request.EndTime = tea.String(req.EndTime.AsTime().Format(dateFormat))
	}
	if req.Page != nil {
		request.PageNumber = tea.String(strconv.FormatInt(*req.Page, 10))
	}
	if req.Size != nil {
		request.PageSize = tea.String(strconv.FormatInt(*req.Size, 10))
	}
	res, err := m.cli.GetTrackList(request)
	if err != nil {
		if res != nil {
			err = m.joinRequestID(err, tea.StringValue(res.Body.RequestId))
		}
		return nil, err
	}
	tracks := make([]*msgv1.MailSendTrack, 0, len(res.Body.Data.Stat))
	for _, v := range res.Body.Data.Stat {
		total, err := m.parseStringToInt64(tea.StringValue(v.TotalNumber))
		if err != nil {
			return nil, err
		}
		open, err := m.parseStringToInt64(tea.StringValue(v.RcptOpenCount))
		if err != nil {
			return nil, err
		}
		click, err := m.parseStringToInt64(tea.StringValue(v.RcptClickCount))
		if err != nil {
			return nil, err
		}
		uniqueOpen, err := m.parseStringToInt64(tea.StringValue(v.RcptUniqueClickCount))
		if err != nil {
			return nil, err
		}
		uniqueClick, err := m.parseStringToInt64(tea.StringValue(v.RcptUniqueClickCount))
		if err != nil {
			return nil, err
		}
		clickRate, err := m.parseStringToFloat64(tea.StringValue(v.RcptClickRate))
		if err != nil {
			return nil, err
		}
		openRate, err := m.parseStringToFloat64(tea.StringValue(v.RcptOpenRate))
		if err != nil {
			return nil, err
		}
		uniqueClickRate, err := m.parseStringToFloat64(tea.StringValue(v.RcptUniqueClickRate))
		if err != nil {
			return nil, err
		}
		uniqueOpenRate, err := m.parseStringToFloat64(tea.StringValue(v.RcptUniqueOpenRate))
		if err != nil {
			return nil, err
		}
		t, err := time.Parse(fullTimeFormat, tea.StringValue(v.CreateTime))
		tracks = append(tracks, &msgv1.MailSendTrack{
			Total:           total,
			Open:            open,
			Click:           click,
			UniqueOpen:      uniqueOpen,
			UniqueClick:     uniqueClick,
			ClickRate:       clickRate,
			OpenRate:        openRate,
			UniqueClickRate: uniqueClickRate,
			UniqueOpenRate:  uniqueOpenRate,
			CreateTime:      timestamppb.New(t),
		})
	}
	resp = &msgv1.QueryMailTracksResp{Tracks: tracks}
	return resp, nil
}

func (m *Mail) parseStringToInt64(v string) (int64, error) {
	return strconv.ParseInt(v, 10, 64)
}

func (m *Mail) parseStringToFloat64(v string) (float64, error) {
	return strconv.ParseFloat(v, 64)
}

func (m *Mail) joinRequestID(err error, requestID string) error {
	if err == nil {
		return nil
	}
	return errors.Join(err, errors.New(requestID))
}

func (m *Mail) statusToSendMailStatus(status enumv1.MailSendStatus) int32 {
	switch status {
	case enumv1.MailSendStatus_MAIL_SEND_STATUS_SUCCESS:
		return 0
	case enumv1.MailSendStatus_MAIL_SEND_STATUS_INVALID_ADDRESS:
		return 2
	case enumv1.MailSendStatus_MAIL_SEND_STATUS_SPAM_MAIL:
		return 3
	case enumv1.MailSendStatus_MAIL_SEND_STATUS_FAILED:
		return 4
	}
	return 0
}

func (m *Mail) sendMailStatusToStatus(status int32) enumv1.MailSendStatus {
	switch status {
	case 0:
		return enumv1.MailSendStatus_MAIL_SEND_STATUS_SUCCESS
	case 2:
		return enumv1.MailSendStatus_MAIL_SEND_STATUS_INVALID_ADDRESS
	case 3:
		return enumv1.MailSendStatus_MAIL_SEND_STATUS_SPAM_MAIL
	case 4:
		return enumv1.MailSendStatus_MAIL_SEND_STATUS_FAILED
	}
	return enumv1.MailSendStatus_MAIL_SEND_STATUS_UNSPECIFIED
}
