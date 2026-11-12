package billing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	epb "github.com/slntopp/nocloud-proto/events"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	ps "github.com/slntopp/nocloud/pkg/pubsub"
	"github.com/slntopp/nocloud/pkg/pubsub/billing"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// AccountSpendingPeriodDays is the window Account.spending covers.
const AccountSpendingPeriodDays = 90

// Real money the account paid: PAID invoices that were not paid from balance.
// Balance top-ups are BALANCE invoices paid by gateway, so they count once and
// whatever is later bought from that balance does not. A returned invoice
// leaves PAID status and drops out by itself.
// Currency key: -1 when the invoice has none (treated as the default
// currency), 0 when the id was omitted as a zero value.
const accountSpendingSumsQuery = `
FOR inv IN @@invoices
  FILTER inv.status == @paid
  FILTER inv.payment >= @from
  FILTER inv.meta.paid_with_balance != true
  %s
  LET currency = inv.currency == null ? -1 : (inv.currency.id == null ? 0 : inv.currency.id)
  COLLECT account = inv.account, cid = currency AGGREGATE total = SUM(inv.total)
  RETURN { account, currency: cid, total }
`

// Writes only accounts whose value changed, so the daily run touches few
// documents and does not race with balance updates on the rest.
const accountSpendingUpdateQuery = `
FOR acc IN @@accounts
  %s
  LET amount = HAS(@amounts, acc._key) ? @amounts[acc._key] : 0
  FILTER acc.spending == null || acc.spending.amount != amount
      || acc.spending.currency != @currency || acc.spending.period_days != @period
  UPDATE acc WITH {
    spending: { amount: amount, currency: @currency, period_days: @period, calculated_at: @now }
  } IN @@accounts OPTIONS { mergeObjects: false }
`

const accountSpendingQueue = "account-spending"

type accountCurrencySum struct {
	Account  string  `json:"account"`
	Currency int32   `json:"currency"`
	Total    float64 `json:"total"`
}

// RecomputeAccountsSpending refreshes Account.spending for the given accounts,
// or for every account when none are given. The window slides, so the full run
// is needed daily even when nothing was paid.
func (s *BillingServiceServer) RecomputeAccountsSpending(ctx context.Context, log *zap.Logger, accounts ...string) error {
	now := time.Now()
	def := s.accountSpendingCurrency(ctx, log)

	vars := map[string]interface{}{
		"@invoices": schema.INVOICES_COL,
		"paid":      pb.BillingStatus_PAID,
		"from":      now.AddDate(0, 0, -AccountSpendingPeriodDays).Unix(),
	}
	filter := ""
	if len(accounts) > 0 {
		filter = "FILTER inv.account IN @accounts"
		vars["accounts"] = accounts
	}
	cur, err := s.db.Query(ctx, fmt.Sprintf(accountSpendingSumsQuery, filter), vars)
	if err != nil {
		return fmt.Errorf("failed to sum invoices: %w", err)
	}
	defer cur.Close()

	rates := map[int32]float64{}
	amounts := map[string]float64{}
	for cur.HasMore() {
		var row accountCurrencySum
		if _, err = cur.ReadDocument(ctx, &row); err != nil {
			return fmt.Errorf("failed to read invoice sum: %w", err)
		}
		if row.Currency < 0 || row.Currency == def.GetId() {
			amounts[row.Account] += row.Total
			continue
		}
		rate, ok := rates[row.Currency]
		if !ok {
			rate, _, err = s.currencies.GetExchangeRate(ctx, &pb.Currency{Id: row.Currency}, &pb.Currency{Id: def.GetId()})
			if err != nil {
				// Counting it as zero would quietly lower grades; fail so the
				// stored value stays as it was until the rate is fixed.
				return fmt.Errorf("no exchange rate %d -> %d: %w", row.Currency, def.GetId(), err)
			}
			rates[row.Currency] = rate
		}
		amounts[row.Account] += row.Total * rate
	}
	for acc, amount := range amounts {
		amounts[acc] = graph.Round(amount, def.GetPrecision(), def.GetRounding())
	}

	uvars := map[string]interface{}{
		"@accounts": schema.ACCOUNTS_COL,
		"amounts":   amounts,
		"currency":  def,
		"period":    AccountSpendingPeriodDays,
		"now":       now.Unix(),
	}
	filter = ""
	if len(accounts) > 0 {
		filter = "FILTER acc._key IN @keys"
		uvars["keys"] = accounts
	}
	ucur, err := s.db.Query(ctx, fmt.Sprintf(accountSpendingUpdateQuery, filter), uvars)
	if err != nil {
		return fmt.Errorf("failed to update accounts spending: %w", err)
	}
	_ = ucur.Close()
	return nil
}

// accountSpendingCurrency is the default currency with its precision and
// rounding taken from the currencies collection.
func (s *BillingServiceServer) accountSpendingCurrency(ctx context.Context, log *zap.Logger) *pb.Currency {
	def := proto.Clone(MakeCurrencyConf(log, &s.settingsClient).Currency).(*pb.Currency)
	if c, err := s.currencies.Get(ctx, def.GetId()); err == nil {
		def.Precision = c.Precision
		def.Rounding = c.Rounding
		if c.Code != "" {
			def.Code = c.Code
		}
		if c.Title != "" {
			def.Title = c.Title
		}
	}
	return def
}

func (s *BillingServiceServer) AccountSpendingCronJob(ctx context.Context, log *zap.Logger) {
	log = log.Named("AccountSpendingCronJob")
	log.Info("Starting account spending recalculation")
	b := time.Now()
	if err := s.RecomputeAccountsSpending(ctx, log); err != nil {
		log.Error("Failed to recalculate account spending", zap.Error(err))
		return
	}
	log.Info("Finished account spending recalculation", zap.Duration("took", time.Since(b)))
}

func (s *BillingServiceServer) ensureAccountSpendingIndex(ctx context.Context, log *zap.Logger) {
	col, err := s.db.Collection(ctx, schema.INVOICES_COL)
	if err != nil {
		log.Error("Failed to open invoices collection", zap.Error(err))
		return
	}
	if _, _, err = col.EnsurePersistentIndex(ctx, []string{"account", "status", "payment"}, &driver.EnsurePersistentIndexOptions{
		Name:         "account_spending",
		InBackground: true,
	}); err != nil {
		log.Error("Failed to ensure account spending index", zap.Error(err))
	}
}

// ConsumeAccountSpending recalculates an account's spending whenever one of
// its invoices changes. It has its own queue so a failure here never retries
// the invoice actions consumer. Errors are logged and dropped: the daily cron
// recalculates everyone anyway.
func (s *BillingServiceServer) ConsumeAccountSpending(log *zap.Logger, _ctx context.Context, p *ps.PubSub[*epb.Event], wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.WithoutCancel(_ctx)
	log = log.Named("ConsumeAccountSpending")

	s.ensureAccountSpendingIndex(ctx, log)

	msgs, err := p.Consume(accountSpendingQueue, ps.DEFAULT_EXCHANGE, billing.Topic("invoices"), ps.ConsumeOptions{
		Durable: true,
	})
	if err != nil {
		log.Fatal("Failed to start consumer", zap.Error(err))
		return
	}

	for {
		select {
		case <-_ctx.Done():
			log.Info("Context is done. Quitting")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Fatal("Messages channel is closed")
			}
			var event epb.Event
			if err = proto.Unmarshal(msg.Body, &event); err != nil {
				log.Error("Failed to unmarshal event. Skip", zap.Error(err))
			} else if err = s.processAccountSpendingEvent(ctx, log, &event); err != nil {
				log.Warn("Failed to recalculate account spending", zap.String("invoice", event.GetUuid()), zap.Error(err))
			}
			if err = msg.Ack(false); err != nil {
				log.Error("Failed to acknowledge the delivery", zap.Error(err))
			}
		}
	}
}

func (s *BillingServiceServer) processAccountSpendingEvent(ctx context.Context, log *zap.Logger, event *epb.Event) error {
	if event.GetKey() == billing.InvoiceCreated {
		return nil // a new invoice is not paid yet
	}
	inv, err := s.invoices.Get(ctx, event.GetUuid())
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	if inv.GetAccount() == "" {
		return nil
	}
	return s.RecomputeAccountsSpending(ctx, log, inv.GetAccount())
}
