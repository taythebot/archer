package controller

import (
	"errors"
	"fmt"

	"github.com/taythebot/archer/cmd/coordinator/form"
	"github.com/taythebot/archer/pkg/model"
	"github.com/taythebot/archer/pkg/queue"
	"github.com/taythebot/archer/pkg/scheduler"
	globalTypes "github.com/taythebot/archer/pkg/types"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type TaskController struct {
	DB        *gorm.DB
	Queue     *queue.Client
	Scheduler *scheduler.Scheduler
}

// Get a task
func (ctrl *TaskController) Get(ctx *fiber.Ctx) error {
	var task model.Tasks
	if err := ctrl.DB.First(&task, "id = ?", ctx.Params("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Task not found")
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to find task: "+err.Error())
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(task)
}

// Started task handler
func (ctrl *TaskController) Started(ctx *fiber.Ctx) error {
	// Get worker id from context
	worker := ctx.Locals("worker")
	if worker == "" {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get worker id from context")
	}

	// Get task from context
	task, ok := ctx.Locals("task").(model.Tasks)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to type assert task")
	}

	// Check if task is already active
	// if task.Status == "active" {
	// 	return fiber.NewError(fiber.StatusBadRequest, "Task is already active")
	// }

	// Execute database transaction
	err := ctrl.DB.Transaction(func(tx *gorm.DB) error {
		// Update column query
		query := map[string]interface{}{
			"status":     "active",
			"started_at": gorm.Expr("NOW()"),
		}

		// Set task status to active
		if err := tx.Omit("Scan").Model(&task).Updates(query).Update("worker_id", worker).Error; err != nil {
			return fmt.Errorf("failed to set task as active: %s", err)
		}

		// Set scan to active
		if task.Scan.Status != "active" {
			if err := tx.Model(&task.Scan).Updates(query).Error; err != nil {
				return fmt.Errorf("failed to set scan as active: %s", err)
			}
		}

		return nil
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update task: "+err.Error())
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
}

// completeScan will attempt to mark a scan as "completed"
func (ctrl *TaskController) completeScan(scan model.Scans) error {
	// Check for active tasks
	var activeTasks int64
	err := ctrl.DB.
		Model(&model.Tasks{}).
		Where("scan_id = ? AND status = ?", scan.ID, "active").
		Count(&activeTasks).
		Error
	if err != nil {
		return fmt.Errorf("failed to get active tasks: %s", err)
	}

	// Complete scan
	if activeTasks == 0 {
		err := ctrl.DB.Model(&scan).Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": gorm.Expr("NOW()"),
		}).Error
		if err != nil {
			return fmt.Errorf("failed to complete scan: %s", err)
		}
	}

	return nil
}

// Completed task handler
func (ctrl *TaskController) Completed(ctx *fiber.Ctx) error {
	// Get worker id from context
	worker := ctx.Locals("worker")
	if worker == "" {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get worker id from context")
	}

	// Get task from context
	task, ok := ctx.Locals("task").(model.Tasks)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to type assert task")
	}

	body := &form.CompletedTask{}

	// Validate body
	if err := ctx.BodyParser(body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := form.ValidateStruct(body); err != nil {
		return err
	}

	// Check task status
	// if task.Status != "active" {
	// 	return fiber.NewError(fiber.StatusBadRequest, "Task is not active")
	// } else if task.Status == "completed" || task.Status == "failed" {
	// 	return fiber.NewError(fiber.StatusBadRequest, "Task has already completed")
	// }
	if task.Status == "completed" || task.Status == "failed" {
		return fiber.NewError(fiber.StatusBadRequest, "Task has already completed")
	}

	// Update task
	err := ctrl.DB.Omit("Scan").Model(&task).Updates(map[string]interface{}{
		"results":      body.Results,
		"status":       "completed",
		"completed_at": gorm.Expr("NOW()"),
	}).Error
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update task: "+err.Error())
	}

	// Find next modules if not scheduler
	var modules []string
	if task.Module != "scheduler" {
		for _, stage := range globalTypes.Stages {
			if stage.Name == task.Module {
				// Check if next modules are in scan
				for _, nextModule := range stage.Next {
					for _, m := range task.Scan.Modules {
						if m == nextModule {
							modules = append(modules, nextModule)
						}
					}
				}
			}
		}
	}

	// Schedule next module
	if task.Module != "scheduler" && len(modules) > 0 && body.Results > 0 {
		for _, module := range modules {
			if _, err := ctrl.Scheduler.Internal(ctx.Context(), module, task.ScanID, task.ID, task.Module); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to schedule next stage: "+err.Error())
			}
		}
	} else {
		// Attempt to complete scan
		if err := ctrl.completeScan(task.Scan); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to complete scan: "+err.Error())
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
}

// Failed task handler
func (ctrl *TaskController) Failed(ctx *fiber.Ctx) error {
	// Get worker id from context
	worker := ctx.Locals("worker")
	if worker == "" {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get worker id from context")
	}

	// Get task from context
	task, ok := ctx.Locals("task").(model.Tasks)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to type assert task")
	}

	// Check task status
	// if task.Status != "active" {
	// 	return fiber.NewError(fiber.StatusBadRequest, "Task is not active")
	// } else if task.Status == "completed" {
	// 	return fiber.NewError(fiber.StatusBadRequest, "Task has already completed")
	// }
	if task.Status == "completed" {
		return fiber.NewError(fiber.StatusBadRequest, "Task has already completed")
	}

	// Update task
	err := ctrl.DB.Omit("Scan").Model(&task).Updates(map[string]interface{}{
		"status":       "failed",
		"completed_at": gorm.Expr("NOW()"),
	}).Error
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update task: "+err.Error())
	}

	// Attempt to complete scan
	if err := ctrl.completeScan(task.Scan); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to complete scan: "+err.Error())
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
}
