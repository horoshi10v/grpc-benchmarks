// server/kafka_consumer.go
package server

import (
	"log"

	"github.com/Shopify/sarama"
	pb "github.com/horoshi10v/grpc-benchmarks/proto
)

// Ініціалізація Kafka консюмера
func InitializeKafkaConsumer(brokers []string, groupID string) sarama.ConsumerGroup {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		log.Fatalf("Error creating consumer group client: %v", err)
	}
	return consumerGroup
}

// KafkaConsumer реалізує інтерфейс консюмера
type KafkaConsumer struct {
	ready  chan bool
	stream pb.BrokerService_ReceiveFromBrokerServer
}

// Setup викликається на початку сесії
func (consumer *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

// Cleanup викликається в кінці сесії
func (consumer *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim обробляє повідомлення
func (consumer *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("Received message: %s", string(msg.Value))
		brokerMsg := &pb.BrokerMessage{
			Queue:   msg.Topic,
			Content: string(msg.Value),
		}
		if err := consumer.stream.Send(brokerMsg); err != nil {
			return err
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
