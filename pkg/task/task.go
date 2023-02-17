package task

import (
	"github.com/taythebot/archer/pkg/coordinator"
	"github.com/taythebot/archer/pkg/elasticsearch"
	"github.com/taythebot/archer/pkg/scheduler"
)

func New(es *elasticsearch.Client, coordinator *coordinator.Client, scheduler *scheduler.Scheduler) *TaskHandler {
	return &TaskHandler{
		Elasticsearch: es,
		Coordinator:   coordinator,
		Scheduler:     scheduler,
		muxEntries:    make(map[string]muxEntry),
	}
}
