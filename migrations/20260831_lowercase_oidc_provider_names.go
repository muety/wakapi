package migrations

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type oidcProviderNameConflict struct {
	Provider string
	Variants int64
}

func init() {
	const name = "20260831-lowercase_oidc_provider_names"
	f := migrationFunc{
		name:       name,
		background: false,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if err := lowercaseOidcProviderNames(db, name); err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}

func lowercaseOidcProviderNames(db *gorm.DB, name string) error {
	if !db.Migrator().HasTable(&models.User{}) {
		slog.Warn("skipping migration because users table does not exist", "name", name)
		return nil
	}

	distinctExpr := "count(distinct auth_type)"
	normalizedExpr := "auth_type <> lower(auth_type)"
	if db.Dialector.Name() == "mysql" {
		// mysql default collations (e.g. utf8mb4_0900_ai_ci) are case-insensitive for string comparisons -> compare raw byte arrays instead
		distinctExpr = "count(distinct cast(auth_type as binary))"
		normalizedExpr = "cast(auth_type as binary) <> cast(lower(auth_type) as binary)"
	}

	var conflicts []oidcProviderNameConflict
	if err := db.Model(&models.User{}).
		Select(fmt.Sprintf("lower(auth_type) as provider, %s as variants", distinctExpr)).
		Where("auth_type is not null and auth_type <> ''").
		Group("lower(auth_type)").
		Having(fmt.Sprintf("%s > 1", distinctExpr)).
		Scan(&conflicts).Error; err != nil {
		return err
	}

	if len(conflicts) > 0 {
		providers := make([]string, len(conflicts))
		for i, c := range conflicts {
			providers[i] = c.Provider
		}
		return fmt.Errorf(
			"found oidc provider names in users.auth_type used with conflicting casing ('%s'), please manually reconcile the affected user accounts and restart",
			strings.Join(providers, "', '"),
		)
	}

	return db.Model(&models.User{}).
		Where(fmt.Sprintf("auth_type is not null and auth_type <> '' and %s", normalizedExpr)).
		Update("auth_type", gorm.Expr("lower(auth_type)")).
		Error
}
