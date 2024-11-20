
# Distributed Logging Testing Guide

This guide helps you test distributed logging in your system, covering HTTP, gRPC, WebSocket, and GraphQL.

---

## 1. **HTTP Testing**

### Test with `curl`:
```bash
curl -X GET http://localhost:8080
```

### Expected Result:
- Response in terminal:
  ```
  Hello, HTTP Logging!
  ```
- Logs in console and collector (e.g., GCP):
  ```json
  {
    "trace_id": "123abc456def",
    "span_id": "789ghi",
    "request_id": "req-123",
    "service_name": "http-service",
    "message": "HTTP request received",
    "timestamp": "2024-11-14T10:00:00Z"
  }
  ```

---

## 2. **gRPC Testing**

### Tool: `grpcurl`

Test the `SayHello` method in the gRPC service:
```bash
grpcurl -plaintext -d '{"name": "World"}' localhost:50051 handlers.SimulatedService/SayHello
```

### Expected Result:
- Response in terminal:
  ```json
  {
    "message": "Hello, World! This is gRPC Logging!"
  }
  ```
- Logs in console and collector (e.g., GCP):
  ```json
  {
    "trace_id": "abc123",
    "span_id": "xyz456",
    "request_id": "req-456",
    "service_name": "grpc-service",
    "message": "Received gRPC request: World",
    "timestamp": "2024-11-14T10:00:00Z"
  }
  ```

---

## 3. **WebSocket Testing**

### Tool: WebSocket Client
1. Install **Simple WebSocket Client** in Chrome/Edge.
2. Connect to:
   ```
   ws://localhost:8080/ws
   ```
3. Send a message:
   ```
   Hello WebSocket!
   ```

### Expected Result:
- Logs in console and collector (e.g., GCP):
  ```json
  {
    "trace_id": "456ghi789jkl",
    "span_id": "xyz987",
    "service_name": "websocket-service",
    "message": "WebSocket connection established",
    "timestamp": "2024-11-14T10:00:00Z"
  }
  {
    "trace_id": "456ghi789jkl",
    "span_id": "xyz987",
    "service_name": "websocket-service",
    "message": "Hello WebSocket!",
    "timestamp": "2024-11-14T10:00:01Z"
  }
  ```

---

## 4. **GraphQL Testing**

### GraphQL Playground:
1. Access the playground:
   ```
   http://localhost:8080/playground
   ```
2. Execute the query:
   ```graphql
   query {
     sayHello
   }
   ```

### Expected Result:
- Playground response:
  ```json
  {
    "data": {
      "sayHello": "Hello, GraphQL Logging!"
    }
  }
  ```
- Logs in console and collector (e.g., GCP):
  ```json
  {
    "trace_id": "789jkl012mno",
    "span_id": "opq345",
    "service_name": "graphql-service",
    "message": "GraphQL query received",
    "timestamp": "2024-11-14T10:00:00Z"
  }
  ```

---

## 5. **Check Logs in GCP**

If GCP Logging is enabled:

1. Access the **Google Cloud Console**.
2. Navigate to **Logging**:
   ```
   https://console.cloud.google.com/logs
   ```
3. Filter logs by `service_name` or other fields like `trace_id`.

---

## 6. **Traceability Testing**

1. **Correlate Logs**:
   Ensure logs share the same `trace_id` across different services (e.g., HTTP and gRPC).

2. **Simulate a Full Workflow**:
   - Make an HTTP request that triggers a gRPC service internally.
   - Verify that the `trace_id` is propagated end-to-end.

---

## Conclusion

This guide ensures that distributed logging is functional, and logs include:
- `trace_id`
- `span_id`
- `request_id`
- `service_name`

Feel free to expand tests for additional workflows or services.
