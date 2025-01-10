package migrations

import (
    "github.com/muety/wakapi/config"
    "github.com/muety/wakapi/models"
    "gorm.io/gorm"
)

func init() {
    const name = "20241220-share_activity_chart_flag"
    f := migrationFunc{
        name: name,
        f: func(db *gorm.DB, cfg *config.Config) error {
            if hasRun(name, db) {
                return nil
            }

            db.
                Model(&models.User{}).
                Where("share_data_max_days < ?", 0).
                Or("share_data_max_days >= ?", 365).
                Update("share_activity_chart", true)

            setHasRun(name, db)
            return nil
        },
    }

    registerPostMigration(f)
}
