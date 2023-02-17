package nuclei

import (
	"time"

	"github.com/taythebot/archer/pkg/types"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

// Module for Nuclei tasks
type Module struct {
	Config types.NucleiConfig
}

// Payload for task
type Payload struct {
	Scan          string   `json:"scan"`           // Scan ID
	Targets       []string `json:"targets"`        // Targets to scan
	TemplateTypes []string `json:"template_types"` // Types of templates of run
}

// Result contains the parsed Nuclei result
type Result struct {
	IP        string          `json:"ip"`
	Detection ResultDetection `json:"detection"`
	Scan      string          `json:"scan"`
	Timestamp time.Time       `json:"timestamp"`
}

// ResultDetection contains the detection object in the results
type ResultDetection struct {
	Port             int                     `json:"port"`
	TemplateId       string                  `json:"template_id"`
	Type             string                  `json:"type"`
	ExtractedResults []string                `json:"extracted_results,omitempty"`
	MatcherName      string                  `json:"matcher_name,omitempty"`
	MatchedAt        string                  `json:"matched_at,omitempty"`
	Name             string                  `json:"name"`
	Description      string                  `json:"description"`
	Severity         string                  `json:"severity"`
	Tags             []string                `json:"tags"`
	Metadata         ResultDetectionMetadata `json:"metadata"`
}

type ResultDetectionMetadata struct {
	Module    string    `json:"module"`
	Task      string    `json:"task"`
	Timestamp time.Time `json:"timestamp"`
}

// Output for Nuclei
type Output struct {
	TemplateId       string   `json:"template-id"`
	ExtractedResults []string `json:"extracted-results,omitempty"`
	MatcherName      string   `json:"matcher-name,omitempty"`
	Type             string   `json:"type"`
	Host             string   `json:"host"`
	MatchedAt        string   `json:"matched-at,omitempty"`
	Timestamp        string   `json:"timestamp"`
	Info             struct {
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Severity    string   `json:"severity,omitempty"`
		Tags        []string `json:"tags,omitempty"`
	} `json:"info"`
}
