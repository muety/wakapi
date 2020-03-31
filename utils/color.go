package utils

import (
	"github.com/muety/wakapi/models"
	"strings"
)

func FilterLanguageColors(all map[string]string, summary *models.Summary) map[string]string {
	subset := make(map[string]string)
	for _, item := range summary.Languages {
		if c, ok := all[strings.ToLower(item.Key)]; ok {
			subset[strings.ToLower(item.Key)] = c
		}
	}
	return subset
}
