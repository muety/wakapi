package config

import "github.com/muety/wakapi/utils"

type SharedData struct {
	*utils.ConcurrentMap[string, interface{}]
}

func NewSharedData() *SharedData {
	return &SharedData{utils.NewConcurrentMap[string, interface{}]()}
}
