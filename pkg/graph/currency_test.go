package graph

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/slntopp/nocloud/pkg/graph/migrations"

	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func TestConvert(t *testing.T) {
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DB_HOST", "localhost:8529")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")
	viper.SetDefault("REDIS_HOST", "redis:6379")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	redisHost = viper.GetString("REDIS_HOST")

	log := nocloud.NewLogger()
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred, schema.DB_NAME)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost,
		DB:   0,
	})

	SIGNING_KEY := []byte(viper.GetString("SIGNING_KEY"))

	auth.SetContext(log, rdb, SIGNING_KEY)

	token, err := auth.MakeToken(schema.ROOT_ACCOUNT_KEY)
	if err != nil {
		log.Fatal("Can't generate token", zap.Error(err))
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "bearer "+token)

	c := NewCurrencyController(log, db)
	list, err := c.GetCurrencies(ctx, true)
	if err != nil {
		t.Error(err)
	}
	if len(list) != len(migrations.LEGACY_CURRENCIES) {
		t.Error("Default currencies haven't been created")
	}

	testRate := 2.0
	c.CreateExchangeRate(ctx, pb.Currency{Id: 1}, pb.Currency{Id: 3}, testRate, 0)
	c.CreateExchangeRate(ctx, pb.Currency{Id: 2}, pb.Currency{Id: 3}, testRate, 0)
	rates, err := c.GetExchangeRates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rates) != 2 {
		t.Error("Didn't fetch all exchange rates")
	}

	rate, _, err := c.GetExchangeRateDirect(ctx, pb.Currency{Id: 1}, pb.Currency{Id: 3})
	if err != nil {
		t.Error(err)
	}
	if rate != testRate {
		t.Error("Wrong exchange rate")
	}

	amount, err := c.Convert(ctx, &pb.Currency{Id: 1}, &pb.Currency{Id: 3}, 1.0)
	if err != nil {
		t.Error(err)
	}
	wanted := 2.0
	if amount != wanted {
		t.Errorf("Wrong conversion. Wanted %f. Got = %f", wanted, amount)
	}

	err = c.DeleteExchangeRate(ctx, &pb.Currency{Id: 1}, &pb.Currency{Id: 3})
	if err != nil {
		t.Error(err)
	}
}
