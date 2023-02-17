package masscan

import (
	"context"
	"fmt"

	"github.com/taythebot/archer/pkg/elasticsearch"
	"github.com/taythebot/archer/pkg/types"

	"github.com/sirupsen/logrus"
)

// Schedule new tasks
func (m *Module) Schedule(_ context.Context, _ types.SchedulerPayload, _ types.Scheduler, _ *elasticsearch.Client, _ *logrus.Entry) (total uint32, err error) {
	return 0, fmt.Errorf("masscan module cannot schedule new tasks")
}
