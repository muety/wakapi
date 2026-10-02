package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

// See https://github.com/muety/wakapi/issues/979
func init() {
	const name = "20261002-fix_neovim_wakatime_editor"
	f := migrationFunc{
		name:       name,
		background: true,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if err := db.Model(&models.Heartbeat{}).
				Where("lower(editor) = ?", "wakatime.nvim").
				Or("lower(ai_model) = ? and lower(user_agent) like ?", "neovim", "%wakatime.nvim%").
				Updates(map[string]any{
					"editor":   "neovim",
					"ai_model": "",
				}).Error; err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
