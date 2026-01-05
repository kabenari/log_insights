package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/kabenari/log-insight/pkg/api"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial grpc: %v", err)
	}
	defer conn.Close()

	c := pb.NewLogIngestorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := c.PushLog(ctx, &pb.LogRequest{
		ServiceName: "test",
		Level:       "ERROR",
		Message:     "test message",
		Timestamp:   time.Now().String(),
		Status:      404,
	})
	if err != nil {
		log.Fatalf("failed to push log: %v", err)
	}

	log.Printf("Log Ack: %s, Success: %v", r.AckId, r.Success)
}
