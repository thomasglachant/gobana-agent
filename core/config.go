package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/go-playground/validator/v10"
)

type SMTPConfig struct {
	Host       string `yaml:"host" validate:"required" default:"localhost"`
	Port       int    `yaml:"port" validate:"required,min=1,max=65535" default:"25"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	SSLEnabled bool   `yaml:"ssl_enabled" default:"false"`
	FromName   string `yaml:"from_name" validate:"required" default:"Spooter"`
	FromEmail  string `yaml:"from_email" validate:"required,email" default:"spooter@localhost"`
}

func ReadConfig(filename string, config interface{}) error {
	var readErr error
	data, readErr := os.ReadFile(filename)
	if readErr != nil {
		return fmt.Errorf("unable to open config file : %s", readErr)
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("unable to decode config : %s", err)
	}

	validationErr := CheckConfig(config)
	if validationErr != nil {
		return fmt.Errorf("validation error : %s", validationErr)
	}

	return nil
}

func CheckConfig(config interface{}) error {
	// apply validator
	validate := validator.New()
	_ = validate.RegisterValidation("slug", ValidateSlug)
	_ = validate.RegisterValidation("simple_name", ValidateSimpleName)

	err := validate.Struct(config)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		for _, err := range err.(validator.ValidationErrors) {
			return TranslateValidationError(err, true)
		}
	}

	return nil
}
