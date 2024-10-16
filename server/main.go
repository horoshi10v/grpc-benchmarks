package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// Імпортуємо згенерований код з .proto файлу
	pb "github.com/horoshi10v/grpc-benchmarks/proto"
	// Імпортуємо наші сервіси
	"github.com/horoshi10v/grpc-benchmarks/server"
)

func main() {
	// Налаштування Kafka
	brokers := []string{"localhost:9092"}
	producer := server.InitializeKafkaProducer(brokers)
	defer producer.Close()

	consumerGroupID := "grpc-broker-service-group"
	consumer := server.InitializeKafkaConsumer(brokers, consumerGroupID)
	defer consumer.Close()

	// Створення gRPC сервера
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	// Реєстрація сервісів
	pb.RegisterSyncAsyncServiceServer(grpcServer, &server.SyncAsyncServiceServer{})
	pb.RegisterPubSubServiceServer(grpcServer, server.NewPubSubServiceServer())
	pb.RegisterBrokerServiceServer(grpcServer, &server.BrokerServiceServer{
		KafkaProducer: producer,
		KafkaConsumer: consumer,
	})

	// Додаємо відображення для gRPC
	reflection.Register(grpcServer)

	// Обробка сигналів завершення
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	log.Println("Server is running on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
