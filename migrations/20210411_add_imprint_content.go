package migrations

import (
	"fmt"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	const name = "20210411-add_imprint_content"
	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			condition := fmt.Sprintf("%s = ?", utils.QuoteDbIdentifier(db, "key"))

			imprintKv := &models.KeyStringValue{Key: "imprint", Value: "no content here"}
			if err := db.
				Clauses(clause.OnConflict{UpdateAll: false, DoNothing: true}).
				Where(condition, imprintKv.Key).
				Assign(imprintKv).
				Create(imprintKv).Error; err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
