package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

func init() {
	const name = "20210411-drop_migrations_table"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			migrator := db.Migrator()

			if !migrator.HasTable("gorp_migrations") {
				return nil
			}

			logbuch.Info("dropping table 'gorp_migrations'")
			return migrator.DropTable("gorp_migrations")
		},
	}

	registerPostMigration(f)
}
