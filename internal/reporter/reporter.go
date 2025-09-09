package reporter

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"open-telemetry/internal/collector"
)

const (
	SendDataMaxRetries = 3
	SendDataRetryDelay = 5 * time.Second
)

type Reporter struct {
	endpoint   string
	httpClient *http.Client
	mu         sync.Mutex
	// In a real-world application, this would be a persistent queue or a file-based storage.
	// For this example using a simple in-memory queue.
	offlineQueue []collector.TelemetryData
}

func NewReporter(endpoint string, tlsConfig *tls.Config) *Reporter {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// TODO: We can add more advanced dialing logic here, e.g., for retries or DNS resolution.
				return net.Dial(network, addr)
			},
		},
		Timeout: 10 * time.Second,
	}

	return &Reporter{
		endpoint:     endpoint,
		httpClient:   httpClient,
		offlineQueue: make([]collector.TelemetryData, 0),
	}
}

func (r *Reporter) ReportData(data collector.TelemetryData) error {
	// Before sending, attempt to clear any data from the offline queue.
	if len(r.offlineQueue) > 0 {
		r.processOfflineQueueInMemory()
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	for i := 0; i < SendDataMaxRetries; i++ {
		err = r.send(body)
		if err == nil {
			log.Println("Data reported successfully.")
			return nil
		}

		log.Printf("Attempt %d failed to send data: %v", i+1, err)
		if i < SendDataMaxRetries-1 {
			time.Sleep(SendDataRetryDelay)
		}
	}

	// If all retries fail, add the data to the offline queue.
	r.addToOfflineQueueInMemory(data)
	return fmt.Errorf("all retries failed, data queued for later delivery")
}

func (r *Reporter) send(body []byte) error {
	resp, err := r.httpClient.Post(r.endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	return nil
}

func (r *Reporter) addToOfflineQueueInMemory(data collector.TelemetryData) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.offlineQueue = append(r.offlineQueue, data)
	log.Printf("Data added to offline queue. Current size: %d", len(r.offlineQueue))
}

func (r *Reporter) processOfflineQueueInMemory() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.offlineQueue) == 0 {
		return
	}

	log.Printf("Attempting to process offline queue of size: %d", len(r.offlineQueue))

	newQueue := make([]collector.TelemetryData, 0)
	for _, data := range r.offlineQueue {
		body, err := json.Marshal(data)
		if err != nil {
			log.Printf("Failed to marshal queued data, skipping: %v", err)
			continue
		}
		if err := r.send(body); err != nil {
			log.Printf("Failed to send queued data: %v", err)
			newQueue = append(newQueue, data)
		} else {
			log.Println("Successfully sent queued data.")
		}
	}
	r.offlineQueue = newQueue
}
