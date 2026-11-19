package billing

import (
	"context"
	"testing"

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
)

func TestSuspensionAction(t *testing.T) {
	cases := []struct {
		name      string
		balance   float64
		suspended bool
		limit     float64
		want      suspensionChange
	}{
		{"below limit suspends", 5, false, 10, suspendAccount},
		{"above limit unsuspends", 15, true, 10, unsuspendAccount},
		{"at limit keeps active", 10, false, 10, keepAccount},
		{"at limit keeps suspended", 10, true, 10, keepAccount},
		{"already suspended below limit", 5, true, 10, keepAccount},
		{"active above limit", 15, false, 10, keepAccount},
	}
	for _, c := range cases {
		if got := suspensionAction(c.balance, c.suspended, c.limit); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}

// Inside a database transaction the check must not reach the registry: it is
// queued for the transaction owner to run after commit.
func TestCheckAccountSuspensionDefersInsideTransaction(t *testing.T) {
	s := &BillingServiceServer{} // nil clients: any registry call would panic
	log := zap.NewNop()

	ctx, _ := s.withDeferredSuspendChecks(context.Background(), log)
	ctx = driver.WithTransactionID(ctx, "42")

	for _, acc := range []string{"a", "b", "a"} {
		if err := s.checkAccountSuspension(ctx, log, acc); err != nil {
			t.Fatalf("checkAccountSuspension(%s): %v", acc, err)
		}
	}

	d := ctx.Value(deferredSuspendChecksKey{}).(*deferredSuspendChecks)
	got := d.drain()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("deferred accounts = %v, want [a b]", got)
	}
	if left := d.drain(); len(left) != 0 {
		t.Fatalf("drain should empty the queue, got %v", left)
	}
}

// A transaction opened without the collector must skip the check rather than
// call the registry while holding the lock.
func TestCheckAccountSuspensionSkipsInUnpreparedTransaction(t *testing.T) {
	s := &BillingServiceServer{}
	ctx := driver.WithTransactionID(context.Background(), "42")
	if err := s.checkAccountSuspension(ctx, zap.NewNop(), "a"); err != nil {
		t.Fatal(err)
	}
}
