package middleware

import (
	"errors"

	"github.com/taythebot/archer/pkg/model"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Task middleware for validating task ids
func Task(db *gorm.DB) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Get worker id
		worker := ctx.Get("X-Worker-Id")
		if worker == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Worker ID is missing")
		}

		// Get task id
		id := ctx.Params("id")
		if id == "" {
			return fiber.NewError(fiber.StatusInternalServerError, "Task ID is missing")
		}

		// Get task
		var task model.Tasks
		err := db.
			Preload("Scan", func(db *gorm.DB) *gorm.DB {
				return db.Select("ID", "Modules", "Arguments", "Status")
			}).
			First(&task, "id = ?", id).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "Task not found")
			} else {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to find task: "+err.Error())
			}
		}

		// Add tasks and worker id to context
		ctx.Locals("task", task)
		ctx.Locals("worker", worker)

		return ctx.Next()
	}
}
