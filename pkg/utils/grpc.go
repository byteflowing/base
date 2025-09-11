package utils

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/byteflowing/go-common/idx"
	"github.com/byteflowing/go-common/logx"
	logv1 "github.com/byteflowing/proto/gen/go/log/v1"
)

func UnaryLogIDInterceptor(conf *logv1.LogConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if conf.CtxLogIdKey != "" {
			md, ok := metadata.FromIncomingContext(ctx)
			var logId string
			if ok {
				if vals := md.Get(conf.CtxLogIdKey); len(vals) > 0 {
					logId = vals[0]
				}
			}
			if logId == "" {
				logId = idx.UUIDv4()
			}
			ctx = logx.CtxWithLogID(ctx, logId)
		}
		return handler(ctx, req)
	}
}

func UnaryOutgoingLogIDInterceptor(conf *logv1.LogConfig) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		logId := logx.GetCtxLogID(ctx)
		if logId != "" {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			} else {
				md = md.Copy()
			}
			md.Set(conf.CtxLogIdKey, logId)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func UnaryLoggingInterceptor(conf *logv1.LogConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()
		logx.CtxDebug(ctx,
			"gRPC request",
			zap.String("method", info.FullMethod),
			zap.Any("req", req),
		)
		resp, err = handler(ctx, req)
		// 记录非业务产生的错误日志
		if err != nil && needLogging(err, conf.RecordAllErrors) {
			logx.CtxError(ctx,
				"gPRC request failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
		}
		duration := time.Since(start)
		logx.CtxDebug(ctx,
			"gRPC response",
			zap.String("method", info.FullMethod),
			zap.Any("resp", resp),
			zap.Int64("latency(ms)", duration.Milliseconds()),
		)
		return resp, err
	}
}

func needLogging(err error, recordAll bool) bool {
	if recordAll {
		return true
	}
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unimplemented, codes.Internal, codes.Unavailable, codes.Unknown:
			return true
		default:
			return false
		}
	}
	return true
}
