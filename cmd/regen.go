package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/api"
	"github.com/muety/wakapi/internal/utilities"
)

var regenCmd = &cobra.Command{
	Use:   "regenerate-summaries",
	Short: "regenerates summaries",
	Long:  `regenerates summaries.`,
	Run: func(cmd *cobra.Command, args []string) {
		regenerateAllSummaries()
	},
}

func init() {
	rootCmd.AddCommand(regenCmd)

	regenCmd.Flags().StringVar(&cfgFile, "config", conf.DefaultConfigPath, fmt.Sprintf("config file (default is %s)", conf.DefaultConfigPath))
}

func regenerateAllSummaries() {
	config := conf.Load(cfgFile, "0.00.01")
	db, sqlDB, err := utilities.InitDB(config)

	if err != nil {
		conf.Log().Fatal("could not connect to database", "error", err)
		os.Exit(1)
		return
	}

	defer sqlDB.Close()

	apiInstance := api.NewAPIv1(config, db)

	fmt.Println("Regenerating all user summaries...")

	apiInstance.RegenerateAllUserSummaries()
}
