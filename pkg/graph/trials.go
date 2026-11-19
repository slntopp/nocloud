package graph

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
)

// TRIALS_COL records which trials have been used. A document's key is derived
// from what the trial is scoped to, so claiming one is a single insert and two
// concurrent claims cannot both succeed.
const TRIALS_COL = "Trials"

// Trial is one used free trial.
type Trial struct {
	Key string `json:"_key,omitempty"`
	// Group is what the trial is for (by default the billing plan type), so
	// one trial covers every plan and product of that kind.
	Group string `json:"group"`
	// Scope and Value say who used it: "account" and an account UUID, or
	// "config:<key>" and that instance config value (e.g. a domain).
	Scope string `json:"scope"`
	Value string `json:"value"`

	Account  string `json:"account"`
	Instance string `json:"instance"`
	Started  int64  `json:"started"`
	Ends     int64  `json:"ends"`
}

func trialKey(group, scope, value string) string {
	sum := sha256.Sum256([]byte(group + "\x00" + scope + "\x00" + value))
	return hex.EncodeToString(sum[:])
}

type TrialsController interface {
	// Claim records the trial and reports whether it was still unused.
	Claim(ctx context.Context, trial Trial) (bool, error)
	// Release gives a claimed trial back, for when the instance it was
	// claimed for could not be started.
	Release(ctx context.Context, group, scope, value string) error
}

type trialsController struct {
	log *zap.Logger
	db  driver.Database

	once sync.Once
	col  driver.Collection
}

// NewTrialsController does not touch the database until first use.
func NewTrialsController(logger *zap.Logger, db driver.Database) TrialsController {
	return &trialsController{log: logger.Named("TrialsController"), db: db}
}

func (c *trialsController) collection(ctx context.Context) driver.Collection {
	c.once.Do(func() { c.col = GetEnsureCollection(c.log, ctx, c.db, TRIALS_COL) })
	return c.col
}

func (c *trialsController) Claim(ctx context.Context, trial Trial) (bool, error) {
	trial.Key = trialKey(trial.Group, trial.Scope, trial.Value)
	_, err := c.collection(ctx).CreateDocument(ctx, trial)
	if driver.IsConflict(err) {
		return false, nil
	}
	return err == nil, err
}

func (c *trialsController) Release(ctx context.Context, group, scope, value string) error {
	_, err := c.collection(ctx).RemoveDocument(ctx, trialKey(group, scope, value))
	if driver.IsNotFound(err) {
		return nil
	}
	return err
}
