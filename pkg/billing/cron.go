package billing

import (
	"context"
	"fmt"
	hpb "github.com/slntopp/nocloud-proto/health"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

func ptr[T any](v T) *T {
	return &v
}

const DailyCronExecutionTimeKey = "billing-daily-cron-exec-time"
const DailyCronLastExecutionKey = "billing-daily-cron-last-execution"

type cronData struct {
	ExecutionTime string `redis:"billing-daily-cron-exec-time"`
	LastExecution int64  `redis:"billing-daily-cron-last-execution"`
}

func getValuesFromTime(t string) (hours int, minutes int, seconds int, err error) {
	trZero := func(s string) string {
		if s == "" {
			return ""
		}
		if s[0] == '0' {
			return s[1:]
		}
		return s
	}
	parts := strings.Split(t, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, 0, 0, fmt.Errorf("invalid time format, expected 'hh:mm:ss' or 'hh:mm'")
	}
	hours, err = strconv.Atoi(trZero(parts[0]))
	if err != nil || hours < 0 || hours > 23 {
		return 0, 0, 0, fmt.Errorf("invalid hours value, expected 0-23")
	}
	minutes, err = strconv.Atoi(trZero(parts[1]))
	if err != nil || minutes < 0 || minutes > 59 {
		return 0, 0, 0, fmt.Errorf("invalid minutes value, expected 0-59")
	}
	if len(parts) == 3 {
		seconds, err = strconv.Atoi(trZero(parts[2]))
		if err != nil || seconds < 0 || seconds > 59 {
			return 0, 0, 0, fmt.Errorf("invalid seconds value, expected 0-59")
		}
	} else {
		seconds = 0
	}
	return hours, minutes, seconds, nil
}

func (s *BillingServiceServer) DailyCronJob(ctx context.Context, log *zap.Logger) {
	log = s.log.Named("DailyCronJob")

start:
	s.cron.Status.Status = hpb.Status_RUNNING
	s.cron.Status.Error = nil

	var d cronData
	if err := s.rdb.MGet(ctx, DailyCronExecutionTimeKey, DailyCronLastExecutionKey).Scan(&d); err != nil {
		log.Error("Error getting cron data", zap.Error(err))
		s.cron.Status.Status = hpb.Status_INTERNAL
		s.cron.Status.Error = ptr(err.Error())
		time.Sleep(time.Second * 10)
		goto start
	}

	now := time.Now()
	last := time.Unix(d.LastExecution, 0)
	h, m, sec, err := getValuesFromTime(d.ExecutionTime)
	if err != nil {
		log.Error("Cron data has invalid time format", zap.Error(err))
		s.cron.Status.Status = hpb.Status_INTERNAL
		s.cron.Status.Error = ptr(err.Error())
		time.Sleep(time.Second * 10)
		goto start
	}

	if now.Unix() < last.Unix() {
		log.Warn("Last execution time is greater than current time", zap.Time("last", last), zap.Time("now", now))
	}

	var t *time.Ticker
	var untilNext time.Duration

	scheduledThisDay := time.Date(now.Year(), now.Month(), now.Day(), h, m, sec, 0, time.UTC)
	scheduledNextDay := scheduledThisDay.AddDate(0, 0, 1)
	if now.Day() <= last.Day() && now.Year() <= last.Year() && now.Month() <= last.Month() {
		untilNext = scheduledNextDay.Sub(now)
		t = time.NewTicker(untilNext)
	} else {
		if now.Before(scheduledThisDay) {
			untilNext = scheduledThisDay.Sub(now)
			t = time.NewTicker(untilNext)
		} else {
			log.Info("Cron wasn't running this day. Starting immediately")
			goto cron
		}
	}

	log.Info("Will be starting next cron job in "+fmt.Sprintf("%v", untilNext), zap.Duration("duration", untilNext))
	<-t.C
cron:
	log.Info("Starting daily cron job")
	b := time.Now().Unix()
	s.dailyCronJobAction(ctx, log)
	a := time.Now().Unix()
	d.LastExecution = b
	s.cron.LastExecution = time.Unix(b, 0).Format("2006-01-02T15:04:05Z07:00")
	log.Info("Finished daily cron job", zap.Int64("duration_seconds", a-b))
	if err = s.rdb.MSet(ctx, DailyCronLastExecutionKey, a).Err(); err != nil {
		log.Error("Error setting cron data", zap.Error(err))
		s.cron.Status.Status = hpb.Status_INTERNAL
		s.cron.Status.Error = ptr(err.Error())
		time.Sleep(time.Second * 10)
		goto start
	}
	goto start
}

func (s *BillingServiceServer) dailyCronJobAction(ctx context.Context, log *zap.Logger) {
	log = log.Named("Action")
	defer func() {
		if err := recover(); err != nil {
			log.Error("Recovered from panic", zap.Any("err", err))
		}
	}()
	// Jobs
	s.InvoiceExpiringInstancesCronJob(ctx, log)
}
