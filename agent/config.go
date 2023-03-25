package agent

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"

	"gobana-agent/core"
)

type ParserConfigStruct struct {
	Name          string            `yaml:"name" validate:"required,simple_name"`
	Mode          string            `yaml:"mode" validate:"required,oneof=json regex"`
	RegexPattern  string            `yaml:"regex_pattern" validate:"required_if=Mode regex"`
	JSONFields    map[string]string `yaml:"json_fields" validate:"required_if=Mode json,dive,required"`
	FilesIncluded []string          `yaml:"files_included" validate:"required,gte=1,dive,required"`
	FilesExcluded []string          `yaml:"files_excluded" validate:"dive,required"`
	DateExtract   struct {
		Field  string `yaml:"field"`
		Format string `yaml:"format"`
	} `yaml:"date_extract"`
}

type RecipientConfigStruct struct {
	Kind      string `yaml:"kind" validate:"required,oneof=email slack_webhook"`
	Recipient string `yaml:"recipient" validate:"required"`
}

type TriggerValueConfigStruct struct {
	Field    string `yaml:"field" validate:"required"`
	Operator string `yaml:"operator" validate:"required,oneof=regex is is_not contains not_contains start_with not_start_with match_regex"`
	Value    string `yaml:"value" validate:"required"`
}

type TriggerConfigStruct struct {
	Name   string                     `yaml:"name" validate:"required,simple_name"`
	Values []TriggerValueConfigStruct `yaml:"values" validate:"required_if=Kind Values,dive,required"`
}

type AlertConfigStruct struct {
	Frequency  int64                   `yaml:"frequency" validate:"required,gt=0" default:"5"`
	Recipients []RecipientConfigStruct `yaml:"recipients" validate:"dive"`
	Triggers   []TriggerConfigStruct   `yaml:"triggers" validate:"dive"`
}

type AgentConfig struct {
	Debug       bool                  `yaml:"debug" default:"false"`
	Application string                `yaml:"application" validate:"required,simple_name"`
	Server      string                `yaml:"server"`
	Parsers     []*ParserConfigStruct `yaml:"parsers" validate:"required,gte=1,unique=Name,dive"`
	Alerts      AlertConfigStruct     `yaml:"alerts" validate:""`
	SMTP        core.SMTPConfig       `yaml:"smtp"`
}

func (s *AgentConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	_ = defaults.Set(s)
	type plain AgentConfig
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
}

func CheckConfig(configFile string) {
	if err := core.ReadConfig(configFile, AppConfig); err != nil {
		fmt.Printf("Invalid config file : %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config file is valid\n")
	os.Exit(0)
}
