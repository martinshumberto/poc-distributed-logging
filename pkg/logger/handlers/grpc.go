package handlers

import (
	"context"
	"fmt"
	"net"

	"distributed-logging-poc/pkg/logger"
	_grpc "distributed-logging-poc/pkg/logger/grpc"

	"google.golang.org/grpc"
)

type SimulatedService struct {
	Logger *logger.Logger
	_grpc.UnimplementedSimulatedServiceServer
}

func (s *SimulatedService) SayHello(ctx context.Context, req *_grpc.HelloRequest) (*_grpc.HelloResponse, error) {
	traceID := ctx.Value("trace_id")
	spanID := ctx.Value("span_id")

	s.Logger.Info(ctx, fmt.Sprintf("Received gRPC request: %s", req.GetName()), logger.RequestMetadata{
		TraceID:     fmt.Sprintf("%v", traceID),
		SpanID:      fmt.Sprintf("%v", spanID),
		ServiceName: "grpc-service",
		RequestID:   logger.GenerateRequestID("req"),
	})

	return &_grpc.HelloResponse{Message: fmt.Sprintf("Hello, %s! This is gRPC Logging!", req.GetName())}, nil
}

func RegisterGRPCServices(server *grpc.Server, logService *logger.Logger) {
	simulatedService := &SimulatedService{Logger: logService}
	_grpc.RegisterSimulatedServiceServer(server, simulatedService)
}

func GetGRPCListener() (net.Listener, error) {
	return net.Listen("tcp", ":50051")
}
