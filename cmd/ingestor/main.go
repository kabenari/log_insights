package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	pb "github.com/kabenari/log-insight/pkg/api"
	"github.com/kabenari/log-insight/pkg/models"
	"google.golang.org/grpc"
)

const (
	KafkaTopic  = "logs.to.analyze"
	KafkaBroker = "localhost:9092"
	GRPCPort    = ":50051"
)

type server struct {
	pb.UnimplementedLogIngestorServer
	producer sarama.SyncProducer
}

func (s *server) PushLog(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	entry := models.LogEntry{
		Timestamp: time.Now(),
		Level:     req.Level,
		Message:   req.Message,
		Service:   req.ServiceName,
		Status:    int(req.Status),
	}
	processLog(s.producer, entry)
	return &pb.LogResponse{Success: true, AckId: uuid.New().String()}, nil
}

func main() {
	producer, err := setupProducer()
	if err != nil {
		log.Fatal("failed to setup producer")
	}

	defer producer.Close()

	go func() {
		lis, err := net.Listen("tcp", GRPCPort)
		if err != nil {
			log.Fatalf("failed to listen on grpc port: %v")
		}

		s := grpc.NewServer()
		pb.RegisterLogIngestorServer(s, &server{producer: producer})
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve grpc: %v")
		}
	}()

	go startLogGenerator(producer)

	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		var entry models.LogEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		processLog(producer, entry)
		w.WriteHeader(http.StatusAccepted)
	})

	http.HandleFunc("/webhook/datadog", func(w http.ResponseWriter, r *http.Request) {
		var ddPayload struct {
			Title string `json:"title"`
			Body  string `json:"body"`
			Date  int64  `json:"date"`
		}

		if err := json.NewDecoder(r.Body).Decode(&ddPayload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		entry := models.LogEntry{
			Timestamp: time.Unix(ddPayload.Date/1000, 0),
			Level:     "ERROR",
			Message:   ddPayload.Body,
			Service:   "External-Datadog",
			Status:    500,
		}

		fmt.Printf("Received Datadog alert: %s\n", ddPayload.Title)
		processLog(producer, entry)
	})

	fmt.Println("ingestor service started on port 8080")
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
