package billing

import (
	"context"
	"sync"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	accounts "github.com/slntopp/nocloud-proto/registry/accounts"
	"go.uber.org/zap"
)

// Suspension is decided by the registry from the account's committed balance,
// and applied by the registry writing to the account. Neither works while
// billing still holds a database transaction on Accounts: the registry reads
// the balance from before the transaction and its write waits on billing's
// lock until ArangoDB aborts the idle transaction. So inside a transaction the
// check is deferred until the owner of the transaction has committed.

type deferredSuspendChecksKey struct{}

type deferredSuspendChecks struct {
	mu       sync.Mutex
	accounts []string
}

func (d *deferredSuspendChecks) add(account string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, a := range d.accounts {
		if a == account {
			return
		}
	}
	d.accounts = append(d.accounts, account)
}

func (d *deferredSuspendChecks) drain() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := d.accounts
	d.accounts = nil
	return out
}

// withDeferredSuspendChecks must be called by whoever opens a database
// transaction around balance changes. The returned function is to be called
// once that transaction is committed.
func (s *BillingServiceServer) withDeferredSuspendChecks(ctx context.Context, log *zap.Logger) (context.Context, func()) {
	d := &deferredSuspendChecks{}
	ctx = context.WithValue(ctx, deferredSuspendChecksKey{}, d)
	return ctx, func() {
		for _, account := range d.drain() {
			if err := s.reconcileAccountSuspension(ctx, log, account); err != nil {
				log.Error("Failed to reconcile account suspension after commit", zap.String("account", account), zap.Error(err))
			}
		}
	}
}

// checkAccountSuspension suspends or unsuspends the account according to its
// balance, now if no database transaction is open and after commit otherwise.
func (s *BillingServiceServer) checkAccountSuspension(ctx context.Context, log *zap.Logger, account string) error {
	if _, inTx := driver.HasTransactionID(ctx); inTx {
		if d, ok := ctx.Value(deferredSuspendChecksKey{}).(*deferredSuspendChecks); ok {
			d.add(account)
			return nil
		}
		log.Warn("Balance changed inside a database transaction that does not defer suspension checks; skipping the check",
			zap.String("account", account))
		return nil
	}
	return s.reconcileAccountSuspension(ctx, log, account)
}

func (s *BillingServiceServer) reconcileAccountSuspension(ctx context.Context, log *zap.Logger, account string) error {
	currencyConf := MakeCurrencyConf(log, &s.settingsClient)
	suspConf := MakeSuspendConf(log, &s.settingsClient)

	dbAcc, err := s.accClient.Get(ctx, &accounts.GetRequest{Uuid: account, Public: false})
	if err != nil {
		log.Error("Failed to get account", zap.String("err", err.Error()))
		return err
	}

	var cur *pb.Currency
	if dbAcc.Currency == nil {
		cur = currencyConf.Currency
	} else {
		cur = dbAcc.Currency
	}

	var rate float64 = 1
	if cur.GetId() != currencyConf.Currency.GetId() {
		rate, _, err = s.currencies.GetExchangeRate(ctx, cur, currencyConf.Currency)
		if err != nil {
			log.Error("Failed to get exchange rate", zap.String("err", err.Error()))
			return err
		}
	}

	switch suspensionAction(dbAcc.GetBalance()*rate, dbAcc.GetSuspended(), suspConf.Limit) {
	case suspendAccount:
		if _, err = s.accClient.Suspend(ctx, &accounts.SuspendRequest{Uuid: account}); err != nil {
			log.Error("Failed to suspend account", zap.String("err", err.Error()))
			return err
		}
	case unsuspendAccount:
		if _, err = s.accClient.Unsuspend(ctx, &accounts.UnsuspendRequest{Uuid: account}); err != nil {
			log.Error("Failed to unsuspend account", zap.String("err", err.Error()))
			return err
		}
	}
	return nil
}

type suspensionChange int

const (
	keepAccount suspensionChange = iota
	suspendAccount
	unsuspendAccount
)

func suspensionAction(balance float64, suspended bool, limit float64) suspensionChange {
	switch {
	case !suspended && balance < limit:
		return suspendAccount
	case suspended && balance > limit:
		return unsuspendAccount
	}
	return keepAccount
}
