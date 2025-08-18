package migrations

import (
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
)

func init() {
	const name = "20250818-add_start-of-week"
	f := migrationFunc{
		name:       name,
		background: true,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			// Check if column already exists
			var count int64
			if err := db.Raw("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = 'start_of_week'").Scan(&count).Error; err != nil {
				// If pragma_table_info fails (e.g., MySQL/PostgreSQL), try a different approach
				if err := db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'start_of_week'").Scan(&count).Error; err != nil {
					// If that also fails, assume column doesn't exist and proceed
					count = 0
				}
			}

			// Only add column if it doesn't exist
			if count == 0 {
				if err := db.Exec("ALTER TABLE users ADD COLUMN start_of_week INTEGER DEFAULT 1").Error; err != nil {
					return err
				}
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
