package billing

import (
	"context"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	spb "github.com/slntopp/nocloud-proto/settings"
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
	invKey      string = "billing-invoices"
)

var _ctx context.Context

func SetupSettingsContext(ctx context.Context) {
	_ctx = ctx
}

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

type InvoicesConf struct {
	Template                 string  `json:"template"`
	NewTemplate              string  `json:"new_template"`
	StartWithNumber          int     `json:"start_with_number"`
	ResetCounterMode         string  `json:"reset_counter_mode"`
	IssueRenewalInvoiceAfter float64 `json:"issue_renewal_invoice_after"`
	TopUpItemMessage         string  `json:"top_up_item_message"`
	TaxIncluded              bool    `json:"tax_included"`
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
				Id:        schema.DEFAULT_CURRENCY_ID,
				Title:     schema.DEFAULT_CURRENCY_NAME,
				Public:    false,
				Precision: 2,
				Rounding:  pb.Rounding_ROUND_HALF,
				Code:      "NCU",
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
	invoicesSetting = &sc.Setting[InvoicesConf]{
		Value: InvoicesConf{
			Template:                 "PAID {YEAR}/{MONTH}/{NUMBER}",
			NewTemplate:              "{NUMBER}",
			ResetCounterMode:         "MONTHLY",
			StartWithNumber:          0,
			IssueRenewalInvoiceAfter: 0.666,
			TopUpItemMessage:         "Пополнение баланса (услуги хостинга, оплата за сервисы)",
			TaxIncluded:              false,
		},
		Description: "Invoices configuration",
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
		Description: "Suspend configuration",
		Level:       access.Level_ADMIN,
	}
)

func MakeRoutineConf(log *zap.Logger, settingsClient *spb.SettingsServiceClient) (conf RoutineConf) {
	sc.Setup(log, _ctx, settingsClient)

	if err := sc.Fetch(monFreqKey, &conf, routineSetting); err != nil {
		conf = routineSetting.Value
	}

	return conf
}

func MakeCurrencyConf(log *zap.Logger, settingsClient *spb.SettingsServiceClient) (conf CurrencyConf) {
	sc.Setup(log, _ctx, settingsClient)

	if err := sc.Fetch(currencyKey, &conf, currencySetting); err != nil {
		conf = currencySetting.Value
	}

	log.Debug("Got currency config", zap.Any("conf", conf))

	if conf.Currency == nil {
		conf.Currency = &pb.Currency{
			Id:        schema.DEFAULT_CURRENCY_ID,
			Title:     schema.DEFAULT_CURRENCY_NAME,
			Public:    false,
			Precision: 2,
			Rounding:  pb.Rounding_ROUND_HALF,
			Code:      "NCU",
		}
	}
	return conf
}

func MakeRoundingConf(log *zap.Logger, settingsClient *spb.SettingsServiceClient) (conf RoundingConf) {
	sc.Setup(log, _ctx, settingsClient)

	if err := sc.Fetch(roundingKey, &conf, roundingSetting); err != nil {
		conf = roundingSetting.Value
	}

	return conf
}

func MakeSuspendConf(log *zap.Logger, settingsClient *spb.SettingsServiceClient) (conf SuspendConf) {
	sc.Setup(log, _ctx, settingsClient)

	if err := sc.Fetch(suspKey, &conf, suspendedSetting); err != nil {
		conf = suspendedSetting.Value
	}

	return conf
}

func MakeInvoicesConf(log *zap.Logger, settingsClient *spb.SettingsServiceClient) (conf InvoicesConf) {
	sc.Setup(log, _ctx, settingsClient)

	if err := sc.Fetch(invKey, &conf, invoicesSetting); err != nil {
		conf = invoicesSetting.Value
	}

	log.Debug("Got invoices config", zap.Any("conf", conf))
	return conf
}
