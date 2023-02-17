package yaml

import (
	"fmt"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

type Yaml struct {
	Validator *validator.Validate
}

// New creates a new YAML instance
func New() (*Yaml, error) {
	// Create new validator
	validate := validator.New()

	// Register custom validators
	if err := validate.RegisterValidation("duration", validateDuration); err != nil {
		return nil, fmt.Errorf("failed to register custom validator 'time': %s", err)
	}

	return &Yaml{Validator: validate}, nil
}

// Validate and decode a YAML file via io.Reader
func (y *Yaml) Validate(file io.Reader, i interface{}) (interface{}, error) {
	dec := yaml.NewDecoder(file, yaml.Validator(y.Validator), yaml.Strict())
	if err := dec.Decode(i); err != nil {
		return nil, err
	}

	return i, nil
}

// ValidateFile validates and decode a YAML file via file path
func (y *Yaml) ValidateFile(path string, i interface{}) (interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}

	return y.Validate(file, i)
}

// FormatError formats the YAML error
func (y *Yaml) FormatError(err error) string {
	return yaml.FormatError(err, true, true)
}
