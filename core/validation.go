package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func TranslateValidationError(err validator.FieldError, ignoreFirstNamespace bool) error {
	field, validationErr := ExtractValidationError(err, ignoreFirstNamespace)
	return fmt.Errorf("%w %s", field, validationErr)
}

//nolint:gocyclo
func ExtractValidationError(err validator.FieldError, ignoreFirstNamespace bool) (string, error) {
	field := ""
	for i, v := range strings.Split(err.Namespace(), ".") {
		if ignoreFirstNamespace && i == 0 {
			continue
		}
		if field != "" {
			field += "."
		}

		if len(v) == 1 {
			field += strings.ToLower(v[0:1])
		} else if len(v) > 1 {
			field += strings.ToLower(v[0:1]) + v[1:]
		}
	}

	switch {
	case err.Tag() == "required":
		return field, fmt.Errorf("must not be empty")
	case err.Tag() == "oneof":
		return field, fmt.Errorf("must be one of \"%s\"", strings.Join(strings.Split(err.Param(), " "), "\", \""))
	case err.Kind().String() == "slice" && err.Tag() == "lt":
		return field, fmt.Errorf("must contains less than %s item(s)", err.Param())
	case err.Kind().String() == "slice" && err.Tag() == "lte":
		return field, fmt.Errorf("must contains maximum %s item(s)", err.Param())
	case err.Kind().String() == "slice" && err.Tag() == "gt":
		return field, fmt.Errorf("must contains more than %s item(s)", err.Param())
	case err.Kind().String() == "slice" && err.Tag() == "gte":
		return field, fmt.Errorf("must contains at least %s item(s)", err.Param())
	case err.Tag() == "eq":
		return field, fmt.Errorf("must be equal to \"%s\"", err.Param())
	case err.Tag() == "ne":
		return field, fmt.Errorf("must not be equal to \"%s\"", err.Param())
	case err.Kind().String() == "slice" && err.Tag() == "unique" && err.Param() != "":
		return fmt.Sprintf("%s.%s", field, err.Param()), fmt.Errorf("must be unique")
	case err.Kind().String() == "slice" && err.Tag() == "unique":
		return field, fmt.Errorf("entries must are unique")
	case err.Tag() == "required_if":
		return field, fmt.Errorf("must not be empty when %s is \"%s\"", strings.Split(err.Param(), " ")[0], strings.Split(err.Param(), " ")[1]) //nolint:lll
	case err.Tag() == "slug":
		return field, fmt.Errorf("must contains only letters, numbers, \"-\" or \"_\"")
	case err.Tag() == "simple_name":
		return field, fmt.Errorf("must contains only letters, numbers, spaces, \"-\" or \"_\"")
	default:
		return field, fmt.Errorf("fails to validate constraint \"%s\"", err.Tag())
	}
}

func ValidateSlug(fl validator.FieldLevel) bool {
	reg := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	return reg.MatchString(fl.Field().String())
}

func ValidateSimpleName(fl validator.FieldLevel) bool {
	reg := regexp.MustCompile(`^[a-zA-Z0-9_ \-]+$`)
	return reg.MatchString(fl.Field().String())
}
