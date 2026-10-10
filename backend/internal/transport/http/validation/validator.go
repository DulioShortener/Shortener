package validation

import (
	"errors"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/DulioShortener/Shortener/backend/internal/apierrors"
	"github.com/go-playground/validator/v10"
)

const maxURLBytes = 4096

var usernamePattern = regexp.MustCompile(`^[a-z0-9._]{2,32}$`)

type Validator struct {
	validator *validator.Validate
}

func New() (*Validator, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	registrations := map[string]validator.Func{
		"username":   validateUsername,
		"notblank":   validateNotBlank,
		"trimmed":    validateTrimmed,
		"haslower":   contains(unicode.IsLower),
		"hasupper":   contains(unicode.IsUpper),
		"hasdigit":   contains(unicode.IsDigit),
		"hasspecial": contains(func(r rune) bool { return unicode.IsPunct(r) || unicode.IsSymbol(r) }),
		"httpurl":    validateHTTPURL,
	}
	for tag, function := range registrations {
		if err := validate.RegisterValidation(tag, function); err != nil {
			return nil, err
		}
	}
	return &Validator{validator: validate}, nil
}

func (v *Validator) Validate(value any) error {
	return v.validator.Struct(value)
}

func InvalidForm(err error) *apierrors.Error {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return apierrors.MalformedBody()
	}
	fields := make(map[string]string, len(validationErrors))
	for _, fieldError := range validationErrors {
		fields[fieldError.Field()] = validationMessage(fieldError)
	}
	return apierrors.InvalidForm(fields)
}

func validateUsername(field validator.FieldLevel) bool {
	value := field.Field().String()
	return usernamePattern.MatchString(value) && !strings.Contains(value, "..")
}

func validateNotBlank(field validator.FieldLevel) bool {
	return strings.TrimSpace(field.Field().String()) != ""
}

func validateTrimmed(field validator.FieldLevel) bool {
	value := field.Field().String()
	return value == strings.TrimSpace(value)
}

func contains(predicate func(rune) bool) validator.Func {
	return func(field validator.FieldLevel) bool {
		for _, character := range field.Field().String() {
			if predicate(character) {
				return true
			}
		}
		return false
	}
}

func validateHTTPURL(field validator.FieldLevel) bool {
	raw := strings.TrimSpace(field.Field().String())
	if raw == "" || len([]byte(raw)) > maxURLBytes || !utf8.ValidString(raw) {
		return false
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Hostname() == "" || parsed.User != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func validationMessage(field validator.FieldError) string {
	switch field.Tag() {
	case "required":
		return "This value is required"
	case "min":
		return "This value must be at least " + field.Param() + " characters long"
	case "max":
		return "This value must be at most " + field.Param() + " characters long"
	case "username":
		return "Use 2-32 lowercase letters, digits, underscores, or periods without consecutive periods"
	case "notblank":
		return "This value cannot contain only whitespace"
	case "trimmed":
		return "This value must not have leading or trailing whitespace"
	case "haslower":
		return "This value must contain at least one lowercase letter"
	case "hasupper":
		return "This value must contain at least one uppercase letter"
	case "hasdigit":
		return "This value must contain at least one digit"
	case "hasspecial":
		return "This value must contain at least one special character"
	case "httpurl":
		return "This value must be an HTTP or HTTPS URL no longer than 4096 bytes"
	default:
		return "This value is invalid"
	}
}
