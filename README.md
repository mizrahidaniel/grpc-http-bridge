# gRPC-HTTP Bridge

Expose gRPC services as REST APIs with zero configuration.

## Quick Start

```bash
# Start your gRPC service with reflection enabled
go run ./server.go

# Start the bridge
grpc-http-bridge --grpc-addr localhost:50051 --http-port 8080

# Call gRPC methods via REST
curl http://localhost:8080/api.v1.UserService/GetUser \
  -d '{"user_id": "123"}'
```

## Status

🚧 **In Development** - Core bridge implementation in progress

**Next Steps:**
1. Basic CLI argument parsing
2. gRPC connection + reflection client
3. HTTP server with JSON ↔ Protobuf translation
4. Health check endpoint

## Why?

gRPC is great for services. HTTP is great for browsers. This bridges the gap:

- ✅ No proto files needed (uses gRPC reflection)
- ✅ Standard HTTP (works with curl, fetch, Postman)
- ✅ Zero config for simple cases
- ✅ Single binary, no dependencies

## Architecture

```
HTTP Request (JSON)
     ↓
HTTP Server (FastAPI/Chi)
     ↓
JSON → Protobuf Translation
     ↓
gRPC Client
     ↓
Your gRPC Service
```

## Comparison

**vs Envoy:** Simpler setup, no config files required  
**vs gRPC-Web:** Standard HTTP, no special client library  
**vs Duplicate REST:** Single source of truth (proto definitions)

## Development

```bash
# Clone
git clone https://github.com/mizrahidaniel/grpc-http-bridge
cd grpc-http-bridge

# Build
go build -o grpc-http-bridge ./cmd/bridge

# Run
./grpc-http-bridge --help
```

## License

MIT
