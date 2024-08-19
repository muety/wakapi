package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"log/slog"
)

func init() {
	const name = "20210202-fix_cascade_for_alias_user_constraint"

	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			migrator := db.Migrator()

			if cfg.Db.Dialect == config.SQLDialectSqlite {
				// see 20201106_migration_cascade_constraints
				slog.Info("not attempting to drop and regenerate constraints on sqlite")
				return nil
			}

			if !migrator.HasTable(&models.KeyStringValue{}) {
				slog.Info("key-value table not yet existing")
				return nil
			}

			if hasRun(name, db) {
				return nil
			}

			if migrator.HasConstraint(&models.Alias{}, "fk_aliases_user") {
				slog.Info("dropping constraint 'fk_aliases_user'")
				if err := migrator.DropConstraint(&models.Alias{}, "fk_aliases_user"); err != nil {
					return err
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPreMigration(f)
}
