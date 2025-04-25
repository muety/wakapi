package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

// In the context if https://github.com/muety/wakapi/issues/777, we retroactively added a primary key column to the durations table.
// However, SQLite doesn't allow to alter an existing table that way. Workaround is to create a new one and copy its contents.

func init() {
	const name = "20250425-add_durations_primary_key"
	f := migrationFunc{
		name:       name,
		background: true,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if cfg.Db.Dialect != config.SQLDialectSqlite {
				return nil
			}
			if !db.Migrator().HasTable("durations") {
				return nil
			}
			if db.Migrator().HasColumn(&models.Duration{}, "id") {
				return nil
			}

			if err := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Migrator().RenameTable("durations", "durations_old"); err != nil {
					return err
				}
				if err := tx.Migrator().DropIndex(&models.Duration{}, "idx_time_duration_user"); err != nil {
					return err
				}
				if err := tx.Migrator().CreateTable(&models.Duration{}); err != nil {
					return err
				}
				if err := tx.Exec("insert into durations(user_id, time, duration, project, language, editor, operating_system, machine, category, branch, entity, num_heartbeats, group_hash, timeout) select * from durations_old").Error; err != nil {
					return err
				}
				if err := tx.Migrator().DropTable("durations_old"); err != nil {
					return err
				}
				return nil
			}); err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPreMigration(f)
}
