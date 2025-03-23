package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/api"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Wakana server",
	Long:  `Start the Wakana server with the specified configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&cfgFile, "config", conf.DefaultConfigPath, fmt.Sprintf("config file (default is %s)", conf.DefaultConfigPath))
}

func runServer() {
	config := conf.Load(cfgFile, "0.00.01")
	api.StartApi(config)
}
