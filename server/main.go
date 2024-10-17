package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// Import the generated code from the .proto file
	pb "github.com/horoshi10v/grpc-benchmarks/proto"
)

func main() {
	// Get the absolute path to the Kafka config file
	absPath, err := filepath.Abs("configs/kafka_config.yaml")
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Load the Kafka configuration
	kafkaConfig, err := loadKafkaConfig(absPath)
	if err != nil {
		log.Fatalf("Failed to load Kafka config: %v", err)
	}

	// Initialize Kafka producer and consumer from the configuration
	producer := InitializeKafkaProducer(kafkaConfig.Brokers)
	defer producer.Close()

	consumer := InitializeKafkaConsumer(kafkaConfig.Brokers, kafkaConfig.Consumer.GroupID)
	defer consumer.Close()

	// Create gRPC server
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterSyncAsyncServiceServer(grpcServer, &SyncAsyncServiceServer{})
	pb.RegisterPubSubServiceServer(grpcServer, NewPubSubServiceServer())
	pb.RegisterBrokerServiceServer(grpcServer, &BrokerServiceServer{
		KafkaProducer: producer,
		KafkaConsumer: consumer,
	})

	// Add reflection for gRPC
	reflection.Register(grpcServer)

	// Handle termination signals
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	log.Println("Server is running on port 50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
