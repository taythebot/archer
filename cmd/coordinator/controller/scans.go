package controller

import (
	"errors"
	"strconv"

	"github.com/taythebot/archer/cmd/coordinator/form"
	"github.com/taythebot/archer/cmd/coordinator/types"
	"github.com/taythebot/archer/pkg/model"
	"github.com/taythebot/archer/pkg/queue"
	"github.com/taythebot/archer/pkg/scheduler"
	globalTypes "github.com/taythebot/archer/pkg/types"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ScanController struct {
	DB        *gorm.DB
	Queue     *queue.Client
	Scheduler *scheduler.Scheduler
}

// GetAll scans
func (ctrl *ScanController) GetAll(ctx *fiber.Ctx) error {
	var scans []model.Scans
	if err := ctrl.DB.Find(&scans).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get scans: "+err.Error())
	}

	return ctx.Status(fiber.StatusOK).JSON(scans)
}

// Create a new scan
func (ctrl *ScanController) Create(ctx *fiber.Ctx) error {
	body := &form.NewScan{}

	// Validate body
	if err := ctx.BodyParser(body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := form.ValidateStruct(body); err != nil {
		return err
	}

	// Validate module
	for _, m := range body.Modules {
		var valid bool
		for _, module := range globalTypes.Modules {
			if module == m {
				valid = true
				break
			}
		}

		// Check for consistency

		if !valid {
			return &types.ValidationError{Errors: []types.ErrorResponse{{
				Type:    "invalid_request_error",
				Param:   "module",
				Message: "Value '" + m + "' is not a valid module",
			}}}
		}
	}

	// Validate targets
	// for _, t := range body.Targets {
	// 	if net.ParseIP(t) == nil {
	// 		return &types.ValidationError{Errors: []types.ErrorResponse{{
	// 			Type:    "invalid_request_error",
	// 			Param:   "targets",
	// 			Message: "Value '" + t + "' is not a valid IP address",
	// 		}}}
	// 	}
	// }

	// Convert ports to int32
	ports := make([]int32, 0, len(body.Ports))
	for _, port := range body.Ports {
		ports = append(ports, int32(port))
	}

	// Create new scan
	scan := model.Scans{
		Modules: body.Modules,
		Targets: body.Targets,
		Ports:   ports,
	}
	if err := ctrl.DB.Create(&scan).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create scan: "+err.Error())
	}

	// Determine which module to schedule
	module := "masscan"
	if len(scan.Modules) == 1 {
		module = scan.Modules[0]
	}

	// Schedule tasks
	var err error
	switch module {
	case "masscan":
		_, err = ctrl.Scheduler.Masscan(ctx.Context(), scan.ID, body.Targets, body.Ports)
	case "httpx":
		// Merge targets and ports
		var targets []string
		for _, target := range body.Targets {
			for _, port := range body.Ports {
				targets = append(targets, target+":"+strconv.Itoa(int(port)))
			}
		}

		_, err = ctrl.Scheduler.Httpx(ctx.Context(), scan.ID, targets)
	case "nuclei":
		// Merge targets and ports
		var targets []string
		for _, target := range body.Targets {
			for _, port := range body.Ports {
				targets = append(targets, target+":"+strconv.Itoa(int(port)))
			}
		}

		_, err = ctrl.Scheduler.Nuclei(ctx.Context(), scan.ID, targets, body.NucleiTypes)
	}

	// Catch task error
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create scan: "+err.Error())
	}

	return ctx.Status(fiber.StatusOK).JSON(scan)
}

// Get a scan
func (ctrl *ScanController) Get(ctx *fiber.Ctx) error {
	var scan model.Scans
	if err := ctrl.DB.First(&scan, "id = ?", ctx.Params("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Scan not found")
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to find scan: "+err.Error())
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(scan)
}

// GetTasks for a scan
func (ctrl *ScanController) GetTasks(ctx *fiber.Ctx) error {
	var scan model.Scans

	if err := ctrl.DB.Preload("Tasks").First(&scan, "id = ?", ctx.Params("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Scan not found")
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to find scan: "+err.Error())
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(scan.Tasks)
}
