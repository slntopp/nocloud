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

func TestProcessTransactions(t *testing.T) {
}

func TestGenerateTransactions(t *testing.T) {
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

	ctx := context.Background()

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

	accountConroller.Update(ctx, acc, map[string]interface{}{
		"balance": 2.0,
	})

	recordsController.Create(ctx, &pb.Record{
		Start:     time.Now().Add(-2 * time.Hour).Unix(),
		End:       time.Now().Add(-time.Hour).Unix(),
		Exec:      time.Now().Add(-time.Hour).Unix(),
		Resource:  "meme",
		Total:     1.0,
		Currency:  pb.Currency_NCU,
		Instance:  instanceMeta.Key,
		Processed: false,
	})

	billingServer.GenTransactions(ctx, log, time.Now(), CurrencyConf{
		Currency: int(pb.Currency_NCU),
	})

	acc, err = accountConroller.Get(ctx, acc.ID.Key())
	if err != nil {
		t.Error(err)
	}
	if acc.GetBalance() != 1.0 {
		t.Error("Got wrong balance")
	}
}
