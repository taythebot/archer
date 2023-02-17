package scheduler

import (
	"time"

	"github.com/taythebot/archer/pkg/queue"

	"gorm.io/gorm"
)

type Scheduler struct {
	DB            *gorm.DB
	Queue         *queue.Client
	TaskRetention time.Duration // TaskRetention is the amount of time a task should be kept in the Queue after completion
}

func New(db *gorm.DB, queue *queue.Client) *Scheduler {
	return &Scheduler{DB: db, Queue: queue}
}
