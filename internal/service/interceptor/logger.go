package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryLoggerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		startTime := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)
		st := status.Convert(err)

		logger.Info("gRPC call",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Int("code", int(st.Code())),
			zap.String("message", st.Message()),
			zap.Any("request", req),
			zap.Bool("success", err == nil),
		)

		return resp, err
	}
}

func StreamLoggerInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		startTime := time.Now()

		wrapped := &serverStreamWrapper{ServerStream: ss, ctx: wrapContext(ss.Context(), logger)}

		err := handler(srv, wrapped)

		duration := time.Since(startTime)
		st := status.Convert(err)

		logger.Info("gRPC stream call",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Int("code", int(st.Code())),
			zap.String("message", st.Message()),
			zap.Bool("success", err == nil),
		)

		return err
	}
}

type serverStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *serverStreamWrapper) Context() context.Context {
	return w.ctx
}

func wrapContext(ctx context.Context, logger *zap.Logger) context.Context {
	return ctx
}
