package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	_trace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Severity string

const (
	SeverityDebug    Severity = "DEBUG"
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityError    Severity = "ERROR"
	SeverityCritical Severity = "CRITICAL"
)

type Logger struct {
	zapLogger *zap.Logger
	tracer    _trace.Tracer
}

type RequestMetadata struct {
	TraceID      string                 `json:"trace_id"`
	SpanID       string                 `json:"span_id"`
	RequestID    string                 `json:"request_id"`
	ServiceName  string                 `json:"service_name"`
	Timestamp    time.Time              `json:"timestamp"`
	UserAgent    string                 `json:"user_agent"`
	ClientIP     string                 `json:"client_ip"`
	Device       string                 `json:"device"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	ResponseTime time.Duration          `json:"response_time"`
	Deadline     *time.Time             `json:"deadline"`
	StatusCode   int                    `json:"status_code"`
	ErrorCode    *string                `json:"error_code"`
	ErrorType    *string                `json:"error_type"`
	Extra        map[string]interface{} `json:"extra,omitempty"`
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

var logCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "log_entries_total",
		Help: "Total number of log entries by level.",
	},
	[]string{"level"},
)

func init() {
	prometheus.MustRegister(logCounter)
}

type Config struct {
	ServiceName string
	Environment string // "dev", "stg", "prd"
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func NewLogger(ctx context.Context, config Config) (*Logger, error) {
	var zapLogger *zap.Logger

	if config.Environment == "prd" {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapLogger, _ = cfg.Build()
	} else {
		zapLogger, _ = zap.NewDevelopment()
	}

	otel.SetLogger(NewZapLogger(zapLogger))

	return &Logger{
		zapLogger: zapLogger,
		tracer:    otel.Tracer(config.ServiceName),
	}, nil
}

func InitTracer(serviceName string, environment string) (*trace.TracerProvider, error) {
	traceExporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to create OTLP trace exporter: %w", err)
	}

	metricExporter, err := otlpmetricgrpc.New(context.Background(),
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to create OTLP metric exporter: %w", err)
	}

	resource := resource.NewWithAttributes(
		resource.Default().SchemaURL(),
		attribute.String("service.name", serviceName),
		attribute.String("service.version", "1.0.0"),
		attribute.String("service.environment", environment),
	)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(resource),
	)

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(resource),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	return tp, nil
}

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.Tracer("http-middleware")
		ctx, span := tracer.Start(r.Context(), "HTTP Request")

		defer func() {
			if span != nil {
				span.End()
			}
		}()

		spanContext := span.SpanContext()
		traceID := spanContext.TraceID().String()
		spanID := spanContext.SpanID().String()

		if traceID == "" || spanID == "" {
			fmt.Println("TraceMiddleware: TraceID or SpanID is missing")
		}

		ctx = context.WithValue(ctx, "trace_id", traceID)
		ctx = context.WithValue(ctx, "span_id", spanID)

		lrw := &LoggingResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}

		start := time.Now()

		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic in TraceMiddleware: %v\n", r)
			}
		}()
		next.ServeHTTP(lrw, r.WithContext(ctx))

		duration := time.Since(start)

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.Int("http.status_code", lrw.StatusCode),
			attribute.String("http.client_ip", r.RemoteAddr),
			attribute.Float64("http.response_time", float64(duration.Milliseconds())),
		)

		fmt.Printf("TraceID: %s, SpanID: %s, Duration: %s, Status: %d\n", traceID, spanID, duration, lrw.StatusCode)
	})
}

func (l *Logger) log(severity Severity, ctx context.Context, msg string, err error, metadata RequestMetadata) {
	logCounter.WithLabelValues(string(severity)).Inc()
	metadata.Timestamp = time.Now()

	spanContext := _trace.SpanFromContext(ctx).SpanContext()
	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID().String()

	if err != nil {
		msg = fmt.Sprintf("%s: %v", msg, err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		metadata.Deadline = &deadline
	} else {
		metadata.Deadline = nil
	}

	logData, _ := json.Marshal(metadata)
	logEntry := fmt.Sprintf("%s: %s, metadata: %s", severity, msg, string(logData))

	logger := l.zapLogger.With(zap.String("trace_id", traceID), zap.String("span_id", spanID))

	switch severity {
	case SeverityError:
		logger.Error(logEntry)
	case SeverityWarning:
		logger.Warn(logEntry)
	case SeverityDebug:
		logger.Debug(logEntry)
	case SeverityCritical:
		logger.Fatal(logEntry)
	default:
		logger.Info(logEntry)
	}
}

func (l *Logger) Info(ctx context.Context, msg string, metadata RequestMetadata) {
	l.log(SeverityInfo, ctx, msg, nil, metadata)
}

func (l *Logger) Error(ctx context.Context, err error, metadata RequestMetadata) {
	errType, errCode := ExtractErrorDetails(err)
	metadata.ErrorCode = errCode
	metadata.ErrorType = errType

	l.log(SeverityError, ctx, "", err, metadata)
}

func (l *Logger) Warn(ctx context.Context, msg string, metadata RequestMetadata) {
	l.log(SeverityWarning, ctx, msg, nil, metadata)
}

func (l *Logger) Debug(ctx context.Context, msg string, metadata RequestMetadata) {
	l.log(SeverityDebug, ctx, msg, nil, metadata)
}

func (l *Logger) Fatal(ctx context.Context, err error, metadata RequestMetadata) {
	l.log(SeverityCritical, ctx, "", err, metadata)
}
