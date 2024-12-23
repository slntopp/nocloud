package billing

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	dpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	epb "github.com/slntopp/nocloud-proto/events_logging"
	hpb "github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ptr[T any](v T) *T {
	return &v
}

const DailyCronExecutionTimeKey = "billing-daily-cron-exec-time"
const DailyCronLastExecutionKey = "billing-daily-cron-last-execution"
const DailyCronLastManualExecutionKey = "billing-daily-cron-last-manual-execution"

type cronData struct {
	ExecutionTime       string `redis:"billing-daily-cron-exec-time"`
	LastExecution       int64  `redis:"billing-daily-cron-last-execution"`
	LastManualExecution int64  `redis:"billing-daily-cron-last-manual-execution"`
}

var cronNotify = make(chan context.Context, 10)

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

func (s *BillingServiceServer) DailyCronJob(globalCtx context.Context, log *zap.Logger, rootToken string, cronTime string, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.WithoutCancel(globalCtx)

	log = s.log.Named("DailyCronJob")

retry:
	if cronTime != "" {
		if err := s.rdb.Set(ctx, DailyCronExecutionTimeKey, cronTime, 0).Err(); err != nil {
			log.Error("Error setting cron execution time", zap.Error(err))
			s.cron.Status.Status = hpb.Status_INTERNAL
			s.cron.Status.Error = ptr(err.Error())
			time.Sleep(time.Second * 30)
			goto retry
		}
	}

start:
	executedManually := false
	requester := "system"
	s.cron.Status.Status = hpb.Status_RUNNING
	s.cron.Status.Error = nil

	var d cronData
	if err := s.rdb.MGet(ctx, DailyCronExecutionTimeKey, DailyCronLastExecutionKey, DailyCronLastManualExecutionKey).Scan(&d); err != nil {
		log.Error("Error getting cron data", zap.Error(err))
		s.cron.Status.Status = hpb.Status_INTERNAL
		s.cron.Status.Error = ptr(err.Error())
		time.Sleep(time.Second * 30)
		goto start
	}

	now := time.Now().In(time.UTC)
	log.Info("Cron is set up to run on time: "+d.ExecutionTime, zap.Time("now", now), zap.Time("last_execution", time.Unix(d.LastExecution, 0)),
		zap.Time("last_manual_execution", time.Unix(d.LastManualExecution, 0)))

	s.cron.LastExecution = time.Unix(d.LastExecution, 0).Format("2006-01-02T15:04:05")
	if d.LastManualExecution > d.LastExecution {
		s.cron.LastExecution = time.Unix(d.LastManualExecution, 0).Format("2006-01-02T15:04:05")
	}

	last := time.Unix(d.LastExecution, 0)
	h, m, sec, err := getValuesFromTime(d.ExecutionTime)
	if err != nil {
		log.Error("Cron data has invalid time format", zap.Error(err))
		s.cron.Status.Status = hpb.Status_INTERNAL
		s.cron.Status.Error = ptr(err.Error())
		time.Sleep(time.Second * 30)
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
	select {
	case <-globalCtx.Done():
		log.Info("Context is done. Quitting")
		return
	case <-t.C:
	case _ctx := <-cronNotify:
		requester, _ = _ctx.Value(nocloud.NoCloudAccount).(string)
		executedManually = true
		log.Info("Received notification to start daily cron job. Starting", zap.String("requester", requester))
	}
cron:
	if err = s.cronPreflightChecks(ctx, log); err != nil {
		log.Error("Cron preflight checks failed", zap.Error(err))
		s.cron.Status.Status = hpb.Status_INTERNAL
		s.cron.Status.Error = ptr(err.Error())
		time.Sleep(time.Second * 30)
		goto start
	}
	log.Info("All preflight checks passed")

	log.Info("Starting daily cron job")
	b := time.Now().Unix()
	s.dailyCronJobAction(ctx, log)
	a := time.Now().Unix()
	d.LastExecution = b
	s.cron.LastExecution = time.Unix(b, 0).Format("2006-01-02T15:04:05")
	nocloud.Log(log, &epb.Event{
		Entity:    "Cron",
		Uuid:      "Billing",
		Scope:     "database",
		Action:    "daily_cron_job",
		Requestor: requester,
		Snapshot:  &epb.Snapshot{Diff: ""},
		Ts:        b,
		Rc:        0,
	})
	log.Info("Finished daily cron job", zap.Int64("duration_seconds", a-b))
	if executedManually {
		if err = s.rdb.MSet(ctx, DailyCronLastManualExecutionKey, a).Err(); err != nil {
			log.Error("Error setting cron data for manual exec", zap.Error(err))
			s.cron.Status.Status = hpb.Status_INTERNAL
			s.cron.Status.Error = ptr(err.Error())
			time.Sleep(time.Second * 30)
			goto start
		}
	} else {
		if err = s.rdb.MSet(ctx, DailyCronLastExecutionKey, a).Err(); err != nil {
			log.Error("Error setting cron data", zap.Error(err))
			s.cron.Status.Status = hpb.Status_INTERNAL
			s.cron.Status.Error = ptr(err.Error())
			time.Sleep(time.Second * 30)
			goto start
		}
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
	s.WhmcsInvoicesSyncerCronJob(ctx, log)
	s.DeleteExpiredBalanceInvoicesCronJob(ctx, log)
	s.NotifyToUpdateOvhPricesCronJob(ctx, log)
}

func (s *BillingServiceServer) cronPreflightChecks(ctx context.Context, log *zap.Logger) error {
	log = log.Named("CronPreflightChecks")
	errGroup := &errgroup.Group{}
	errGroup.Go(func() error {
		if _, err := s.db.Info(ctx); err != nil {
			log.Error("No connection with arangodb", zap.Error(err))
			return err
		}
		return nil
	})
	errGroup.Go(func() error {
		if err := s.rdb.Ping(ctx).Err(); err != nil {
			log.Error("No connection with redis", zap.Error(err))
			return err
		}
		return nil
	})
	errGroup.Go(func() error {
		if s.rbmq.IsClosed() {
			log.Error("RabbitMQ connection is closed")
			return errors.New("RabbitMQ connection is closed")
		}
		return nil
	})
	for name, client := range s.drivers {
		errGroup.Go(func() error {
			if _, err := client.GetType(ctx, &dpb.GetTypeRequest{}); err != nil {
				log.Error("No connection with driver", zap.Error(err), zap.String("driver", name))
				return fmt.Errorf("no connection with driver %s: %w", name, err)
			}
			return nil
		})
	}
	return errGroup.Wait()
}

func delaySeconds(c int) {
	time.Sleep(time.Duration(c) * time.Second)
}

func (s *BillingServiceServer) RunDailyCronJob(ctx context.Context, _ *connect.Request[pb.RunDailyCronJobRequest]) (*connect.Response[pb.RunDailyCronJobResponse], error) {
	requester, _ := ctx.Value(nocloud.NoCloudAccount).(string)
	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights")
	}
	cronNotify <- ctx
	return connect.NewResponse(&pb.RunDailyCronJobResponse{}), nil
}
