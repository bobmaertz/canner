package config

import "time"

// Configurations exported
type Configurations struct {
	Server   ServerConfigurations
	Matchers []Matchers
}

// ServerConfigurations exported
type ServerConfigurations struct {
	Port int
    ReadTimeout time.Duration
    WriteTimeout time.Duration
    IdleTimeout time.Duration
    ReadHeaderTimeout time.Duration
}

// Matchers
type Matchers struct {
	Request  Request
	Response Response
}

type Request struct {
	Path    string
	Method  string
	Headers map[string]string
	Body    string
}

type Response struct {
	Body       string
	StatusCode int
	Headers    map[string]string
	Latency    *LatencyConfig
}

type LatencyConfig struct {
	Type  string
	Delay time.Duration
}
