package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	grpcAddr := flag.String("grpc-addr", "", "gRPC backend address (e.g., localhost:50051)")
	httpPort := flag.Int("http-port", 8080, "HTTP server port")
	protoPath := flag.String("proto", "", "Path to .proto files (optional if reflection enabled)")

	flag.Parse()

	if *grpcAddr == "" {
		fmt.Fprintf(os.Stderr, "Error: --grpc-addr is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Starting gRPC-HTTP bridge...")
	log.Printf("  gRPC backend: %s", *grpcAddr)
	log.Printf("  HTTP port: %d", *httpPort)
	if *protoPath != "" {
		log.Printf("  Proto files: %s", *protoPath)
	} else {
		log.Printf("  Using gRPC reflection (no proto files)")
	}

	// TODO: Implement bridge logic
	log.Fatal("Bridge implementation coming soon...")
}
