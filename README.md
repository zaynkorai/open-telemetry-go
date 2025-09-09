## Go Network Telemetry Agent
This project is a small, cross-platform telemetry agent written in Go. It is designed to collect system and network statistics and report them to a central server. This project serves as a showcase of a lightweight, enterprise grade agent.

### Key Features
- Collects hostname, OS type, uptime, active connections, open ports, and data transfer rates.

- Reads its configuration from configs/config.json, allowing easy modification of the collection interval and server endpoint.

- Built-in error handling and a retry mechanism. In case of network failure, it queues data for later delivery.

- Uses Go's concurrency primitives (goroutines and channels) to handle data collection and reporting asynchronously, ensuring non-blocking operations.

- Includes both unit tests for individual components and an integration test to validate the entire data flow from collection to reporting.

