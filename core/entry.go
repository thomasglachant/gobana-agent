package core

import "time"

type EntryMetadata struct {
	AgentVersion string    `json:"agent_version" yaml:"agent_version"`
	Application  string    `json:"application" yaml:"application"`
	Server       string    `json:"server" yaml:"server"`
	Filename     string    `json:"filename" yaml:"filename"`
	Parser       string    `json:"parser" yaml:"parser"`
	CaptureDate  time.Time `json:"capture_date" yaml:"capture_date"`
}

type Entry struct {
	Metadata EntryMetadata     `yaml:"metadata" json:"metadata"`
	Date     time.Time         `yaml:"date" json:"date"`
	Raw      string            `yaml:"raw" json:"raw"`
	Fields   map[string]string `yaml:"fields" json:"fields"`
}
