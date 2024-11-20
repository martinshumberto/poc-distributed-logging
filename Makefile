gen-proto:
	protoc --go_out=. --go-grpc_out=. pkg/logger/grpc/*.proto

start:
	CGO_ENABLED=0 go run ./cmd/main.go