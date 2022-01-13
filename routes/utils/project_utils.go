package utils

import (
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
	projectsMap := make(map[string]bool) // proper sets as part of stdlib would be nice...

	// extract actual projects from heartbeats
	realProjects, err := heartbeatSrvc.GetEntitySetByUser(models.SummaryProject, user)
	if err != nil {
		return []string{}, err
	}

	// create a "set" / lookup table
	for _, p := range realProjects {
		projectsMap[p] = true
	}

	// fetch aliases
	projectAliases, err := aliasSrvc.GetByUserAndType(user.ID, models.SummaryProject)
	if err != nil {
		return []string{}, err
	}

	// remove alias values (source of a mapping)
	// add alias key (target of a mapping) instead
	for _, a := range projectAliases {
		if projectsMap[a.Value] {
			projectsMap[a.Value] = false
		}
		projectsMap[a.Key] = true
	}

	projects := make([]string, 0, len(projectsMap))
	for key, val := range projectsMap {
		if !val {
			continue
		}
		projects = append(projects, key)
	}

	sort.Strings(projects)
	return projects, nil
}
