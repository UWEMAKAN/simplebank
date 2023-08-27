package gapi

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	startTime := time.Now()
	resp, err = handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	logger := log.Info()
	if err != nil {
		logger = log.Error().Err(err)
	}

	logger.
		Str("protocol", "grpc").
		Str("method", info.FullMethod).
		Int("statusCode", int(statusCode)).
		Str("statusText", statusCode.String()).
		Dur("duration", duration).
		Msg("received a gRPC unary request")

	return resp, err
}

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body []byte
}

func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func (rec *ResponseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		rec := &ResponseRecorder{
			ResponseWriter: res,
			StatusCode: http.StatusOK,
		}
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)
		logger := log.Info()
		if rec.StatusCode >= http.StatusBadRequest {
			logger = log.Error().Bytes("body", rec.Body)
		}
		logger.
			Str("protocol", "http").
			Str("method", req.Method).
			Str("path", req.RequestURI).
			Int("statusCode", int(rec.StatusCode)).
			Str("statusText", http.StatusText(rec.StatusCode)).
			Dur("duration", duration).
			Msg("received a http request")
	})
}
