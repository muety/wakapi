package utils

import (
	"github.com/muety/wakapi/models"
	"strings"
)

func FilterColors(all map[string]string, haystack models.SummaryItems) map[string]string {
	subset := make(map[string]string)
	for _, item := range haystack {
		if c, ok := all[strings.ToLower(item.Key)]; ok {
			subset[strings.ToLower(item.Key)] = c
		}
	}
	return subset
}
