package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

func init() {
	const name = "20260820-fix_ai_model_as_editor"
	f := migrationFunc{
		name:       name,
		background: true,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			// See https://github.com/muety/wakapi/issues/962
			// Fix at least the most common ones ...
			if err := db.Exec(`
				update heartbeats
				set editor = case
					when user_agent like '%claude-code%' then 'Claude-code'
					when user_agent like '%codex-cli%' then 'Codex-cli'
					when user_agent like '%opencode-cli%' then 'Opencode-cli'
					when user_agent like '%OpenCode%' then 'Opencode'
					else editor
				end
				where category = 'ai coding'
				  and (
					user_agent like '%claude-code%'
					or user_agent like '%codex-cli%'
					or user_agent like '%opencode-cli%'
					or user_agent like '%OpenCode%'
				  )
			`).Error; err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
