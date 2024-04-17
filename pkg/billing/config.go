package billing

import (
	"context"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	sc "github.com/slntopp/nocloud/pkg/settings/client"
	"go.uber.org/zap"
)

// Settings Key storing billing configuration
const (
	monFreqKey  string = "billing-gen-transactions-routine"
	currencyKey string = "billing-platform-currency"
	roundingKey string = "billing-rounding"
	suspKey     string = "global-suspend-conf"
)

type RoutineConf struct {
	Frequency int `json:"freq"` // Frequency in Seconds
}

type CurrencyConf struct {
	Currency *pb.Currency `json:"currency"` // Default currency for platform
}

type RoundingConf struct {
	Rounding string `json:"rounding"` // Rounding used in payments
}

type SuspendSchedule struct {
	Day  int  `json:"day"`
	Off  bool `json:"off"`
	From int  `json:"from"`
	To   int  `json:"to"`
}

type SuspendConf struct {
	AutoResume     bool              `json:"auto_resume"`
	IsEnabled      bool              `json:"is_enabled"`
	Limit          float64           `json:"limit"`
	Schedule       []SuspendSchedule `json:"schedule"`
	IsExtraEnabled bool              `json:"is_extra_enabled"`
	ExtraLimit     float64           `json:"extra_limit"`
}

var (
	routineSetting = &sc.Setting[RoutineConf]{
		Value: RoutineConf{
			Frequency: 60,
		},
		Description: "Transactions Generating and Processing Routine Configuration",
		Level:       access.Level_ADMIN,
	}
	currencySetting = &sc.Setting[CurrencyConf]{
		Value: CurrencyConf{
			Currency: &pb.Currency{
				Id:    schema.DEFAULT_CURRENCY_ID,
				Title: schema.DEFAULT_CURRENCY_NAME,
			},
		},
		Description: "Default currency for platform",
		Level:       access.Level_ADMIN,
	}
	roundingSetting = &sc.Setting[RoundingConf]{
		Value: RoundingConf{
			Rounding: "CEIl",
		},
		Description: "Rounding used in records, transactions and other payments",
		Level:       access.Level_ADMIN,
	}
	suspendedSetting = &sc.Setting[SuspendConf]{
		Value: SuspendConf{
			AutoResume: true,
			IsEnabled:  true,
			Limit:      10,
			Schedule: []SuspendSchedule{
				{
					Day: 0,
					Off: true,
				},
				{
					Day:  1,
					From: 10,
					To:   22,
				},
				{
					Day:  2,
					From: 10,
					To:   22,
				},
				{
					Day:  3,
					From: 10,
					To:   22,
				},
				{
					Day:  4,
					From: 10,
					To:   22,
				},
				{
					Day:  5,
					From: 10,
					To:   22,
				},
				{
					Day: 6,
					Off: true,
				},
			},
			// IsExtraEnabled: true,
			// ExtraLimit: -100,
		},
	}
)

func MakeRoutineConf(ctx context.Context, log *zap.Logger) (conf RoutineConf) {
	sc.Setup(log, ctx, &settingsClient)

	if err := sc.Fetch(monFreqKey, &conf, routineSetting); err != nil {
		conf = routineSetting.Value
	}

	return conf
}

func MakeCurrencyConf(ctx context.Context, log *zap.Logger) (conf CurrencyConf) {
	sc.Setup(log, ctx, &settingsClient)

	if err := sc.Fetch(currencyKey, &conf, currencySetting); err != nil {
		conf = currencySetting.Value
	}

	return conf
}

func MakeRoundingConf(ctx context.Context, log *zap.Logger) (conf RoundingConf) {
	sc.Setup(log, ctx, &settingsClient)

	if err := sc.Fetch(currencyKey, &conf, roundingSetting); err != nil {
		conf = roundingSetting.Value
	}

	return conf
}

func MakeSuspendConf(ctx context.Context, log *zap.Logger) (conf SuspendConf) {
	sc.Setup(log, ctx, &settingsClient)

	if err := sc.Fetch(suspKey, &conf, suspendedSetting); err != nil {
		conf = suspendedSetting.Value
	}

	return conf
}
