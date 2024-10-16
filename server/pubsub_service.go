// server/pubsub_service.go
package main

import (
	"context"
	"log"
	"sync"

	pb "github.com/horoshi10v/grpc-benchmarks/proto"
)

type PubSubServiceServer struct {
	pb.UnimplementedPubSubServiceServer
	subscribers sync.Map // Ключ: тема, Значення: список каналів
}

func NewPubSubServiceServer() *PubSubServiceServer {
	return &PubSubServiceServer{}
}

// Метод публікації
func (s *PubSubServiceServer) Publish(ctx context.Context, msg *pb.PubSubMessage) (*pb.PubSubAck, error) {
	log.Printf("Publishing message to topic %s: %s", msg.Topic, msg.Content)
	if subs, ok := s.subscribers.Load(msg.Topic); ok {
		channels := subs.([]chan *pb.PubSubMessage)
		for _, ch := range channels {
			ch <- msg
		}
	}
	return &pb.PubSubAck{Status: "Message published"}, nil
}

// Метод підписки
func (s *PubSubServiceServer) Subscribe(req *pb.SubscriptionRequest, stream pb.PubSubService_SubscribeServer) error {
	log.Printf("Client subscribed to topic: %s", req.Topic)
	msgCh := make(chan *pb.PubSubMessage, 10)

	// Додаємо підписника
	var channels []chan *pb.PubSubMessage
	if subs, ok := s.subscribers.Load(req.Topic); ok {
		channels = subs.([]chan *pb.PubSubMessage)
	}
	channels = append(channels, msgCh)
	s.subscribers.Store(req.Topic, channels)

	// Видаляємо підписника при завершенні
	defer func() {
		if subs, ok := s.subscribers.Load(req.Topic); ok {
			channels := subs.([]chan *pb.PubSubMessage)
			for i, ch := range channels {
				if ch == msgCh {
					channels = append(channels[:i], channels[i+1:]...)
					break
				}
			}
			s.subscribers.Store(req.Topic, channels)
		}
		close(msgCh)
		log.Printf("Client unsubscribed from topic: %s", req.Topic)
	}()

	// Відправляємо повідомлення клієнту
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg := <-msgCh:
			if err := stream.Send(msg); err != nil {
				return err
			}
		}
	}
}
