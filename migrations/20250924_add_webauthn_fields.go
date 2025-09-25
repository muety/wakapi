package migrations

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20250924-add_webauthn_fields"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}
			migrator := db.Migrator()

			if err := db.Transaction(func(tx *gorm.DB) error {
				if migrator.HasTable(&models.User{}) {
					if !migrator.HasColumn(&models.User{}, "webauthn_credentials") {
						if err := migrator.AddColumn(&models.User{}, "webauthn_credentials"); err != nil {
							return err
						}
					}

					if !migrator.HasColumn(&models.User{}, "webauthn_session") {
						if err := migrator.AddColumn(&models.User{}, "webauthn_session"); err != nil {
							return err
						}
					}

					if !migrator.HasColumn(&models.User{}, "webauthn_session_expiry") {
						if err := migrator.AddColumn(&models.User{}, "webauthn_session_expiry"); err != nil {
							return err
						}
					}
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
