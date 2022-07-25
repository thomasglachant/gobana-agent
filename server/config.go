package main

import (
	"github.com/creasty/defaults"
	"github.com/thomasglachant/spooter/core"
)

var config = &ServerConfigStruct{}

type ServerConfigStruct struct {
	Debug bool            `yaml:"debug"`
	SMTP  core.SMTPConfig `yaml:"smtp"`
}

func (s *ServerConfigStruct) UnmarshalYAML(unmarshal func(interface{}) error) error {
	_ = defaults.Set(s)
	type plain ServerConfigStruct
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
}
