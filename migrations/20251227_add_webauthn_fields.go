package migrations

import (
	"github.com/gofrs/uuid/v5"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

func init() {
	const name = "20260103_add_webauthn_fields"
	f := migrationFunc{
		name:       name,
		background: false,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if !db.Migrator().HasColumn(&models.User{}, "webauthn_id") {
				if err := db.Exec("ALTER TABLE users ADD COLUMN webauthn_id TEXT").Error; err != nil {
					return err
				}
			}

			var users []*models.User
			if err := db.Where("webauthn_id IS NULL OR webauthn_id = ''").Find(&users).Error; err != nil {
				return err
			}

			for _, u := range users {
				u.WebauthnID = uuid.Must(uuid.NewV4()).String()
				if err := db.Model(u).Update("webauthn_id", u.WebauthnID).Error; err != nil {
					return err
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
