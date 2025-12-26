package main

import (
	"encoding/json"
	"log"
	"net/http"
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

	http.Handlefunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		var entry models.LogEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		processLog(producer, entry)
		w.WriteHeader(http.StatusOK)
	})

	func processLog(producer sarama.SyncProducer, entry models.LogEntry){
		//filter logs
	}
}
