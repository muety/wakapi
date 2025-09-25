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
			return addWebAuthnFields(db, name)
		},
	}
	registerPreMigration(f)
}

func addWebAuthnFields(db *gorm.DB, migrationName string) error {
	if hasRun(migrationName, db) {
		return nil
	}

	migrator := db.Migrator()
	if !migrator.HasTable(&models.User{}) {
		setHasRun(migrationName, db)
		return nil
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return addWebAuthnColumnsToUser(migrator)
	})
	if err != nil {
		return err
	}

	setHasRun(migrationName, db)
	return nil
}

func addWebAuthnColumnsToUser(migrator gorm.Migrator) error {
	columns := []string{
		"webauthn_credentials",
		"webauthn_session",
		"webauthn_session_expiry",
	}

	for _, column := range columns {
		if !migrator.HasColumn(&models.User{}, column) {
			if err := migrator.AddColumn(&models.User{}, column); err != nil {
				return err
			}
		}
	}
	return nil
}
