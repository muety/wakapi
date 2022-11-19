package config

import (
	"fmt"
	"github.com/muety/artifex"
)

var jobqueues map[string]*artifex.Dispatcher

const (
	QueueDefault    = ""
	QueueProcessing = "processing"
	QueueMails      = "mails"
)

func init() {
	jobqueues = make(map[string]*artifex.Dispatcher)
}

func InitQueue(name string, workers int) error {
	if _, ok := jobqueues[name]; ok {
		return fmt.Errorf("queue '%s' already existing", name)
	}
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
