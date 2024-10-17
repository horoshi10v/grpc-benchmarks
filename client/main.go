// main.go
package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	// Прапорці для вибору режиму роботи
	mode := flag.String("mode", "sync", "Mode: sync | async | pubsub | broker")
	action := flag.String("action", "", "Action: publish | subscribe (for pubsub and broker modes)")
	topicOrQueue := flag.String("topic", "", "Topic or Queue name (for pubsub and broker modes)")
	message := flag.String("message", "", "Message to send (for publish actions)")
	flag.Parse()

	switch *mode {
	case "sync":
		RunSyncClient()
	case "async":
		RunAsyncClient()
	case "pubsub":
		if *action == "publish" {
			if *topicOrQueue == "" || *message == "" {
				log.Fatal("Please provide topic and message for publishing")
			}
			RunPubSubPublisher(*topicOrQueue, *message)
		} else if *action == "subscribe" {
			if *topicOrQueue == "" {
				log.Fatal("Please provide topic for subscription")
			}
			RunPubSubSubscriber(*topicOrQueue)
		} else {
			log.Fatal("Invalid action for pubsub mode. Use 'publish' or 'subscribe'")
		}
	case "broker":
		if *action == "publish" {
			if *topicOrQueue == "" || *message == "" {
				log.Fatal("Please provide queue and message for publishing")
			}
			RunBrokerPublisher(*topicOrQueue, *message)
		} else if *action == "subscribe" {
			if *topicOrQueue == "" {
				log.Fatal("Please provide queue for subscription")
			}
			RunBrokerSubscriber(*topicOrQueue)
		} else {
			log.Fatal("Invalid action for broker mode. Use 'publish' or 'subscribe'")
		}
	default:
		fmt.Println("Invalid mode. Use 'sync', 'async', 'pubsub', or 'broker'")
	}
}
