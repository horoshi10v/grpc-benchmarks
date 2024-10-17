// broker_client.go
package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/horoshi10v/grpc-benchmarks/proto"
	"google.golang.org/grpc"
)

func RunBrokerPublisher(queue, message string) {
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewBrokerServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	brokerMessage := &pb.BrokerMessage{
		Queue:   queue,
		Content: message,
	}

	log.Printf("Sending message to queue '%s': %s", queue, message)
	response, err := client.SendToBroker(ctx, brokerMessage)
	if err != nil {
		log.Fatalf("Could not send message to broker: %v", err)
	}
	log.Printf("Received acknowledgment: %v", response)
}

func RunBrokerSubscriber(queue string) {
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewBrokerServiceClient(conn)

	ctx := context.Background()
	subscription := &pb.BrokerSubscription{
		Queue: queue,
	}

	log.Printf("Subscribing to queue '%s'", queue)
	stream, err := client.ReceiveFromBroker(ctx, subscription)
	if err != nil {
		log.Fatalf("Could not subscribe to broker: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Println("Stream closed by server")
			break
		}
		if err != nil {
			log.Fatalf("Error receiving message from broker: %v", err)
		}
		log.Printf("Received message from broker: %s", msg.Content)
	}
}
