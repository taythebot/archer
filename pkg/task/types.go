package task

import (
	"bufio"
	"context"
	"sync"

	"github.com/taythebot/archer/pkg/coordinator"
	"github.com/taythebot/archer/pkg/elasticsearch"
	"github.com/taythebot/archer/pkg/scheduler"
	"github.com/taythebot/archer/pkg/types"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

var json = jsoniter.ConfigFastest

// TaskHandler handles queue tasks via a custom mux router
type TaskHandler struct {
	Elasticsearch    *elasticsearch.Client
	Coordinator      *coordinator.Client
	Scheduler        *scheduler.Scheduler
	muxMutex         sync.RWMutex
	muxEntries       map[string]muxEntry
	muxEntriesSorted []muxEntry // slice of entries sorted from longest to shortest.
}

type muxEntry struct {
	h       TaskModule
	pattern string
}

// TaskModule is the abstraction for all tasks
type TaskModule interface {
	// Name of the module
	Name() string

	// Payload returns the TaskPayload struct
	Payload() types.TaskPayload

	// ProcessTask executes the task
	ProcessTask(context.Context, string, types.TaskPayload, *elasticsearch.Indexer, *logrus.Entry) (uint32, error)

	// ProcessStdout from child process
	ProcessStdout(context.Context, *bufio.Scanner, chan<- types.TaskResult, string, string, *logrus.Entry) error

	// ProcessStderr from child process
	ProcessStderr(context.Context, *bufio.Scanner, *logrus.Entry) error

	// Schedule module with previous task data
	Schedule(context.Context, types.SchedulerPayload, types.Scheduler, *elasticsearch.Client, *logrus.Entry) (uint32, error)
}
