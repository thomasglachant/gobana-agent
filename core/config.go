package core

import (
	"fmt"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

var AppConfig *ConfigStruct

type SMTPConfig struct {
	Host       string `yaml:"host" validate:"required"`
	Port       int    `yaml:"port" validate:"required,min=1,max=65535"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	SSLEnabled bool   `yaml:"ssl_enabled"`
	FromName   string `yaml:"from_name" validate:"required"`
	FromEmail  string `yaml:"from_email" validate:"required,email"`
}

type MetadataConfig struct {
	Application string `yaml:"application" validate:"required,alphanum"`
	Server      string `yaml:"server"`
}

type ParserConfig struct {
	Name          string            `yaml:"name" validate:"required,alphanum"`
	Mode          string            `yaml:"mode" validate:"required,oneof=json regex"`
	RegexPattern  string            `yaml:"regex_pattern" validate:"required_if=Mode regex"`
	RegexFields   []string          `yaml:"regex_fields" validate:"required_if=Mode regex,dive,required"`
	JsonFields    map[string]string `yaml:"json_fields" validate:"required_if=Mode json,dive,required"`
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
	Metadata MetadataConfig  `yaml:"metadata" validate:"required"`
	Parsers  []*ParserConfig `yaml:"parsers" validate:"required,gte=1,unique=Name,dive"`
	Alerts   AlertConfig     `yaml:"alerts" validate:""`
}

type CommonConfig struct {
	SMTP SMTPConfig `yaml:"smtp"`
}

type ConfigStruct struct {
	Agent  AgentConfig  `yaml:"agent" validate:"required"`
	Common CommonConfig `yaml:"common" validate:"required"`
}

func (s *ConfigStruct) UnmarshalYAML(unmarshal func(interface{}) error) error {
	_ = defaults.Set(s)
	type plain ConfigStruct
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
}

func CheckConfig(config *ConfigStruct) error {
	// apply validator
	validate := validator.New()
	err := validate.Struct(config)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		for _, err := range err.(validator.ValidationErrors) {
			field := strings.Join(strings.Split(err.Namespace(), ".")[1:], ".")
			if err.Tag() == "required" {
				return fmt.Errorf("\"%s\" is required", field)
			} else if err.Tag() == "oneof" {
				return fmt.Errorf("\"%s\" must be one of \"%s\"", field, strings.Join(strings.Split(err.Param(), " "), "\", \""))
			} else if err.Kind().String() == "slice" && err.Tag() == "lt" {
				return fmt.Errorf("\"%s\" must contains less than %s item(s)", field, err.Param())
			} else if err.Kind().String() == "slice" && err.Tag() == "lte" {
				return fmt.Errorf("\"%s\" must contains maximum %s item(s)", field, err.Param())
			} else if err.Kind().String() == "slice" && err.Tag() == "gt" {
				return fmt.Errorf("\"%s\" must contains more than %s item(s)", field, err.Param())
			} else if err.Kind().String() == "slice" && err.Tag() == "gte" {
				return fmt.Errorf("\"%s\" must contains at least %s item(s)", field, err.Param())
			} else if err.Tag() == "eq" {
				return fmt.Errorf("\"%s\" must be equal to \"%s\"", field, err.Param())
			} else if err.Tag() == "ne" {
				return fmt.Errorf("\"%s\" must not be equal to \"%s\"", field, err.Param())
			} else if err.Kind().String() == "slice" && err.Tag() == "unique" && err.Param() != "" {
				return fmt.Errorf("\"%s.%s\" property must be unique", field, err.Param())
			} else if err.Kind().String() == "slice" && err.Tag() == "unique" {
				return fmt.Errorf("\"%s\" entries must be unique", field)
			} else if err.Tag() == "required_if" {
				return fmt.Errorf("\"%s\" is required when %s is \"%s\"", field, strings.Split(err.Param(), " ")[0], strings.Split(err.Param(), " ")[1])
			} else {
				return fmt.Errorf("\"%s\" fails to validate constraint \"%s\"", field, err.Tag())
			}
		}
	}

	return nil
}
