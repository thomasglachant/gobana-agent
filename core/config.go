package core

import "github.com/creasty/defaults"

type SlackConfig struct {
	Webhook string `yaml:"webhook"`
}

type SMTPConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	SSLEnabled bool   `yaml:"ssl_enabled"`
	FromName   string `yaml:"from_name"`
	FromEmail  string `yaml:"from_email"`
}

type LookupConfig struct {
	Name     string `yaml:"name"`
	Patterns []struct {
		Name  string `yaml:"name"`
		Type  string `yaml:"type"`
		Value string `yaml:"value"`
	} `yaml:"patterns"`
	Files []string `yaml:"files"`
}

type MetadataConfig struct {
	Application string `yaml:"application"`
	Server      string `yaml:"server"`
}

type AlertSubscriptionConfig struct {
	Type    string   `yaml:"type"`
	Value   string   `yaml:"value"`
	Lookups []string `yaml:"lookups"`
}

type AlertsConfig struct {
	Subscriptions []AlertSubscriptionConfig `yaml:"subscriptions"`
}

type ConfigClient struct {
	Metadata MetadataConfig `yaml:"metadata"`
	Mode     string         `yaml:"mode"`
	Lookups  []LookupConfig `yaml:"lookups"`
	Alerts   AlertsConfig   `yaml:"alerts"`
	SMTP     SMTPConfig     `yaml:"smtp"`
}

type ConfigStruct struct {
	Client ConfigClient `yaml:"client"`
}

func (s *ConfigStruct) UnmarshalYAML(unmarshal func(interface{}) error) error {
	_ = defaults.Set(s)
	type plain ConfigStruct
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
}
