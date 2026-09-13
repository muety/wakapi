package utils

import (
	"testing"

	"github.com/muety/wakapi/models"
)

func TestCheckFilterPermissions(t *testing.T) {
	sharingUser := &models.User{
		ShareProjects:  true,
		ShareLanguages: true,
	}

	tests := []struct {
		name    string
		filters *models.Filters
		user    *models.User
		wantErr bool
	}{
		{"no filters, no sharing", models.NewFiltersWith(models.SummaryProject, ""), &models.User{}, false},
		{"allowed project filter", models.NewFiltersWith(models.SummaryProject, "wakapi"), sharingUser, false},
		{"disallowed project filter", models.NewFiltersWith(models.SummaryProject, "wakapi"), &models.User{}, true},
		{"allowed language filter", models.NewFiltersWith(models.SummaryLanguage, "Go"), sharingUser, false},
		{"disallowed language filter", models.NewFiltersWith(models.SummaryLanguage, "Go"), &models.User{}, true},
		{"disallowed editor filter", models.NewFiltersWith(models.SummaryEditor, "vscode"), sharingUser, true},
		{"disallowed os filter", models.NewFiltersWith(models.SummaryOS, "linux"), sharingUser, true},
		{"disallowed machine filter", models.NewFiltersWith(models.SummaryMachine, "desktop"), sharingUser, true},
		{"disallowed label filter", models.NewFiltersWith(models.SummaryLabel, "work"), sharingUser, true},
		{"branch filter requires project sharing", models.NewFiltersWith(models.SummaryBranch, "main"), &models.User{}, true},
		{"branch filter with project sharing", models.NewFiltersWith(models.SummaryBranch, "main"), sharingUser, false},
		{"category filter is never allowed", models.NewFiltersWith(models.SummaryCategory, "coding"), sharingUser, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckFilterPermissions(tt.filters, tt.user); (err != nil) != tt.wantErr {
				t.Errorf("CheckFilterPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
