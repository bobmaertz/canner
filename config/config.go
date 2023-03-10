package config

// Configurations exported
type Configurations struct {
	Server   ServerConfigurations
	Matchers []Matchers
}

// ServerConfigurations exported
type ServerConfigurations struct {
	Port int
}

// Matchers
type Matchers struct {
	Request  Request
	Response Response
}

type Request struct {
	Path    string
	Headers map[string]string
}

type Response struct {
	//Type    string
	Body       string
	StatusCode int
	Headers    map[string]string
}
