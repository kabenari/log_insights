package models

import "time"

//raw data co,ing in
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"Level`
	Message   string    `json:"message"`
	Service   string    `json:"service"`
	Status    int       `json:"status"`
}

//the data processort the worker will process
type AIResult struct {
	OriginalLog LogEntry `json:"original_log"`
	Analysis    string   `json:"analysis"`
	Fixed       bool     `json:"fixed"`
}
