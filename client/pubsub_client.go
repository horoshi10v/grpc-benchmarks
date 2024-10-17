// pubsub_client.go
package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/horoshi10v/grpc-benchmarks/proto"
	"google.golang.org/grpc"
)

func RunPubSubPublisher(topic, message string) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewPubSubServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pubMessage := &pb.PubSubMessage{
		Topic:   topic,
		Content: message,
	}

	log.Printf("Publishing message to topic '%s': %s", topic, message)
	response, err := client.Publish(ctx, pubMessage)
	if err != nil {
		log.Fatalf("Could not publish message: %v", err)
	}
	log.Printf("Received acknowledgment: %v", response)
}

func RunPubSubSubscriber(topic string) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewPubSubServiceClient(conn)

	ctx := context.Background()
	subRequest := &pb.SubscriptionRequest{
		Topic: topic,
	}

	log.Printf("Subscribing to topic '%s'", topic)
	stream, err := client.Subscribe(ctx, subRequest)
	if err != nil {
		log.Fatalf("Could not subscribe: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Println("Stream closed by server")
			break
		}
		if err != nil {
			log.Fatalf("Error receiving message: %v", err)
		}
		log.Printf("Received message: %s", msg.Content)
	}
}
