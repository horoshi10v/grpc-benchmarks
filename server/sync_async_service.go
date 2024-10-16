package server

import (
	"context"
	"log"
	"time"

	pb "github.com/horoshi10v/grpc-benchmarks/proto"
)

type SyncAsyncServiceServer struct {
	pb.UnimplementedSyncAsyncServiceServer
}

// Синхронний метод
func (s *SyncAsyncServiceServer) SynchronousMethod(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	log.Printf("Received synchronous request: %v", req)
	// Обробка запиту
	result := "Processed: " + req.Payload
	return &pb.Response{
		Id:     req.Id,
		Result: result,
	}, nil
}

// Асинхронний метод
func (s *SyncAsyncServiceServer) AsynchronousMethod(ctx context.Context, req *pb.Request) (*pb.AsyncResponse, error) {
	log.Printf("Received asynchronous request: %v", req)
	// Емуляція асинхронної обробки
	go func(r *pb.Request) {
		time.Sleep(5 * time.Second)
		log.Printf("Asynchronously processed request ID %d", r.Id)
		// Тут можна зберегти результат або відправити його кудись
	}(req)
	return &pb.AsyncResponse{
		Id:     req.Id,
		Status: "Processing started",
	}, nil
}
