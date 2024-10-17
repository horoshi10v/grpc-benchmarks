// async_client.go
package main

import (
	"context"
	"log"
	"time"

	pb "github.com/horoshi10v/grpc-benchmarks/proto"
	"google.golang.org/grpc"
)

func RunAsyncClient() {
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewSyncAsyncServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.Request{
		Id:      2,
		Payload: "Hello, asynchronous world!",
	}

	log.Println("Sending asynchronous request...")
	start := time.Now()
	response, err := client.AsynchronousMethod(ctx, request)
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("Could not get response: %v", err)
	}
	log.Printf("Received response: %v", response)
	log.Printf("Time taken: %v", elapsed)

	// Тут можна реалізувати логіку очікування результату асинхронної обробки, якщо це необхідно
}
