package yaml

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// validateDuration parses a string into time.Duration
func validateDuration(fl validator.FieldLevel) bool {
	_, err := time.ParseDuration(fl.Field().String())
	return err == nil
}
