package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/IBM/sarama"
	"github.com/kabenari/log-insight/pkg/models"
)

const (
	KafkaTopic  = "logs.to.analyze"
	KafkaBroker = "localhost:9092"
)

func main() {
	fmt.Println("ai worker working")

	consumer, err := sarama.NewConsumer([]string{KafkaBroker}, nil)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(KafkaTopic, 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}

	defer partitionConsumer.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			var entry models.LogEntry
			json.Unmarshal(msg.Value, &entry)
			analyzeLog(entry)

		case <-signals:
			fmt.Println("woker shutting down")
			return
		}
	}
}

func analyzeLog(entry models.LogEntry) {
	fmt.Println("------------------------------------------------")
	fmt.Printf("Received Critical Log: %s\n", entry.Message)
	fmt.Println("Contacting OpenAI (Mocking latency)...")

	// Simulate API Latency
	time.Sleep(2 * time.Second)

	// Mock Analysis based on keywords
	analysis := "Unknown issue."
	if entry.Message == "Payment Gateway Timeout" {
		analysis = "Root Cause: The 3rd party API is not responding. Recommendation: Check outbound firewall rules or vendor status page."
	} else if entry.Message == "Null Pointer Exception" {
		analysis = "Root Cause: Uninitialized variable in UserStruct. Recommendation: Check line 42 in auth_service.go."
	}

	result := models.AIResult{
		OriginalLog: entry,
		Analysis:    analysis,
		Fixed:       false,
	}

	if err := saveInsight(result); err != nil {
		log.Printf("error saving to file :", err)
	} else {
		fmt.Println("saved in file")
	}

	fmt.Printf(">>> AI INSIGHT: %s\n", result.Analysis)
	fmt.Println("------------------------------------------------")
}

func saveInsight(result models.AIResult) error {
	f, err := os.OpenFile("insights.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}

	return nil
}
