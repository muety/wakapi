package jobs

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muety/wakapi/config"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

var EVERY_MONDAY_MORNING, _ = cron.ParseStandard("0 6 * * 1")

type Jobs struct {
	DB *gorm.DB
}

func RiverMigrate(appConf *config.Config) error {
	ctx := context.Background()
	connectionString := appConf.GetPgConnectionString()

	dbPool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		fmt.Println("Failed to get db connection:", err)
		return err
	}
	defer dbPool.Close()

	driver := riverpgxv5.New(dbPool)
	migrator, err := rivermigrate.New(driver, &rivermigrate.Config{})
	if err != nil {
		fmt.Println("failed to create migrator:", err)
		return err
	}

	printVersions := func(res *rivermigrate.MigrateResult) {
		for _, version := range res.Versions {
			fmt.Printf("Migrated [%s] version %d\n", strings.ToUpper(string(res.Direction)), version.Version)
		}
	}

	// Migrate to version 3. An actual call may want to omit all MigrateOpts,
	// which will default to applying all available up migrations.
	res, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{
		MaxSteps: 10,
	})
	if err != nil {
		return err
	}
	printVersions(res)
	return nil
}

// creates a river client against the db, registers workers and periodic jobs. requires at least one worker to be registered
func NewRiverClient(ctx context.Context, workers *river.Workers, appConf *config.Config) (*river.Client[pgx.Tx], error) {
	connectionString := appConf.GetPgConnectionString()

	dbPool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		fmt.Println("failed to connect to database:", err)
		return nil, err
	}

	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 5},
		},
		Workers: workers,
	})
	if err != nil {
		return nil, err
	}
	return riverClient, nil
}
