package collector

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/net"
)

type TelemetryData struct {
	Timestamp         string            `json:"timestamp"`
	Hostname          string            `json:"hostname"`
	OS                string            `json:"os"`
	Uptime            uint64            `json:"uptime"`
	TotalConnections  int               `json:"total_connections"`
	OpenTCPPorts      []string          `json:"open_tcp_ports"`
	OpenUDPPorts      []string          `json:"open_udp_ports"`
	DataTransferBytes map[string]uint64 `json:"data_transfer_bytes"`
}

type Collector struct{}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) Collect() (*TelemetryData, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("get hostname: %w", err)
	}

	info, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("get host info: %w", err)
	}

	connections, err := net.Connections("inet")
	if err != nil {
		return nil, fmt.Errorf("get network connections: %w", err)
	}

	tcpPorts, udpPorts := c.getOpenPorts(connections)
	dataTransfer, err := c.getDataTransferRates()
	if err != nil {
		return nil, fmt.Errorf("get data transfer rates: %w", err)
	}

	return &TelemetryData{
		Timestamp:         time.Now().Format(time.RFC3339),
		Hostname:          hostname,
		OS:                runtime.GOOS,
		Uptime:            info.Uptime,
		TotalConnections:  len(connections),
		OpenTCPPorts:      tcpPorts,
		OpenUDPPorts:      udpPorts,
		DataTransferBytes: dataTransfer,
	}, nil
}

func (c *Collector) getOpenPorts(connections []net.ConnectionStat) ([]string, []string) {
	tcpPorts := make(map[string]struct{})
	udpPorts := make(map[string]struct{})
	var tcpList, udpList []string

	for _, conn := range connections {
		if conn.Status == "LISTEN" {
			addr := fmt.Sprintf("%s:%d", conn.Laddr.IP, conn.Laddr.Port)
			if conn.Type == 1 { // 1 == (TCP)
				tcpPorts[addr] = struct{}{}
			} else if conn.Type == 2 { // 2 == (UDP)
				udpPorts[addr] = struct{}{}
			}
		}
	}

	for p := range tcpPorts {
		tcpList = append(tcpList, p)
	}
	for p := range udpPorts {
		udpList = append(udpList, p)
	}

	return tcpList, udpList
}

func (c *Collector) getDataTransferRates() (map[string]uint64, error) {
	counters, err := net.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("get IO counters: %w", err)
	}

	if len(counters) == 0 {
		return nil, fmt.Errorf("no network interfaces found")
	}

	return map[string]uint64{
		"bytes_sent":     counters[0].BytesSent,
		"bytes_received": counters[0].BytesRecv,
	}, nil
}
