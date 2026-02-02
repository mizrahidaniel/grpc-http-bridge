package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Bridge struct {
	grpcConn   *grpc.ClientConn
	grpcAddr   string
	httpPort   int
	reflClient grpc_reflection_v1alpha.ServerReflectionClient
}

func main() {
	grpcAddr := flag.String("grpc-addr", "", "gRPC backend address (e.g., localhost:50051)")
	httpPort := flag.Int("http-port", 8080, "HTTP server port")
	flag.Parse()

	if *grpcAddr == "" {
		fmt.Fprintf(os.Stderr, "Error: --grpc-addr is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	bridge, err := NewBridge(*grpcAddr, *httpPort)
	if err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}
	defer bridge.Close()

	log.Printf("Starting gRPC-HTTP bridge...")
	log.Printf("  gRPC backend: %s", *grpcAddr)
	log.Printf("  HTTP server: http://localhost:%d", *httpPort)

	if err := bridge.Serve(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func NewBridge(grpcAddr string, httpPort int) (*Bridge, error) {
	// Connect to gRPC backend
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC backend: %w", err)
	}

	// Create reflection client
	reflClient := grpc_reflection_v1alpha.NewServerReflectionClient(conn)

	return &Bridge{
		grpcConn:   conn,
		grpcAddr:   grpcAddr,
		httpPort:   httpPort,
		reflClient: reflClient,
	}, nil
}

func (b *Bridge) Close() {
	if b.grpcConn != nil {
		b.grpcConn.Close()
	}
}

func (b *Bridge) Serve() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "ok",
			"grpc_addr":   b.grpcAddr,
			"reflection":  true,
			"timestamp":   time.Now().Unix(),
		})
	})

	// Main RPC handler: POST /{service}/{method}
	r.Post("/*", b.handleRPC)

	addr := fmt.Sprintf(":%d", b.httpPort)
	log.Printf("✓ Bridge ready - listening on %s", addr)
	log.Printf("  Example: curl http://localhost:%d/health", b.httpPort)
	log.Printf("  RPC format: curl http://localhost:%d/{service}/{method} -d '{...}'", b.httpPort)

	return http.ListenAndServe(addr, r)
}

func (b *Bridge) handleRPC(w http.ResponseWriter, r *http.Request) {
	// Extract service/method from URL
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid path format. Use: /{service}/{method}", http.StatusBadRequest)
		return
	}

	service := parts[0]
	method := parts[1]
	fullMethod := fmt.Sprintf("/%s/%s", service, method)

	log.Printf("→ RPC call: %s", fullMethod)

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read body: %v", err), http.StatusBadRequest)
		return
	}

	// For now, return a placeholder response with reflection info
	// Full dynamic invocation requires method descriptor resolution via reflection
	response := map[string]interface{}{
		"status": "bridge_working",
		"note":   "Full dynamic invocation coming in next iteration",
		"request": map[string]interface{}{
			"service": service,
			"method":  method,
			"body":    string(body),
		},
		"next_steps": []string{
			"1. Resolve service descriptor via reflection",
			"2. Parse method input/output types",
			"3. Unmarshal JSON → protobuf",
			"4. Invoke gRPC method dynamically",
			"5. Marshal protobuf → JSON response",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("✓ Response sent (placeholder)")
}

// invokeRPC performs dynamic gRPC invocation (next PR)
func (b *Bridge) invokeRPC(ctx context.Context, fullMethod string, reqJSON []byte) ([]byte, error) {
	// TODO: Implement dynamic invocation using grpc/reflection
	// Steps:
	// 1. Use b.reflClient.ServerReflectionInfo() to get service descriptors
	// 2. Find method descriptor by name
	// 3. Get input/output message types from descriptor
	// 4. Create dynamic message: dynamicpb.NewMessage(inputDesc)
	// 5. Unmarshal JSON into dynamic message: protojson.Unmarshal(reqJSON, dynamicMsg)
	// 6. Invoke: grpc.Invoke(ctx, fullMethod, dynamicMsg, respMsg, b.grpcConn)
	// 7. Marshal response: protojson.Marshal(respMsg)

	return nil, fmt.Errorf("not implemented yet")
}

// Helper: convert protobuf Message to JSON
func messageToJSON(msg proto.Message) ([]byte, error) {
	marshaler := protojson.MarshalOptions{
		Indent:          "  ",
		EmitUnpopulated: true,
	}
	return marshaler.Marshal(msg)
}

// Helper: convert JSON to protobuf Message
func jsonToMessage(data []byte, msgDesc protoreflect.MessageDescriptor) (*dynamicpb.Message, error) {
	msg := dynamicpb.NewMessage(msgDesc)
	if err := protojson.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}
