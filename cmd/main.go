package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"

	"distributed-logging-poc/pkg/logger"
	"distributed-logging-poc/pkg/logger/handlers"
)

func main() {
	ctx := context.Background()
	serviceName := "distributed-logging-poc"
	environment := os.Getenv("ENVIRONMENT")

	// Initialize OpenTelemetry tracer
	tp, err := logger.InitTracer(serviceName, environment)
	if err != nil {
		log.Fatalf("Error initializing tracer: %v", err)
	}

	defer func() { _ = tp.Shutdown(ctx) }()

	// Initialize Logger
	logService, err := logger.NewLogger(ctx, logger.Config{
		ServiceName: serviceName,
		Environment: environment,
	})
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}

	// Setup HTTP server
	httpServer := &http.Server{
		Addr: ":8080",
		Handler: logger.TraceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.HandleHTTP(r.Context(), w, r, logService)
		})),
	}

	// Setup gRPC server
	grpcServer := grpc.NewServer()
	handlers.RegisterGRPCServices(grpcServer, logService)

	go func() {
		listener, err := handlers.GetGRPCListener()
		if err != nil {
			log.Fatalf("Failed to get gRPC listener: %v", err)
		}

		log.Println("gRPC server started on :50051")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Setup WebSocket handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleWebSocket(ctx, w, r, logService)
	})

	// Start HTTP server
	log.Println("HTTP server started on :8080")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
