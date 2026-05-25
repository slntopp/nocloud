package billing

import (
	"context"
	"fmt"
	"time"

	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"connectrpc.com/connect"
)

const (
	payWithBalanceLockKeyPrefix = "billing:pay_with_balance:"
	payWithBalanceLockTTL       = 2 * time.Minute
	payWithBalanceLockMaxWait   = 2 * time.Minute
	payWithBalanceLockRetryMin  = 25 * time.Millisecond
	payWithBalanceLockRetryMax  = 400 * time.Millisecond
)

const releasePayWithBalanceLockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
end
return 0
`

func payWithBalanceRedisLockKey(accountID string) string {
	return payWithBalanceLockKeyPrefix + accountID
}

func (s *BillingServiceServer) withPayWithBalanceLock(
	ctx context.Context,
	accountID string,
	log *zap.Logger,
	fn func(context.Context) (*connect.Response[pb.PayWithBalanceResponse], error),
) (*connect.Response[pb.PayWithBalanceResponse], error) {
	key := payWithBalanceRedisLockKey(accountID)
	token := uuid.New().String()
	deadline := time.Now().Add(payWithBalanceLockMaxWait)
	retryAfter := payWithBalanceLockRetryMin

	for {
		if err := ctx.Err(); err != nil {
			return nil, status.Error(codes.Canceled, err.Error())
		}

		ok, err := s.rdb.SetNX(ctx, key, token, payWithBalanceLockTTL).Result()
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("redis lock error: %v", err))
		}
		if ok {
			defer func() {
				relCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if _, relErr := s.rdb.Eval(relCtx, releasePayWithBalanceLockScript, []string{key}, token).Result(); relErr != nil {
					log.Warn("failed to release pay-with-balance redis lock", zap.String("key", key), zap.Error(relErr))
				}
			}()
			return fn(ctx)
		}

		if time.Now().After(deadline) {
			return nil, status.Error(codes.Unavailable, "Pay with balance is busy for this account; please retry shortly")
		}

		timer := time.NewTimer(retryAfter)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, status.Error(codes.Canceled, ctx.Err().Error())
		case <-timer.C:
		}
		if retryAfter < payWithBalanceLockRetryMax {
			retryAfter += 25 * time.Millisecond
		}
	}
}
