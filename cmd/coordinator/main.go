package main

import (
	"flag"

	"github.com/taythebot/archer/cmd/coordinator/controller"
	"github.com/taythebot/archer/cmd/coordinator/middleware"
	"github.com/taythebot/archer/internal/yaml"
	"github.com/taythebot/archer/pkg/model"
	"github.com/taythebot/archer/pkg/queue"
	"github.com/taythebot/archer/pkg/scheduler"
	"github.com/taythebot/archer/pkg/types"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Parse flags
	configFile := flag.String("config", "configs/coordinator.yaml", "Config file")
	migrate := flag.Bool("migrate", false, "Perform database migration")
	debug := flag.Bool("debug", false, "Enable debug level logging")
	flag.Parse()

	// Validate flags
	if *configFile == "" {
		log.Fatal("Exiting Program: config file is required")
	}

	// Create new YAML validator
	y, err := yaml.New()
	if err != nil {
		log.Fatalf("Exiting Program: failed to create YAML validator: %s", err)
	}

	// Parse config file
	parsed, err := y.ValidateFile(*configFile, &types.CoordinatorConfig{})
	if err != nil {
		log.Error(y.FormatError(err))
		log.Fatalf("Exiting Program: failed to parse config file: %s", err)
	}

	// Type assertion for config
	config, ok := parsed.(*types.CoordinatorConfig)
	if !ok {
		log.Fatal("Exiting Program: failed to parse config file: types assertion failed")
	}

	// Set log level
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	// Connect to Database
	db, err := model.ConnectToDB(config.Postgresql)
	if err != nil {
		log.Fatalf("Exiting Program: failed to connect to Postgresql: %s", err)
	}

	// Run database migration
	if *migrate {
		log.Debug("Running database auto migration")
		if err := db.RunMigration(); err != nil {
			log.Fatalf("Exiting Program: failed to run database migration: %s", err)
		}
	}

	// Create new queue client
	client, err := queue.NewClient(config.Redis)
	if err != nil {
		log.Fatalf("Exiting Program: failed to create Queue cilent: %s", err)
	}

	// Create new scheduler
	sch := scheduler.New(db.DB, client)

	// Create new Fiber
	app := fiber.New(fiber.Config{
		AppName:      "Archer Coordinator v1.0.0",
		ServerHeader: "archer",
		ErrorHandler: middleware.Error,
	})

	// Add global middlewares
	app.Use(recover.New())
	app.Use(logger.New())

	// Create version group
	v1 := app.Group("/v1")

	// Health route
	v1.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
	})

	// Add scan routes
	scanCtrl := controller.ScanController{DB: db.DB, Queue: client, Scheduler: sch}
	scan := v1.Group("/scans")
	scan.Get("/", scanCtrl.GetAll)
	scan.Post("/", scanCtrl.Create)
	scan.Get("/:id", scanCtrl.Get)
	scan.Get("/:id/tasks", scanCtrl.GetTasks)

	// Add task routes
	taskCtrl := controller.TaskController{DB: db.DB, Queue: client, Scheduler: sch}
	task := v1.Group("/tasks")
	task.Get("/:id", taskCtrl.Get)
	task.Use("/:id/*", middleware.Task(db.DB))
	task.Post("/:id/started", taskCtrl.Started)
	task.Post("/:id/completed", taskCtrl.Completed)
	task.Post("/:id/failed", taskCtrl.Failed)

	// Listen
	if err := app.Listen(config.Listen); err != nil {
		log.Fatalf("Exiting Program: %s", err)
	}
}
