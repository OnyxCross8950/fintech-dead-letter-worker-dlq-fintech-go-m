package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type paymentEvent struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Attempts int    `json:"attempts"`
}

type queuedMessage struct {
	MessageID string `json:"message_id"`
	Payload   string `json:"payload"`
}

func shouldQuarantine(p paymentEvent) bool { return p.Attempts >= 3 || p.Amount >= 100000 }

func handleDeadLetters(c *queueClient, w http.ResponseWriter, r *http.Request) {
	var messages []queuedMessage
	if err := c.consume(10, 30, &messages); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	processed := 0
	for _, msg := range messages {
		var p paymentEvent
		if json.Unmarshal([]byte(msg.Payload), &p) == nil && shouldQuarantine(p) {
			log.Printf("audit payment=%s action=quarantine amount=%d currency=%s at=%s", p.ID, p.Amount, p.Currency, time.Now().UTC().Format(time.RFC3339))
			processed++
		}
		if err := c.ack(msg.MessageID); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
	}
	fmt.Fprintf(w, "processed=%d", processed)
}

func main() {
	c := newQueueClient()
	http.HandleFunc("/dead-letters", func(w http.ResponseWriter, r *http.Request) { handleDeadLetters(c, w, r) })
	log.Println("dead-letter worker listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
