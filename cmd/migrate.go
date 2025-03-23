package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/api"
	"github.com/muety/wakapi/migrations"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Runs pending db migrations",
	Long:  `Runs pending db migrations.`,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().StringVar(&cfgFile, "config", conf.DefaultConfigPath, fmt.Sprintf("config file (default is %s)", conf.DefaultConfigPath))
}

func runMigrations() {
	config := conf.Load(cfgFile, "0.00.01")
	db, err := api.InitDB(config)
	if err != nil {
		fmt.Println("Error connecting to database")
		return
	}
	migrations.Run(db, config)
}
