package middleware

import (
	"github.com/taythebot/archer/cmd/coordinator/types"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

// Error implements a middleware for error handling
func Error(ctx *fiber.Ctx, err error) error {
	// Default to status code 500
	code := fiber.StatusInternalServerError

	// Store errors
	var errors []types.ErrorResponse

	// Check if validation error
	if e, ok := err.(*types.ValidationError); ok {
		code = fiber.StatusBadRequest
		errors = e.Errors
	}

	// Check if Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		errors = []types.ErrorResponse{{
			Type:    "invalid_request_error",
			Message: e.Error(),
		}}
	}

	// Add existing error
	if len(errors) == 0 {
		errors = []types.ErrorResponse{{
			Type:    "api_error",
			Message: err.Error(),
		}}
	}

	// Log error
	if code >= 500 {
		log.Error(err)
	}

	// Send JSON error
	err = ctx.Status(code).JSON(fiber.Map{"errors": errors})
	if err != nil {
		log.Error("Failed to send error response: %s", err)

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": []types.ErrorResponse{{
				Type:    "invalid_request_error",
				Message: err.Error(),
			}},
		})
	}

	return nil
}
