package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
	"log/slog"
)

func init() {
	const name = "20210411-drop_migrations_table"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if err := db.Migrator().DropTable("gorp_migrations"); err == nil {
				slog.Info("dropped table 'gorp_migrations'")
			}
			return nil
		},
	}

	registerPostMigration(f)
}
