package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/api"
)

var startJobsCmd = &cobra.Command{
	Use:   "start-jobs",
	Short: "Starts river jobs client.",
	Long:  `Starts river jobs client.`,
	Run: func(cmd *cobra.Command, args []string) {
		startJobs()
	},
}

func init() {
	rootCmd.AddCommand(startJobsCmd)

	startJobsCmd.Flags().StringVar(&cfgFile, "configg", conf.DefaultConfigPath, fmt.Sprintf("config file (default is %s)", conf.DefaultConfigPath))
}

func startJobs() {
	config := conf.Load(cfgFile, "0.00.01")
	api.StartApiRiverClient(config)
}
