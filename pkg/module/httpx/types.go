package httpx

import (
	"time"

	"github.com/taythebot/archer/pkg/types"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

// Module for Httpx tasks
type Module struct {
	Config types.HttpxConfig
}

// Payload for task
type Payload struct {
	Scan    string   `json:"scan"`    // Scan ID
	Targets []string `json:"targets"` // Targets to scan
}

// Result contains the parsed Httpx result
type Result struct {
	IP        string       `json:"ip"`
	Output    ResultHttp   `json:"output"` // Rename to output since http is the field in Elasticsearch
	Http      []ResultHttp `json:"http,omitempty"`
	Scan      string       `json:"scan"`
	Timestamp time.Time    `json:"timestamp"`
}

// ResultHttp contains the http object in the results
type ResultHttp struct {
	Port         int                 `json:"port"`
	Csp          []string            `json:"csp,omitempty"`
	Tls          TLS                 `json:"tls,omitempty"`
	Body         string              `json:"body,omitempty"`
	Headers      map[string]string   `json:"headers,omitempty"`
	Hashes       ResultHttpHashes    `json:"hashes"`
	Redirects    ResultHttpRedirects `json:"redirects,omitempty"`
	Technologies []string            `json:"technologies,omitempty"`
	Title        string              `json:"title,omitempty"`
	Scheme       string              `json:"scheme"`
	StatusCode   int                 `json:"status_code"`
	Metadata     ResultHttpMetadata  `json:"metadata"`
}

type ResultHttpHashes struct {
	BodyMmh3     string `json:"body_mmh3,omitempty"`
	BodySha256   string `json:"body_sha256,omitempty"`
	HeaderMmh3   string `json:"header_mmh3"`
	HeaderSha256 string `json:"header_sha256"`
}

type ResultHttpRedirects struct {
	Chains   []ResultHttpRedirectsChain `json:"chains"`
	FinalUrl string                     `json:"final_url"`
}

type ResultHttpRedirectsChain struct {
	Request    string `json:"request,omitempty"`
	Response   string `json:"response,omitempty"`
	StatusCode int    `json:"status_code"`
	Location   string `json:"location"`
	RequestUrl string `json:"request_url"`
}

type ResultHttpMetadata struct {
	Module    string    `json:"module"`
	Task      string    `json:"task"`
	Timestamp time.Time `json:"timestamp"`
}

// Output for Httpx
type Output struct {
	Timestamp time.Time `json:"timestamp"`
	Csp       struct {
		Domains []string `json:"domains"`
	} `json:"csp,omitempty"`
	TLS  TLS `json:"tls,omitempty"`
	Hash struct {
		BodyMmh3     string `json:"body_mmh3,omitempty"`
		BodySha256   string `json:"body_sha256,omitempty"`
		HeaderMmh3   string `json:"header_mmh3"`
		HeaderSha256 string `json:"header_sha256"`
	} `json:"hash"`
	Port      string   `json:"port"`
	Input     string   `json:"input"`
	Title     string   `json:"title,omitempty"`
	Scheme    string   `json:"scheme"`
	FinalURL  string   `json:"final-url,omitempty"`
	RawHeader string   `json:"raw_header"`
	Body      string   `json:"body,omitempty"`
	Tech      []string `json:"tech,omitempty"`
	Chain     []struct {
		Request    string `json:"request,omitempty"`
		Response   string `json:"response,omitempty"`
		StatusCode int    `json:"status_code"`
		Location   string `json:"location"`
		RequestURL string `json:"request-url"`
	} `json:"chain,omitempty"`
	StatusCode int `json:"status_code"`
}

// TLS from Httpx
type TLS struct {
	Cipher                   string   `json:"cipher"`
	Version                  string   `json:"tls_version"`
	ExtensionName            string   `json:"extension_name"`
	DNSNames                 []string `json:"dns_names"`
	CommonNames              []string `json:"common_names"`
	Organization             []string `json:"organization"`
	IssuerCommonName         []string `json:"issuer_common_name"`
	IssuerOrganization       []string `json:"issue_organization"`
	FingerprintSha256        string   `json:"fingerprint_sha256"`
	FingerprintSha256Openssl string   `json:"fingerprint_sha256_openssl"`
}
