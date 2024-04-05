/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package billing

/*
import (
	"context"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/arangodb/go-driver"
	"github.com/go-redis/redis/v8"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	ipb "github.com/slntopp/nocloud-proto/instances"
	accpb "github.com/slntopp/nocloud-proto/registry/accounts"
	srvpb "github.com/slntopp/nocloud-proto/services"
	nograph "github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

var (
	log           *zap.Logger
	port          string
	arangodbHost  string
	arangodbCred  string
	redisHost     string
	SIGNING_KEY   []byte
	db            driver.Database
	billingServer *BillingServiceServer
)

func init() {

	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")

	viper.SetDefault("REDIS_HOST", "redis:6379")
	viper.SetDefault("DB_HOST", "localhost:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DRIVERS", "")
	viper.SetDefault("EXTENTION_SERVERS", "")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	port = viper.GetString("PORT")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	redisHost = viper.GetString("REDIS_HOST")
	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatal("Failed to listen", zap.String("address", port), zap.Error(err))
	}

	defer func() {
		_ = log.Sync()
	}()

	log.Info("Setting up DB Connection")
	db = connectdb.MakeDBConnection(log, arangodbHost, arangodbCred)
	log.Info("DB connection established")

	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost,
		DB:   0,
	})

	auth.SetContext(log, rdb, SIGNING_KEY)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(log),
			grpc.UnaryServerInterceptor(auth.JWT_AUTH_INTERCEPTOR),
		)),
	)

	billingServer = NewBillingServiceServer(log, db)

	pb.RegisterBillingServiceServer(s, billingServer)
	go s.Serve(lis)

}

func TestGenerateTransactions(t *testing.T) {
	testCases := []struct {
		Rounding       string
		UserCurrency   pb.Currency
		RecCurrency    pb.Currency
		TotalPrice     float64
		FinalBalance   float64
		InitialBalance float64
	}{
		// EUR -> USD: 2.0
		// USD -> NCU: 2.0
		// EUR -> USD -> NCU: 2.0 * 2.0 = 4.0
		// Unspecified exchange rates should be defaulted to 1.0
		{Rounding: "CEIL", UserCurrency: pb.Currency_NCU, RecCurrency: pb.Currency_NCU, TotalPrice: 1.0, FinalBalance: 0.0, InitialBalance: 1.0},
		{Rounding: "CEIL", UserCurrency: pb.Currency_NCU, RecCurrency: pb.Currency_USD, TotalPrice: 1.0, FinalBalance: 0.0, InitialBalance: 2.0},
		{Rounding: "FLOOR", UserCurrency: pb.Currency_NCU, RecCurrency: pb.Currency_NCU, TotalPrice: 0.5, FinalBalance: 1.0, InitialBalance: 1.0},
		{Rounding: "ROUND", UserCurrency: pb.Currency_USD, RecCurrency: pb.Currency_USD, TotalPrice: 0.5, FinalBalance: 0.0, InitialBalance: 1.0},
		{Rounding: "CEIL", UserCurrency: pb.Currency_NCU, RecCurrency: pb.Currency_EUR, TotalPrice: 1.0, FinalBalance: 1.0, InitialBalance: 5.0},
		{Rounding: "CEIL", UserCurrency: pb.Currency_NCU, RecCurrency: pb.Currency_PLN, TotalPrice: 1.0, FinalBalance: 4.0, InitialBalance: 5.0},
	}
	// currencyController := nograph.NewCurrencyController(log, db)
	accountController := nograph.NewAccountsController(log, db)
	recordsController := nograph.NewRecordsController(log, db)
	nsConroller := nograph.NewNamespacesController(log, db)
	srvConroller := nograph.NewServicesController(log, db)
	currencyController := nograph.NewCurrencyController(log, db)

	ctx := context.Background()

	currencyController.DeleteExchangeRate(ctx, pb.Currency_NCU, pb.Currency_USD)
	currencyController.DeleteExchangeRate(ctx, pb.Currency_USD, pb.Currency_NCU)
	currencyController.DeleteExchangeRate(ctx, pb.Currency_EUR, pb.Currency_NCU)
	currencyController.DeleteExchangeRate(ctx, pb.Currency_EUR, pb.Currency_USD)

	currencyController.CreateExchangeRate(ctx, pb.Currency_USD, pb.Currency_NCU, 2.0)
	currencyController.CreateExchangeRate(ctx, pb.Currency_EUR, pb.Currency_USD, 2.0)

	acc, err := accountController.Create(ctx, accpb.Account{Title: "test_user"})
	if err != nil {
		t.Error(err)
	}

	ctx = context.WithValue(context.Background(), nocloud.NoCloudAccount, acc.ID.Key())

	ns, err := nsConroller.Create(ctx, "test_routine_ns")
	if err != nil {
		t.Error(err)
	}

	nsConroller.Link(ctx, acc, ns, access.Level_ADMIN, roles.OWNER)

	sp := "biliboba"
	srv, err := srvConroller.Create(ctx, &srvpb.Service{
		Version: "1",
		Title:   randomdata.SillyName(),
		InstancesGroups: []*ipb.InstancesGroup{
			{
				Type:  "ione",
				Sp:    &sp,
				Title: randomdata.SillyName(),
				Instances: []*ipb.Instance{
					{
						Title: randomdata.SillyName(),
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
	}

	instance := &ipb.Instance{}
	c, err := db.Query(ctx, `
	FOR instance in 2 OUTBOUND @service
	GRAPH @permissions
	FILTER IS_SAME_COLLECTION(@@instances, instance)
		RETURN instance
	`, map[string]interface{}{
		"@instances":  schema.INSTANCES_COL,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"service":     driver.NewDocumentID(schema.SERVICES_COL, srv.GetUuid()),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !c.HasMore() {
		t.Fatal("Query hasn't found new instance")
	}
	instanceMeta, err := c.ReadDocument(ctx, instance)
	if err != nil {
		t.Error(err)
	}

	if err := srvConroller.Join(ctx, srv, &ns, access.Level_ADMIN, roles.OWNER); err != nil {
		t.Error(err)
	}

	// Use i in records to prevent overlapping
	for i, tc := range testCases {
		accountController.Update(ctx, acc, map[string]interface{}{
			"balance": tc.InitialBalance,
		})

		recordsController.Create(ctx, &pb.Record{
			Exec:      time.Now().Add(time.Duration(-i-1) * time.Hour).Unix(),
			Resource:  "meme",
			Total:     tc.TotalPrice,
			Currency:  tc.RecCurrency,
			Instance:  instanceMeta.Key,
			Processed: false,
		})

		billingServer.GenTransactions(ctx, log, time.Now(),
			CurrencyConf{
				Currency: int32(tc.UserCurrency),
			},
			RoundingConf{
				Rounding: tc.Rounding,
			},
		)

		acc, err = accountController.Get(ctx, acc.ID.Key())
		if err != nil {
			t.Error(err)
		}
		if acc.GetBalance() != tc.FinalBalance {
			t.Errorf("Got wrong balance. Got %f. Wanted %f.", acc.GetBalance(), tc.FinalBalance)
		}
	}
}

func TestReprocessTransactions(t *testing.T) {
	total := 10.0
	rate := 2.1
	amount := 4
	accountUuid := randomdata.RandStringRunes(10)

	ctx := context.TODO()
	graph := nograph.GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	acccol := nograph.GraphGetVertexEnsure(log, ctx, db, graph, schema.ACCOUNTS_COL)
	nscol := nograph.GraphGetVertexEnsure(log, ctx, db, graph, schema.NAMESPACES_COL)
	currencyController := nograph.NewCurrencyController(log, db)
	trController := nograph.NewTransactionsController(log, db)

	for i := 0; i < amount; i++ {
		trController.Create(ctx, &pb.Transaction{
			Account:  accountUuid,
			Currency: pb.Currency_NCU,
			Total:    10,
		})
	}

	currencyController.DeleteExchangeRate(ctx, pb.Currency_NCU, pb.Currency_USD)
	if err := currencyController.CreateExchangeRate(ctx, pb.Currency_NCU, pb.Currency_USD, rate); err != nil {
		t.Fatal(err)
	}

	acccol.CreateDocument(ctx, map[string]interface{}{
		"_key":     accountUuid,
		"currency": pb.Currency_USD,
	})

	nscol.CreateDocument(ctx, map[string]interface{}{
		"_key": schema.ROOT_NAMESPACE_KEY,
	})

	acc2ns := nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.ACC2NS, schema.ACCOUNTS_COL, schema.NAMESPACES_COL)
	if _, err := acc2ns.CreateDocument(ctx, &nograph.Access{
		From:  driver.NewDocumentID(schema.ACCOUNTS_COL, accountUuid),
		To:    driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY),
		Role:  roles.OWNER,
		Level: access.Level_ROOT,
	}); err != nil {
		t.Fatal(err)
	}

	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, accountUuid)
	if _, err := billingServer.Reprocess(ctx, &pb.ReprocessTransactionsRequest{Account: accountUuid}); err != nil {
		t.Fatal(err)
	}

	account := &accpb.Account{}
	acccol.ReadDocument(ctx, accountUuid, account)
	want := -float64(amount) * total * rate
	got := account.GetBalance()
	if got != want {
		t.Errorf("Wrong balance: got %f, wanted %f\n", got, want)
	}
}
*/
