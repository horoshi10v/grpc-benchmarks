package main

import (
	"context"
	"log"

	pb "github.com/horoshi10v/grpc-benchmarks/proto"

	"github.com/Shopify/sarama"
)

type BrokerServiceServer struct {
	pb.UnimplementedBrokerServiceServer
	KafkaProducer sarama.SyncProducer
	KafkaConsumer sarama.ConsumerGroup
}

// Відправка повідомлення в брокер (Kafka)
func (s *BrokerServiceServer) SendToBroker(ctx context.Context, msg *pb.BrokerMessage) (*pb.BrokerAck, error) {
	log.Printf("Sending message to Kafka queue %s: %s", msg.Queue, msg.Content)
	message := &sarama.ProducerMessage{
		Topic: msg.Queue,
		Value: sarama.StringEncoder(msg.Content),
	}
	partition, offset, err := s.KafkaProducer.SendMessage(message)
	if err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		return &pb.BrokerAck{Status: "Error"}, err
	}
	log.Printf("Message stored in topic(%s)/partition(%d)/offset(%d)", msg.Queue, partition, offset)
	return &pb.BrokerAck{Status: "Message sent to broker"}, nil
}

// Отримання повідомлень з брокера
func (s *BrokerServiceServer) ReceiveFromBroker(req *pb.BrokerSubscription, stream pb.BrokerService_ReceiveFromBrokerServer) error {
	log.Printf("Client subscribed to Kafka queue: %s", req.Queue)
	consumer := &KafkaConsumer{
		ready:  make(chan bool),
		stream: stream,
	}
	ctx := stream.Context()
	go func() {
		for {
			if err := s.KafkaConsumer.Consume(ctx, []string{req.Queue}, consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready // Чекаємо, поки консюмер буде готовий
	<-ctx.Done()
	return ctx.Err()
}
