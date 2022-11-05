/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

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

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/arangodb/go-driver"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/slntopp/nocloud/pkg/access"
	"github.com/slntopp/nocloud/pkg/edge/auth"
	nograph "github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	pb "github.com/slntopp/nocloud/pkg/billing/proto"
	ipb "github.com/slntopp/nocloud/pkg/instances/proto"
	srvpb "github.com/slntopp/nocloud/pkg/services/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

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

	viper.AutomaticEnv()
	log := nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")

	viper.SetDefault("DB_HOST", "localhost:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DRIVERS", "")
	viper.SetDefault("EXTENTION_SERVERS", "")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	port := viper.GetString("PORT")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatal("Failed to listen", zap.String("address", port), zap.Error(err))
	}

	arangodbHost := viper.GetString("DB_HOST")
	arangodbCred := viper.GetString("DB_CRED")
	SIGNING_KEY := []byte(viper.GetString("SIGNING_KEY"))

	defer func() {
		_ = log.Sync()
	}()

	log.Info("Setting up DB Connection")
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred)
	log.Info("DB connection established")

	auth.SetContext(log, SIGNING_KEY)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(log),
			grpc.UnaryServerInterceptor(auth.JWT_AUTH_INTERCEPTOR),
		)),
	)

	billingServer := NewBillingServiceServer(log, db)

	pb.RegisterBillingServiceServer(s, billingServer)
	go s.Serve(lis)
	// currencyController := nograph.NewCurrencyController(log, db)
	accountConroller := nograph.NewAccountsController(log, db)
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

	acc, err := accountConroller.Create(ctx, "test_account")
	if err != nil {
		t.Error(err)
	}

	ctx = context.WithValue(context.Background(), nocloud.NoCloudAccount, acc.ID.Key())

	ns, err := nsConroller.Create(ctx, "test_routine_ns")
	if err != nil {
		t.Error(err)
	}

	nsConroller.Link(ctx, acc, ns, int32(access.Level_ADMIN), roles.OWNER)

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

	if err := srvConroller.Join(ctx, srv, &ns, int32(access.Level_ADMIN), roles.OWNER); err != nil {
		t.Error(err)
	}

	// Use i in records to prevent overlapping
	for i, tc := range testCases {
		accountConroller.Update(ctx, acc, map[string]interface{}{
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
				Currency: int(tc.UserCurrency),
			},
			RoundingConf{
				Rounding: tc.Rounding,
			},
		)

		acc, err = accountConroller.Get(ctx, acc.ID.Key())
		if err != nil {
			t.Error(err)
		}
		if acc.GetBalance() != tc.FinalBalance {
			t.Errorf("Got wrong balance. Got %f. Wanted %f.", acc.GetBalance(), tc.FinalBalance)
		}
	}

}
