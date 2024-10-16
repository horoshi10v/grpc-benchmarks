// server/kafka_producer.go
package server

import (
	"log"

	"github.com/Shopify/sarama"
)

// Ініціалізація Kafka продюсера
func InitializeKafkaProducer(brokers []string) sarama.SyncProducer {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to start Kafka producer: %v", err)
	}
	return producer
}
