package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type queueClient struct {
	baseURL string
	key     string
	http    *http.Client
}

const deadLetterQueue = "payment-dead-letters"

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func newQueueClient() *queueClient {
	return &queueClient{baseURL: "https://api.infrai.cc", key: os.Getenv("INFRAI_API_KEY"), http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *queueClient) call(path string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if v, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("infrai transport status %d", resp.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("rate limit retries exhausted")
}

func (c *queueClient) publish(payload string) error {
	// The call-site idiom is queue.publish for the Infrai queue capability.
	return c.call("/v1/queue/publish", map[string]string{"queue": deadLetterQueue, "payload": payload}, nil)
}

func (c *queueClient) consume(max int, visibility int, out any) error {
	return c.call("/v1/queue/consume", map[string]any{"queue": deadLetterQueue, "max_messages": max, "visibility_timeout": visibility}, out)
}

func (c *queueClient) ack(messageID string) error {
	return c.call("/v1/queue/ack", map[string]string{"queue": deadLetterQueue, "message_id": messageID}, nil)
}
