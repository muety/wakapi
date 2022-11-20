package config

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/muety/artifex"
	"math"
	"runtime"
)

var jobqueues map[string]*artifex.Dispatcher

const (
	QueueDefault    = "wakapi.default"
	QueueProcessing = "wakapi.processing"
	QueueReports    = "wakapi.reports"
)

func init() {
	jobqueues = make(map[string]*artifex.Dispatcher)

	InitQueue(QueueDefault, 1)
	InitQueue(QueueProcessing, int(math.Ceil(float64(runtime.NumCPU())/2.0)))
	InitQueue(QueueReports, 1)
}

func InitQueue(name string, workers int) error {
	if _, ok := jobqueues[name]; ok {
		return fmt.Errorf("queue '%s' already existing", name)
	}
	logbuch.Info("creating job queue '%s' (%d workers)", name, workers)
	jobqueues[name] = artifex.NewDispatcher(workers, 4096)
	jobqueues[name].Start()
	return nil
}

func GetDefaultQueue() *artifex.Dispatcher {
	return GetQueue("")
}

func GetQueue(name string) *artifex.Dispatcher {
	if _, ok := jobqueues[name]; !ok {
		InitQueue(name, 1)
	}
	return jobqueues[name]
}

func CloseQueues() {
	for _, q := range jobqueues {
		q.Stop()
	}
}
