package grpc_server

import (
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func ServeGRPC(_log *zap.Logger, s *grpc.Server, port string, timeoutSecs ...int64) {
	log := _log.Named("ServeGRPC")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatal("Failed to listen", zap.String("address", port), zap.Error(err))
	}

	shutdown := make(chan struct{})
	go func() {
		log.Info(fmt.Sprintf("Serving gRPC on 0.0.0.0:%v", port), zap.Skip())
		if err := s.Serve(lis); err != nil {
			log.Fatal("Failed to serve gRPC", zap.Error(err))
		}
		log.Info("Serve gracefully stopped")
		shutdown <- struct{}{}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Got signal. Trying to stop server gracefully", zap.String("signal", (<-sig).String()))

	stopped := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(stopped)
	}()

	var stopTimeout = time.Duration(60)
	if len(timeoutSecs) > 0 {
		stopTimeout = time.Duration(timeoutSecs[0])
	}
	t := time.NewTimer(stopTimeout * time.Second)
	select {
	case <-t.C:
		log.Warn("Forcing server stop due to timeout", zap.Int("timeout", int(stopTimeout)))
		s.Stop()
	case <-stopped:
		t.Stop()
	}
	<-shutdown
	log.Info("Graceful shutdown complete.")
}
