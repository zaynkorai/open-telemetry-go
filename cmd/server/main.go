package main

import (
	"encoding/json"
	"log"
	"net/http"

	"open-telemetry/internal/collector"
)

const port = ":8080"

func main() {
	http.HandleFunc("/telemetry", telemetryHandler)

	log.Printf("Starting mock telemetry server on port %s...", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func telemetryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data collector.TelemetryData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received telemetry data from %s:", data.Hostname)
	log.Printf("  - OS: %s", data.OS)
	log.Printf("  - Uptime: %d seconds", data.Uptime)
	log.Printf("  - Total Connections: %d", data.TotalConnections)
	log.Printf("  - Open TCP Ports: %v", data.OpenTCPPorts)
	log.Printf("  - Data Transfer: %+v", data.DataTransferBytes)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
