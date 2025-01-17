package lifecycle

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const terminationDelay = time.Second * 5

func ListenForApplicationShutDown(ctx context.Context, shutdownFunc func(), signalChannel chan os.Signal) {
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)

	select {
	case sig := <-signalChannel:
		slog.With("sig", sig.String()).Warn("shutdown signal received")
		shutdownFunc()
		time.Sleep(terminationDelay)
	case <-ctx.Done():
		slog.Warn("context deadline exceeded or canceled")
		shutdownFunc()
		time.Sleep(terminationDelay)
	}
}
