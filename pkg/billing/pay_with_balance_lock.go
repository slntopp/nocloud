package billing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	payWithBalanceLockKeyPrefix = "billing:pay_balance_lock:"
	payWithBalanceLockTTL       = 45 * time.Second
	payWithBalanceLockPoll      = 50 * time.Millisecond
	payWithBalanceLockMaxWait   = 30 * time.Second
)

const payWithBalanceUnlockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
else
  return 0
end
`

func payWithBalanceLockToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type redisPayBalanceLocker interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
}

func (s *BillingServiceServer) payBalanceRedisLocker() (redisPayBalanceLocker, error) {
	r, ok := s.rdb.(redisPayBalanceLocker)
	if !ok {
		return nil, status.Error(codes.Internal, "redis client cannot run pay-with-balance lock")
	}
	return r, nil
}

func (s *BillingServiceServer) payWithBalanceRedisUnlock(lockKey, token string) {
	rd, err := s.payBalanceRedisLocker()
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = rd.Eval(ctx, payWithBalanceUnlockScript, []string{lockKey}, token).Result()
}

func (s *BillingServiceServer) payWithBalanceAcquireRedisLock(ctx context.Context, accountKey string) (unlock func(), err error) {
	rd, err := s.payBalanceRedisLocker()
	if err != nil {
		return nil, err
	}
	lockKey := payWithBalanceLockKeyPrefix + accountKey
	token := payWithBalanceLockToken()
	deadline := time.Now().Add(payWithBalanceLockMaxWait)
	for {
		ok, err := rd.SetNX(ctx, lockKey, token, payWithBalanceLockTTL).Result()
		if err != nil {
			return nil, status.Errorf(codes.Unavailable, "balance payment lock error: %v", err)
		}
		if ok {
			released := false
			return func() {
				if released {
					return
				}
				released = true
				s.payWithBalanceRedisUnlock(lockKey, token)
			}, nil
		}
		if time.Now().After(deadline) {
			return nil, status.Error(codes.ResourceExhausted, "Another balance payment is in progress for this account; try again shortly")
		}
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil, status.Error(codes.Canceled, ctx.Err().Error())
			}
			return nil, status.Error(codes.DeadlineExceeded, ctx.Err().Error())
		case <-time.After(payWithBalanceLockPoll):
		}
	}
}
