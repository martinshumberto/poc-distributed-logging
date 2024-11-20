package handlers

import (
	"context"

	"distributed-logging-poc/pkg/logger"
)

type Resolver struct {
	Logger *logger.Logger
}

func (r *Resolver) Query_SayHello(ctx context.Context) (string, error) {
	r.Logger.Info(ctx, "GraphQL query received", logger.RequestMetadata{
		ServiceName: "graphql-service",
		RequestID:   logger.GenerateRequestID("req"),
	})
	return "Hello, GraphQL Logging!", nil
}
