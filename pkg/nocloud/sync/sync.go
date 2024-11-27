package sync

import (
	"context"
	"fmt"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"go.uber.org/zap"
	"time"
)

type DataSyncer interface {
	WaitUntilOpenedAndCloseAfter() error
	Open() error
	Close() error
	IsOpened() bool
}

const DataUpdateKeyTemplate = "data.update.open.%s"

var mm = newMutexMap()

// TODO: Maybe refactor to use Redis pub/sub
type dataSyncer struct {
	log      *zap.Logger
	rdb      redisdb.Client
	sp       string
	retries  int
	ctx      context.Context
	interval time.Duration
	m        *mutexM
}

func NewDataSyncer(log *zap.Logger, rdb redisdb.Client, sp string, retries int, millisecondsInterval ...int) DataSyncer {
	var interval = time.Duration(1000) * time.Millisecond
	if len(millisecondsInterval) > 0 {
		interval = time.Duration(millisecondsInterval[0]) * time.Millisecond
	}
	return &dataSyncer{
		rdb:      rdb,
		sp:       sp,
		retries:  retries,
		ctx:      context.Background(),
		log:      log.Named(fmt.Sprintf("DataSyncer.%s", sp)),
		interval: interval,
		m:        mm,
	}
}

func (s *dataSyncer) WaitUntilOpenedAndCloseAfter() error {
	s.log.Debug("Entering WaitUntilOpenedAndCloseAfter loop")
	if s.retries < 0 {
		s.log.Warn("Retries < 0 meaning infinite retries. Can cause infinite lock")
	}
	currentRetries := 0
	for {
		s.m.Lock(s.sp)
		if s.IsOpened() {
			err := s.Close()
			s.m.Unlock(s.sp)
			return err
		}
		currentRetries++
		if currentRetries > s.retries && s.retries >= 0 {
			s.log.Debug("Retries exceeded. Forced ending waiting loop", zap.Int("retries", s.retries))
			err := s.Close()
			s.m.Unlock(s.sp)
			return err
		}
		s.m.Unlock(s.sp)
		time.Sleep(s.interval)
		s.log.Debug("Next retry", zap.String("retry", fmt.Sprintf("%d/%d", currentRetries, s.retries)))
	}
}

func (s *dataSyncer) Open() error {
	if status := s.rdb.Set(s.ctx, fmt.Sprintf(DataUpdateKeyTemplate, s.sp), "true", 0); status.Err() != nil {
		s.log.Error("Failed to set redis key", zap.String("key", fmt.Sprintf(DataUpdateKeyTemplate, s.sp)), zap.Error(status.Err()))
		return status.Err()
	}
	s.log.Debug("Opened")
	return nil
}

func (s *dataSyncer) Close() error {
	if status := s.rdb.Set(s.ctx, fmt.Sprintf(DataUpdateKeyTemplate, s.sp), "false", 0); status.Err() != nil {
		s.log.Error("Failed to set redis key", zap.String("key", fmt.Sprintf(DataUpdateKeyTemplate, s.sp)), zap.Error(status.Err()))
		return status.Err()
	}
	s.log.Debug("Closed")
	return nil
}

func (s *dataSyncer) IsOpened() bool {
	status := s.rdb.Get(s.ctx, fmt.Sprintf(DataUpdateKeyTemplate, s.sp))
	if status.Err() != nil {
		s.log.Error("Failed to get redis key", zap.String("key", fmt.Sprintf(DataUpdateKeyTemplate, s.sp)), zap.Error(status.Err()))
		return true
	}
	// Check for 'true' and empty string to check non-existence
	if status.Val() == "true" || status.Val() == "" {
		return true
	}
	return false
}
