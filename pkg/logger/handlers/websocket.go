package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	_logger "distributed-logging-poc/pkg/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(ctx context.Context, w http.ResponseWriter, r *http.Request, logger *_logger.Logger) {
	requestId := _logger.GenerateRequestID("req")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(ctx, err, _logger.RequestMetadata{
			ServiceName: "websocket-service",
			RequestID:   requestId,
		})
		return
	}
	defer conn.Close()

	logger.Info(ctx, "WebSocket connection established", _logger.RequestMetadata{
		ServiceName: "websocket-service",
		RequestID:   requestId,
	})
}
