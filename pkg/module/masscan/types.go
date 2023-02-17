package masscan

import (
	"time"

	"github.com/taythebot/archer/pkg/types"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

// Module for Masscan tasks
type Module struct {
	Config types.MasscanConfig
}

// Payload for task
type Payload struct {
	Scan    string   `json:"scan"`    // Scan ID
	Targets []string `json:"targets"` // Targets to scan
	Ports   []uint16 `json:"ports"`   // Ports to scan
	Shard   string   `json:"shard"`   // Shard for distribution
	Seed    int64    `json:"seed"`    // Seed for distribution
}

// Result contains the parsed Masscan result
type Result struct {
	IP        string       `json:"ip"`
	Port      ResultPort   `json:"port,omitempty"`
	Ports     []ResultPort `json:"ports,omitempty"`
	Scan      string       `json:"scan"`
	Timestamp time.Time    `json:"timestamp"`
}

type ResultPort struct {
	Port     int            `json:"port"`
	Metadata ResultMetadata `json:"metadata"`
}

type ResultMetadata struct {
	Module    string    `json:"module"`
	Task      string    `json:"task"`
	Timestamp time.Time `json:"timestamp"`
}
