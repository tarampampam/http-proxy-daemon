// Package serve contains CLI `serve` command implementation.
package serve

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/tarampampam/http-proxy-daemon/internal/pkg/breaker"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/config"
	appHttp "github.com/tarampampam/http-proxy-daemon/internal/pkg/http"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewCommand creates `serve` command.
func NewCommand(ctx context.Context, log *zap.Logger) *cobra.Command {
	var f flags

	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s", "server"},
		Short:   "Start HTTP server",
		Long:    "Environment variables have higher priority then flags",
		PreRunE: func(*cobra.Command, []string) error {
			if err := f.overrideUsingEnv(); err != nil {
				return err
			}

			return f.validate()
		},
		RunE: func(*cobra.Command, []string) error {
			return run(ctx, log, f.toConfig(), f.listen.ip, f.listen.port)
		},
	}

	f.init(cmd.Flags())

	return cmd
}

const serverShutdownTimeout = 5 * time.Second

// run current command.
func run(parentCtx context.Context, log *zap.Logger, cfg config.Config, ip string, port uint16) error { //nolint:funlen
	var (
		ctx, cancel = context.WithCancel(parentCtx) // serve context creation
		oss         = breaker.NewOSSignals(ctx)     // OS signals listener
	)

	// subscribe for system signals
	oss.Subscribe(func(sig os.Signal) {
		log.Warn("Stopping by OS signal..", zap.String("signal", sig.String()))

		cancel()
	})

	defer func() {
		cancel()   // call the cancellation function after all
		oss.Stop() // stop system signals listening
	}()

	// create HTTP server
	server := appHttp.NewServer(log)

	// register server routes, middlewares, etc.
	if err := server.Register(ctx, cfg); err != nil {
		return err
	}

	startingErrCh := make(chan error, 1) // channel for server starting error

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		defer close(errCh)

		log.Info("Server starting",
			zap.String("addr", ip),
			zap.Uint16("port", port),
			zap.String("proxy route prefix", cfg.Proxy.Prefix),
			zap.Duration("proxy request timeout", cfg.Proxy.RequestTimeout),
		)

		if err := server.Start(ip, port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}(startingErrCh)

	// and wait for..
	select {
	case err := <-startingErrCh: // ..server starting error
		return err

	case <-ctx.Done(): // ..or context cancellation
		log.Debug("Server stopping")

		// create context for server graceful shutdown
		ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer ctxCancelShutdown()

		// stop the server using created context above
		if err := server.Stop(ctxShutdown); err != nil {
			return err
		}
	}

	return nil
}
