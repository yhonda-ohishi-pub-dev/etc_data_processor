package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/yhonda-ohishi/etc_data_processor/src/proto"
	"github.com/yhonda-ohishi/etc_data_processor/src/pkg/handler"
	"github.com/yhonda-ohishi/etc_data_processor/src/pkg/db"
	"github.com/yhonda-ohishi/etc_data_processor/src/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port       = flag.Int("port", 50051, "The server port")
	dbAddr     = flag.String("db", "", "Database service address")
	configFile = flag.String("config", "", "Config file path")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override with command line flags if provided
	if *port != 50051 {
		cfg.Port = *port
	}
	if *dbAddr != "" {
		cfg.DBServiceAddr = *dbAddr
	}

	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Create DB client
	var dbClient handler.DBClient
	if cfg.DBServiceAddr != "" {
		log.Printf("Connecting to db_service at: %s", cfg.DBServiceAddr)
		client, err := db.NewETCMeisaiClient(cfg.DBServiceAddr)
		if err != nil {
			log.Printf("Warning: Failed to connect to db_service: %v", err)
			log.Printf("Continuing without database integration...")
		} else {
			dbClient = client
			log.Printf("Successfully connected to db_service")
			defer client.Close()
		}
	} else {
		log.Printf("No db_service address configured - running without database integration")
	}

	// Register service
	service := handler.NewDataProcessorService(dbClient)
	pb.RegisterDataProcessorServiceServer(grpcServer, service)

	// Register reflection service for grpcurl
	reflection.Register(grpcServer)

	// Start server in goroutine
	go func() {
		log.Printf("Starting gRPC server on port %d...", cfg.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down server...")
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}

func loadConfig(configFile string) (*config.Config, error) {
	// Default configuration
	cfg := &config.Config{
		Port:          50051,
		DBServiceAddr: "",
		MaxBatchSize:  100,
		ValidateData:  true,
	}

	// If config file specified, load it
	if configFile != "" {
		fileCfg, err := config.LoadFromFile(configFile)
		if err != nil {
			return nil, err
		}
		cfg = fileCfg
	}

	// Environment variables override file config
	if port := os.Getenv("ETC_PROCESSOR_PORT"); port != "" {
		var p int
		fmt.Sscanf(port, "%d", &p)
		if p > 0 {
			cfg.Port = p
		}
	}

	if dbAddr := os.Getenv("ETC_PROCESSOR_DB_ADDR"); dbAddr != "" {
		cfg.DBServiceAddr = dbAddr
	}

	return cfg, nil
}