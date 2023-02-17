package form

import (
	"github.com/taythebot/archer/cmd/coordinator/types"

	"github.com/go-playground/validator/v10"
)

// validate holds the global validator instance
var validate = validator.New()

// ValidateStruct validates a struct and returns errors
func ValidateStruct(i interface{}) *types.ValidationError {
	if err := validate.Struct(i); err != nil {
		var errors []types.ErrorResponse

		for _, validatorErr := range err.(validator.ValidationErrors) {
			errors = append(errors, types.ErrorResponse{
				Type:    "invalid_request_error",
				Param:   validatorErr.Field(), // TODO: Extract json tag name
				Message: validatorErr.Error(), // TODO: Extract better error message
			})
		}

		return &types.ValidationError{Errors: errors}
	}

	return nil
}
