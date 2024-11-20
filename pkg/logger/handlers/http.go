package handlers

import (
	"context"
	"net/http"
	"time"

	"distributed-logging-poc/pkg/logger"
	_logger "distributed-logging-poc/pkg/logger"
)

func HandleHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, logService *logger.Logger) {
	traceID, _ := ctx.Value("trace_id").(string)
	spanID, _ := ctx.Value("span_id").(string)

	start := time.Now()

	duration := time.Since(start)

	metadata := _logger.RequestMetadata{
		TraceID:      traceID,
		SpanID:       spanID,
		RequestID:    _logger.GenerateRequestID("req"),
		ServiceName:  "http-service",
		Timestamp:    time.Now(),
		UserAgent:    r.UserAgent(),
		ClientIP:     r.RemoteAddr,
		Method:       r.Method,
		Path:         r.URL.Path,
		ResponseTime: duration,
		StatusCode:   http.StatusOK,
		Extra: map[string]interface {
		}{
			"extra_field": "extra_value",
		},
	}

	logService.Info(ctx, "HTTP request processed", metadata)
}
