package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type DatadogPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Date  int64  `json:"date"`
}

func main() {
	targetUrl := "http://localhost:8080/webhook/datadog"

	alerts := []struct{ title, body string }{
		{"hihg memory", "host i - 1111"},
	}

	fmt.Println("mocking od DD started....")

	for {
		alert := alerts[rand.Intn(len(alerts))]

		payload := DatadogPayload{
			Title: alert.title,
			Body:  alert.body,
			Date:  time.Now().UnixMilli(),
		}

		data, _ := json.Marshal(payload)

		resp, err := http.Post(targetUrl, "applicaiton/json", bytes.NewBuffer(data))
		if err != nil {
			fmt.Printf("fdailed tosed dd paylaod")
		} else {
			fmt.Printf("sent on webhook")
			resp.Body.Close()
		}

		time.Sleep(10 * time.Second)
	}
}
