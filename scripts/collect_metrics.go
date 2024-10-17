// collect_metrics.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"encoding/csv"
	"strconv"

	pb "github.com/horoshi10v/grpc-benchmarks/proto"
	"google.golang.org/grpc"
)

func main() {
	outputFile := flag.String("output", "metrics.csv", "Output CSV file for metrics")
	serverAddress := flag.String("server", "localhost:50051", "gRPC server address")
	mode := flag.String("mode", "sync", "Mode: sync | async | pubsub | broker")
	action := flag.String("action", "", "Action: publish | subscribe (for pubsub and broker modes)")
	topicOrQueue := flag.String("topic", "", "Topic or Queue name (for pubsub and broker modes)")
	requests := flag.Int("requests", 100, "Number of requests")
	concurrency := flag.Int("concurrency", 1, "Number of concurrent goroutines")
	flag.Parse()

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()

	// Відкриваємо файл для запису метрик
	file, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Could not create metrics file: %v", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Записуємо заголовки стовпців
	writer.Write([]string{"Timestamp", "ResponseTime(ms)", "NumGoroutines", "Alloc(MB)", "TotalAlloc(MB)", "Sys(MB)", "NumGC"})

	var wg sync.WaitGroup
	requestsPerGoroutine := *requests / *concurrency

	switch *mode {
	case "sync":
		client := pb.NewSyncAsyncServiceClient(conn)
		for i := 0; i < *concurrency; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				collectSyncMetrics(client, id, requestsPerGoroutine, writer)
			}(i)
		}
	case "async":
		client := pb.NewSyncAsyncServiceClient(conn)
		for i := 0; i < *concurrency; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				collectAsyncMetrics(client, id, requestsPerGoroutine, writer)
			}(i)
		}
	case "pubsub":
		if *action == "subscribe" {
			client := pb.NewPubSubServiceClient(conn)
			collectPubSubMetrics(client, *topicOrQueue, writer)
		}
	case "broker":
		if *action == "subscribe" {
			client := pb.NewBrokerServiceClient(conn)
			collectBrokerMetrics(client, *topicOrQueue, writer)
		}
	default:
		log.Fatalf("Invalid mode: %s", *mode)
	}

	wg.Wait()
}

func collectSyncMetrics(client pb.SyncAsyncServiceClient, id, numRequests int, writer *csv.Writer) {
	for i := 0; i < numRequests; i++ {
		startTime := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		request := &pb.Request{
			Id:      int32(id*numRequests + i),
			Payload: fmt.Sprintf("Request %d from goroutine %d", i, id),
		}

		_, err := client.SynchronousMethod(ctx, request)
		responseTime := time.Since(startTime).Milliseconds()

		// Збираємо метрики використання пам'яті
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		allocMB := float64(memStats.Alloc) / 1024.0 / 1024.0
		totalAllocMB := float64(memStats.TotalAlloc) / 1024.0 / 1024.0
		sysMB := float64(memStats.Sys) / 1024.0 / 1024.0

		numGoroutines := runtime.NumGoroutine()

		// Записуємо метрики в CSV
		record := []string{
			time.Now().Format(time.RFC3339),
			strconv.FormatInt(responseTime, 10),
			strconv.Itoa(numGoroutines),
			strconv.FormatFloat(allocMB, 'f', 2, 64),
			strconv.FormatFloat(totalAllocMB, 'f', 2, 64),
			strconv.FormatFloat(sysMB, 'f', 2, 64),
			strconv.FormatUint(uint64(memStats.NumGC), 10),
		}
		writer.Write(record)
		writer.Flush()

		if err != nil {
			log.Printf("Error in synchronous request: %v", err)
		}

		// Опціональна затримка між запитами
		// time.Sleep(10 * time.Millisecond)
	}
}

func collectAsyncMetrics(client pb.SyncAsyncServiceClient, id, numRequests int, writer *csv.Writer) {
	for i := 0; i < numRequests; i++ {
		startTime := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		request := &pb.Request{
			Id:      int32(id*numRequests + i),
			Payload: fmt.Sprintf("Request %d from goroutine %d", i, id),
		}

		_, err := client.AsynchronousMethod(ctx, request)
		responseTime := time.Since(startTime).Milliseconds()

		// Збираємо метрики використання пам'яті
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		allocMB := float64(memStats.Alloc) / 1024.0 / 1024.0
		totalAllocMB := float64(memStats.TotalAlloc) / 1024.0 / 1024.0
		sysMB := float64(memStats.Sys) / 1024.0 / 1024.0

		numGoroutines := runtime.NumGoroutine()

		// Записуємо метрики в CSV
		record := []string{
			time.Now().Format(time.RFC3339),
			strconv.FormatInt(responseTime, 10),
			strconv.Itoa(numGoroutines),
			strconv.FormatFloat(allocMB, 'f', 2, 64),
			strconv.FormatFloat(totalAllocMB, 'f', 2, 64),
			strconv.FormatFloat(sysMB, 'f', 2, 64),
			strconv.FormatUint(uint64(memStats.NumGC), 10),
		}
		writer.Write(record)
		writer.Flush()

		if err != nil {
			log.Printf("Error in asynchronous request: %v", err)
		}

		// Опціональна затримка між запитами
		// time.Sleep(10 * time.Millisecond)
	}
}

func collectPubSubMetrics(client pb.PubSubServiceClient, topic string, writer *csv.Writer) {
	ctx := context.Background()
	subRequest := &pb.SubscriptionRequest{
		Topic: topic,
	}

	stream, err := client.Subscribe(ctx, subRequest)
	if err != nil {
		log.Fatalf("Could not subscribe: %v", err)
	}

	for {
		startTime := time.Now()
		msg, err := stream.Recv()
		responseTime := time.Since(startTime).Milliseconds()

		if err != nil {
			log.Printf("Error receiving message: %v", err)
			break
		}

		// Збираємо метрики використання пам'яті
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		allocMB := float64(memStats.Alloc) / 1024.0 / 1024.0
		totalAllocMB := float64(memStats.TotalAlloc) / 1024.0 / 1024.0
		sysMB := float64(memStats.Sys) / 1024.0 / 1024.0

		numGoroutines := runtime.NumGoroutine()

		// Записуємо метрики в CSV
		record := []string{
			time.Now().Format(time.RFC3339),
			strconv.FormatInt(responseTime, 10),
			strconv.Itoa(numGoroutines),
			strconv.FormatFloat(allocMB, 'f', 2, 64),
			strconv.FormatFloat(totalAllocMB, 'f', 2, 64),
			strconv.FormatFloat(sysMB, 'f', 2, 64),
			strconv.FormatUint(uint64(memStats.NumGC), 10),
		}
		writer.Write(record)
		writer.Flush()

		log.Printf("Received message: %s", msg.Content)
	}
}

func collectBrokerMetrics(client pb.BrokerServiceClient, queue string, writer *csv.Writer) {
	ctx := context.Background()
	subscription := &pb.BrokerSubscription{
		Queue: queue,
	}

	stream, err := client.ReceiveFromBroker(ctx, subscription)
	if err != nil {
		log.Fatalf("Could not subscribe to broker: %v", err)
	}

	for {
		startTime := time.Now()
		msg, err := stream.Recv()
		responseTime := time.Since(startTime).Milliseconds()

		if err != nil {
			log.Printf("Error receiving message from broker: %v", err)
			break
		}

		// Збираємо метрики використання пам'яті
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		allocMB := float64(memStats.Alloc) / 1024.0 / 1024.0
		totalAllocMB := float64(memStats.TotalAlloc) / 1024.0 / 1024.0
		sysMB := float64(memStats.Sys) / 1024.0 / 1024.0

		numGoroutines := runtime.NumGoroutine()

		// Записуємо метрики в CSV
		record := []string{
			time.Now().Format(time.RFC3339),
			strconv.FormatInt(responseTime, 10),
			strconv.Itoa(numGoroutines),
			strconv.FormatFloat(allocMB, 'f', 2, 64),
			strconv.FormatFloat(totalAllocMB, 'f', 2, 64),
			strconv.FormatFloat(sysMB, 'f', 2, 64),
			strconv.FormatUint(uint64(memStats.NumGC), 10),
		}
		writer.Write(record)
		writer.Flush()

		log.Printf("Received message from broker: %s", msg.Content)
	}
}
