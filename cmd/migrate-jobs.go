package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/jobs"
)

var migrateJobsCmd = &cobra.Command{
	Use:   "migrate-jobs",
	Short: "Runs pending river job migrations",
	Long:  `Runs pending river job migrations.`,
	Run: func(cmd *cobra.Command, args []string) {
		runJobMigrations()
	},
}

func init() {
	rootCmd.AddCommand(migrateJobsCmd)

	migrateCmd.Flags().StringVar(&cfgFile, "configf", conf.DefaultConfigPath, fmt.Sprintf("config file (default is %s)", conf.DefaultConfigPath))
	migrateCmd.Flags().StringVar(&cfgFile, "direction", "up", fmt.Sprintf("migration direction %s", "up"))
}

func runJobMigrations() {
	config := conf.Load(cfgFile, "0.00.01")
	err := jobs.RiverMigrate(config)
	if err != nil {
		fmt.Println("Running river job migrations")
		return
	}
	fmt.Println("All migrations run")
}
