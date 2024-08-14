package sync

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

const DataUpdateKeyTemplate = "data.update.open.%s"

type DataSyncer struct {
	log      *zap.Logger
	rdb      *redis.Client
	sp       string
	retries  int
	ctx      context.Context
	interval time.Duration
}

func NewDataSyncer(log *zap.Logger, rdb *redis.Client, sp string, retries int, millisecondsInterval ...int) *DataSyncer {
	var interval = time.Duration(1000) * time.Millisecond
	if len(millisecondsInterval) > 0 {
		interval = time.Duration(millisecondsInterval[0]) * time.Millisecond
	}
	return &DataSyncer{
		rdb:      rdb,
		sp:       sp,
		retries:  retries,
		ctx:      context.Background(),
		log:      log.Named(fmt.Sprintf("DataSyncer.%s", sp)),
		interval: interval,
	}
}

func (s *DataSyncer) WaitUntilOpenedAndCloseAfter() error {
	s.log.Debug("Entering WaitUntilOpenedAndCloseAfter loop")
	if s.retries < 0 {
		s.log.Warn("Retries < 0 meaning infinite retries. Can cause infinite lock")
	}
	currentRetries := 0
	for {
		if s.IsOpened() {
			return s.Close()
		}
		currentRetries++
		if currentRetries > s.retries && s.retries >= 0 {
			s.log.Debug("Retries exceeded. Forced ending waiting loop", zap.Int("retries", s.retries))
			return s.Close()
		}
		time.Sleep(s.interval)
		go s.log.Debug("Next retry", zap.String("retry", fmt.Sprintf("%d/%d", currentRetries, s.retries)))
	}
}

func (s *DataSyncer) Open() error {
	s.log.Debug("Opening")
	if status := s.rdb.Set(s.ctx, fmt.Sprintf(DataUpdateKeyTemplate, s.sp), "true", 0); status.Err() != nil {
		s.log.Error("Failed to set redis key", zap.String("key", fmt.Sprintf(DataUpdateKeyTemplate, s.sp)), zap.Error(status.Err()))
		return status.Err()
	}
	return nil
}

func (s *DataSyncer) Close() error {
	s.log.Debug("Closing")
	if status := s.rdb.Set(s.ctx, fmt.Sprintf(DataUpdateKeyTemplate, s.sp), "false", 0); status.Err() != nil {
		s.log.Error("Failed to set redis key", zap.String("key", fmt.Sprintf(DataUpdateKeyTemplate, s.sp)), zap.Error(status.Err()))
		return status.Err()
	}
	return nil
}

func (s *DataSyncer) IsOpened() bool {
	status := s.rdb.Get(s.ctx, fmt.Sprintf(DataUpdateKeyTemplate, s.sp))
	if status.Err() != nil {
		s.log.Error("Failed to get redis key", zap.String("key", fmt.Sprintf(DataUpdateKeyTemplate, s.sp)), zap.Error(status.Err()))
		return true
	}
	//s.log.Debug("Got redis key", zap.String("key", fmt.Sprintf(DataUpdateKeyTemplate, s.sp)), zap.String("val", status.Val()))
	// Check for 'true' and empty string to check non-existence
	if status.Val() == "true" || status.Val() == "" {
		return true
	}
	return false
}
