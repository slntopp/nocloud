package http_server

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Serve(_log *zap.Logger, addr string, handler http.Handler, timeoutSecs ...int64) {
	log := _log.Named("Serve")

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	shutdown := make(chan struct{})

	go func() {
		log.Info("Serving", zap.String("address", addr))
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to listen and serve", zap.Error(err))
		}
		log.Info("Stopped listen and serve")
		shutdown <- struct{}{}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Got signal. Trying to stop server gracefully", zap.String("signal", (<-sig).String()))

	var stopTimeout = time.Duration(60)
	if len(timeoutSecs) > 0 {
		stopTimeout = time.Duration(timeoutSecs[0])
	}
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), stopTimeout*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warn("HTTP server forcibly stopped due to timeout", zap.Int("timeout", int(stopTimeout)))
		} else {
			log.Fatal("Failed HTTP Shutdown", zap.Error(err))
		}
	}

	<-shutdown
	log.Info("Graceful shutdown complete.")
}
