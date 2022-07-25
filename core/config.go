//nolint:goconst
package core

import (
	"fmt"
	"os"
	"regexp"
	"strings"

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

//nolint:gocyclo
func CheckConfig(config interface{}) error {
	// apply validator
	validate := validator.New()
	_ = validate.RegisterValidation("regular_name", ValidateRegularName)

	err := validate.Struct(config)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		for _, err := range err.(validator.ValidationErrors) {
			field := strings.Join(strings.Split(err.Namespace(), ".")[1:], ".")
			switch {
			case err.Tag() == "required":
				return fmt.Errorf("\"%s\" is required", field)
			case err.Tag() == "oneof":
				return fmt.Errorf("\"%s\" must be one of \"%s\"", field, strings.Join(strings.Split(err.Param(), " "), "\", \""))
			case err.Kind().String() == "slice" && err.Tag() == "lt":
				return fmt.Errorf("\"%s\" must contains less than %s item(s)", field, err.Param())
			case err.Kind().String() == "slice" && err.Tag() == "lte":
				return fmt.Errorf("\"%s\" must contains maximum %s item(s)", field, err.Param())
			case err.Kind().String() == "slice" && err.Tag() == "gt":
				return fmt.Errorf("\"%s\" must contains more than %s item(s)", field, err.Param())
			case err.Kind().String() == "slice" && err.Tag() == "gte":
				return fmt.Errorf("\"%s\" must contains at least %s item(s)", field, err.Param())
			case err.Tag() == "eq":
				return fmt.Errorf("\"%s\" must be equal to \"%s\"", field, err.Param())
			case err.Tag() == "ne":
				return fmt.Errorf("\"%s\" must not be equal to \"%s\"", field, err.Param())
			case err.Kind().String() == "slice" && err.Tag() == "unique" && err.Param() != "":
				return fmt.Errorf("\"%s.%s\" property must be unique", field, err.Param())
			case err.Kind().String() == "slice" && err.Tag() == "unique":
				return fmt.Errorf("\"%s\" entries must be unique", field)
			case err.Tag() == "required_if":
				return fmt.Errorf("\"%s\" is required when %s is \"%s\"", field, strings.Split(err.Param(), " ")[0], strings.Split(err.Param(), " ")[1])
			case err.Tag() == "regular_name":
				return fmt.Errorf("\"%s\" must contains only letters, numbers, space, \"-\" or \"_\"", field)
			default:
				return fmt.Errorf("\"%s\" fails to validate constraint \"%s\"", field, err.Tag())
			}
		}
	}

	return nil
}

func ValidateRegularName(fl validator.FieldLevel) bool {
	reg := regexp.MustCompile(`^[a-zA-Z0-9_ \-]+$`)
	return reg.MatchString(fl.Field().String())
}
