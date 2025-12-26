package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/kabenari/log-insight/pkg/models"
)

const (
	KafkaTopic  = "logs.to.analyze"
	KafkaBroker = "localhost:9092"
)

func main() {
	producer, err := setupProducer()
	if err != nil {
		log.Fatal("failed to setup producer")
	}

	defer producer.Close()

	go startLogGenerator(producer)

	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		var entry models.LogEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		processLog(producer, entry)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/webhook/datadog", func(w http.ResponseWriter, r *http.Request) {
		var ddPayload struct {
			Title string `json:"title"`
			Body  string `json:"body"`
			Date  int64  `json:"date"`
		}

		if err := json.NewDecoder(r.Body).Decode(&ddPayload); err != nil {
			http.Error(w, "invalid datadog json", http.StatusBadRequest)
			return
		}

		entry := models.LogEntry{
			Timestamp: time.Unix(ddPayload.Date/1000, 0),
			Level:     "ERROR",
			Service:   "External-Datadog",
			Message:   fmt.Sprintf("%s - %s", ddPayload.Title, ddPayload.Body),
			Status:    500,
		}

		fmt.Printf("[Webhook] datadig has hit us boom: ", ddPayload.Title)
		processLog(producer, entry)
	})

	fmt.Println("Ingestor running on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func processLog(producer sarama.SyncProducer, entry models.LogEntry) {
	//filter logs
	if entry.Level == "ERROR" || entry.Status >= 500 {
		fmt.Printf("[Hot path] critical log: %s\n", entry.Message)
		bytes, _ := json.Marshal(entry)
		msg := &sarama.ProducerMessage{
			Topic: KafkaTopic,
			Value: sarama.StringEncoder(bytes),
		}
		_, _, err := producer.SendMessage(msg)
		if err != nil {
			log.Println("failed to sned kafka :", err)
		}
	} else {
		fmt.Printf("Info : %s\n", entry.Message)
	}
}

func startLogGenerator(producer sarama.SyncProducer) {
	levels := []string{"INFO"}
	msgs := []string{"User login"}

	for {
		time.Sleep(10 * time.Second)
		lvl := levels[rand.Intn(len(levels))]

		entry := models.LogEntry{
			Timestamp: time.Now(),
			Level:     lvl,
			Service:   "payment-service",
			Message:   msgs[rand.Intn(len(msgs))],
			Status:    200,
		}

		if lvl == "ERROR" {
			entry.Status = 500
		}

		processLog(producer, entry)
	}
}

func setupProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	return sarama.NewSyncProducer([]string{KafkaBroker}, config)
}
