package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

func init() {
	const name = "20251005-drop_duplicate_emails"
	f := migrationFunc{
		name:       name,
		background: false,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			// til: https://chat.mistral.ai/chat/86e338b7-dad7-4478-8950-11efbf94aa4d
			const q = "update users " +
				"set email = null " +
				"where users.id in " +
				"(select id from (select id, row_number() over (partition by email order by id) as row_num from users where email is not null) as t1 " +
				"where row_num > 1);"

			if err := db.Exec(q).Error; err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
