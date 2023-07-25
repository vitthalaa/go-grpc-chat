# Simple Go GRPC stream chat server

1. Run server: `go run server/main.go`

### Generate Proto
```shell
protoc -I ./proto \
  --go_out ./gen/go --go_opt paths=source_relative \
  --go-grpc_out ./gen/go --go-grpc_opt paths=source_relative \
  ./proto/chat/v1/chat.proto
```
