package main

import (
	"github.com/creasty/defaults"

	"github.com/thomasglachant/spooter/core"
)

var agentConfig *AgentConfig

type MetadataConfig struct {
	Application string `yaml:"application" validate:"required,alphanum"`
	Server      string `yaml:"server"`
}

type ParserConfig struct {
	Name          string            `yaml:"name" validate:"required,alphanum"`
	Mode          string            `yaml:"mode" validate:"required,oneof=json regex"`
	RegexPattern  string            `yaml:"regex_pattern" validate:"required_if=Mode regex"`
	RegexFields   []string          `yaml:"regex_fields" validate:"required_if=Mode regex,dive,required"`
	JSONFields    map[string]string `yaml:"json_fields" validate:"required_if=Mode json,dive,required"`
	FilesIncluded []string          `yaml:"files_included" validate:"required,gte=1,dive,required"`
	FilesExcluded []string          `yaml:"files_excluded" validate:"dive,file,required"`
}

type RecipientConfig struct {
	Type      string `yaml:"type" validate:"required,oneof=email slack_webhook"`
	Recipient string `yaml:"recipient" validate:"required"`
}

type TriggerValueConfig struct {
	Field    string `yaml:"field" validate:"required"`
	Operator string `yaml:"operator" validate:"required,oneof=regex is is_not contains not_contains start_with not_start_with match_regex"`
	Value    string `yaml:"value" validate:"required"`
}

type TriggerConfig struct {
	Name   string               `yaml:"name" validate:"required"`
	Values []TriggerValueConfig `yaml:"values" validate:"required_if=Type values,dive,required"`
}

type AlertConfig struct {
	Recipients []RecipientConfig `yaml:"recipients" validate:"required,gt=0,dive"`
	Triggers   []TriggerConfig   `yaml:"triggers" validate:"dive"`
}

type AgentConfig struct {
	Debug    bool            `yaml:"debug"`
	Metadata MetadataConfig  `yaml:"metadata" validate:"required"`
	Parsers  []*ParserConfig `yaml:"parsers" validate:"required,gte=1,unique=Name,dive"`
	Alerts   AlertConfig     `yaml:"alerts" validate:""`
	SMTP     core.SMTPConfig `yaml:"smtp"`
}

func (s *AgentConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	_ = defaults.Set(s)
	type plain AgentConfig
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
}
