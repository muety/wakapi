package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/api"
	"github.com/muety/wakapi/internal/utilities"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var newServeCmd = &cobra.Command{
	Use:  "new-serve",
	Long: "Start API server",
	Run: func(cmd *cobra.Command, args []string) {
		runServer2(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(newServeCmd)

	serveCmd.Flags().StringVar(&cfgFile, "cfg", conf.DefaultConfigPath, fmt.Sprintf("config file (default is %s)", conf.DefaultConfigPath))
}

func runServer2(ctx context.Context) {
	config := conf.Load(cfgFile, "0.00.01")

	db, sqlDB, err := utilities.InitDB(config)
	if err != nil {
		conf.Log().Fatal("could not connect to database", "error", err)
		os.Exit(1)
		return
	}

	defer sqlDB.Close()

	addr := net.JoinHostPort(config.API.Host, config.API.Port)

	a := api.NewAPI(config, db)
	ah := utilities.NewAtomicHandler(a)
	logrus.WithField("version", "0.0.1").Infof("Wakana API started on: %s", addr)

	baseCtx, baseCancel := context.WithCancel(context.Background())
	defer baseCancel()

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           ah,
		ReadHeaderTimeout: 2 * time.Second, // to mitigate a Slowloris attack
		BaseContext: func(net.Listener) context.Context {
			return baseCtx
		},
	}
	log := logrus.WithField("component", "api")

	var wg sync.WaitGroup
	defer wg.Wait() // Do not return to caller until this goroutine is done.

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()

		defer baseCancel() // close baseContext

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Minute)
		defer shutdownCancel()

		if err := httpSrv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			log.WithError(err).Error("shutdown failed")
		}
	}()

	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var serr error
			if err := c.Control(func(fd uintptr) {
				serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, 0x200, 1)
			}); err != nil {
				return err
			}
			return serr
		},
	}
	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		log.WithError(err).Fatal("http server listen failed")
	}
	if err := httpSrv.Serve(listener); err != nil {
		log.WithError(err).Fatal("http server serve failed")
	}
}
