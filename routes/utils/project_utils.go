package utils

import (
	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"sort"
)

// GetEffectiveProjectsList returns the user's projects, including all alias targets and excluding all remapped project names (alias sources)
// Example: "A" mapped to "AB" using an alias
// -> "A" itself should not appear as a project anymore
// -> Instead, the "virtual" project "AB" shall appear
// See https://github.com/muety/wakapi/issues/231
func GetEffectiveProjectsList(user *models.User, heartbeatSrvc services.IHeartbeatService, aliasSrvc services.IAliasService) ([]string, error) {
	// extract actual projects from heartbeats
	realProjects, err := heartbeatSrvc.GetEntitySetByUser(models.SummaryProject, user.ID)
	if err != nil {
		return []string{}, err
	}

	// fetch aliases
	projectAliases, err := aliasSrvc.GetByUserAndType(user.ID, models.SummaryProject)
	if err != nil {
		return []string{}, err
	}

	projects := datastructure.New[string](realProjects...)

	// remove alias values (source of a mapping)
	// add alias key (target of a mapping) instead
	for _, a := range projectAliases {
		projects.Delete(a.Value)
		projects.Add(a.Key)
	}

	sorted := projects.Values()
	sort.Strings(sorted)
	return sorted, nil
}
