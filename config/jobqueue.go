package config

import (
	"fmt"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/utils"
	"log/slog"
)

var jobQueues map[string]*artifex.Dispatcher
var jobCounts map[string]int

const (
	QueueDefault      = "wakapi.default"
	QueueProcessing   = "wakapi.processing"
	QueueReports      = "wakapi.reports"
	QueueMails        = "wakapi.mail"
	QueueImports      = "wakapi.imports"
	QueueHousekeeping = "wakapi.housekeeping"
)

type JobQueueMetrics struct {
	Queue        string
	EnqueuedJobs int
	FinishedJobs int
}

func init() {
	jobQueues = make(map[string]*artifex.Dispatcher)
}

func StartJobs() {
	InitQueue(QueueDefault, 1)
	InitQueue(QueueProcessing, utils.HalfCPUs())
	InitQueue(QueueReports, 1)
	InitQueue(QueueMails, 1)
	InitQueue(QueueImports, 1)
	InitQueue(QueueHousekeeping, utils.HalfCPUs())
}

func InitQueue(name string, workers int) error {
	if _, ok := jobQueues[name]; ok {
		return fmt.Errorf("queue '%s' already existing", name)
	}
	slog.Info("creating job queue '%s' (%d workers)", name, workers)
	jobQueues[name] = artifex.NewDispatcher(workers, 4096)
	jobQueues[name].Start()
	return nil
}

func GetDefaultQueue() *artifex.Dispatcher {
	return GetQueue(QueueDefault)
}

func GetQueue(name string) *artifex.Dispatcher {
	if _, ok := jobQueues[name]; !ok {
		InitQueue(name, 1)
	}
	return jobQueues[name]
}

func GetQueueMetrics() []*JobQueueMetrics {
	metrics := make([]*JobQueueMetrics, 0, len(jobQueues))
	for name, queue := range jobQueues {
		metrics = append(metrics, &JobQueueMetrics{
			Queue:        name,
			EnqueuedJobs: queue.CountEnqueued(),
			FinishedJobs: queue.CountDispatched(),
		})
	}
	return metrics
}

func CloseQueues() {
	for _, q := range jobQueues {
		q.Stop()
	}
}
