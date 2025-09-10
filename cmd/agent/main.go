package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"open-telemetry/internal/collector"
	"open-telemetry/internal/reporter"
)

type AgentConfig struct {
	CollectionInterval int    `json:"collection_interval"`
	Endpoint           string `json:"endpoint"`
	// CertFile           string `json:"cert_file"`
	// KeyFile            string `json:"key_file"`
	// CACertFile         string `json:"ca_cert_file"`
}

func main() {
	log.Println("Starting Telemetry Agent...")

	config, err := loadJsonConfig("configs/config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	tlsConfig, err := setupTLSClient(config)
	if err != nil {
		log.Fatalf("Failed to set up TLS client: %v", err)
	}

	coll := collector.NewCollector()
	rep := reporter.NewReporter(config.Endpoint, tlsConfig)

	dataChan := make(chan collector.TelemetryData)
	errorChan := make(chan error)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Concurrently collect and report data
	go collectAndReport(coll, rep, dataChan, errorChan, config.CollectionInterval)

	for {
		select {
		case data := <-dataChan:
			log.Printf("Collected data: %+v", data)
			go func(d collector.TelemetryData) {
				if err := rep.ReportData(d); err != nil {
					log.Printf("Failed to report data: %v", err)
				}
			}(data)
		case err := <-errorChan:
			log.Printf("Collection error: %v", err)
		case <-stopChan:
			log.Println("Shutting down agent gracefully...")
			return
		}
	}
}

func loadJsonConfig(path string) (*AgentConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var config AgentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}
	return &config, nil
}

func setupTLSClient(config *AgentConfig) (*tls.Config, error) {
	// TODO: load the client certificates and key here.
	return &tls.Config{
		InsecureSkipVerify: true, // WARNING: This is for demo purposes. DO NOT use in production.
	}, nil
}

func collectAndReport(coll *collector.Collector, rep *reporter.Reporter, dataChan chan<- collector.TelemetryData, errorChan chan<- error, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		data, err := coll.Collect()
		if err != nil {
			errorChan <- err
			continue
		}
		dataChan <- *data
	}
}
