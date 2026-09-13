package utils

import (
	"errors"

	"github.com/muety/wakapi/models"
)

func CheckFilterPermissions(filters *models.Filters, user *models.User) error {
	if filters.Project.Exists() && !user.ShareProjects {
		return errors.New("user did not opt in to share entity-specific data")
	}
	if filters.Language.Exists() && !user.ShareLanguages {
		return errors.New("user did not opt in to share entity-specific data")
	}
	if filters.Editor.Exists() && !user.ShareEditors {
		return errors.New("user did not opt in to share entity-specific data")
	}
	if filters.OS.Exists() && !user.ShareOSs {
		return errors.New("user did not opt in to share entity-specific data")
	}
	if filters.Machine.Exists() && !user.ShareMachines {
		return errors.New("user did not opt in to share entity-specific data")
	}
	if filters.Label.Exists() && !user.ShareLabels {
		return errors.New("user did not opt in to share entity-specific data")
	}
	// branches are only shared in combination with a project filter, see GetBadgeParams
	if filters.Branch.Exists() && !user.ShareProjects {
		return errors.New("user did not opt in to share entity-specific data")
	}
	// there is no opt-in flag for categories, so third parties must never filter by them
	if filters.Category.Exists() {
		return errors.New("user did not opt in to share entity-specific data")
	}
	return nil
}
