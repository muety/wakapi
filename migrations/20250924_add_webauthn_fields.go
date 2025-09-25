package migrations

import (
	"log/slog"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	f := migrationFunc{
		name: "20240924-add_webauthn_fields",
		f:    migration20240924_add_webauthn_fields,
	}
	registerPreMigration(f)
}

func migration20240924_add_webauthn_fields(db *gorm.DB, cfg *config.Config) error {
	migrator := db.Migrator()
	
	if migrator.HasTable(&models.User{}) {
		slog.Info("adding webauthn fields to users table")
		
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
}