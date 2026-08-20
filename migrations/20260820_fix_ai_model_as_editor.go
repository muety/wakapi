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
			if err := db.Exec(`
				update heartbeats
				set editor = case
					when lower(editor) in ('opus', 'fable', 'mythos', 'sonnet', 'haiku') and lower(user_agent) like '%claude-code/%' then 'Claude'
					when lower(editor) in ('gpt') and lower(user_agent) like '%codex-cli/%' then 'Codex-cli'
					else editor
				end
				where (lower(editor) in ('opus', 'fable', 'mythos', 'sonnet', 'haiku') and lower(user_agent) like '%claude-code/%') or (lower(editor) in ('gpt') and lower(user_agent) like '%codex-cli/%')
			`).Error; err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
