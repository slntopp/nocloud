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
	"flag"
	"testing"
	"time"

	pb "github.com/slntopp/nocloud/pkg/billing/proto"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	log 			*zap.Logger

	arangodbHost 	string
	arangodbCred 	string
	nocloudRootPass string

	SIGNING_KEY 	[]byte
	ctrl TransactionsController

	processed chan *pb.Transaction
)

func init() {
	flag.Set("test.timeout", "10m0s")

	viper.AutomaticEnv()
	log = nocloud.NewLogger()
	
	viper.SetDefault("DB_HOST", "localhost:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("NOCLOUD_ROOT_PASSWORD", "secret")

	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	arangodbHost 	= viper.GetString("DB_HOST")
	arangodbCred 	= viper.GetString("DB_CRED")
	nocloudRootPass = viper.GetString("NOCLOUD_ROOT_PASSWORD")

	SIGNING_KEY 	= []byte(viper.GetString("SIGNING_KEY"))
}

const resetCollectionQuery = `
FOR t IN @@transactions
REMOVE t IN @@transactions
`

func TestMain(m *testing.M) {
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Setting up DB Connection")
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred)
	log.Info("DB connection established")

	ctrl = NewTransactionsController(log, db)

	ctrl.db.Query(context.TODO(), resetCollectionQuery, map[string]interface{}{
		"@transactions": schema.TRANSACTIONS_COL,
	})

	processed = make(chan *pb.Transaction)
	go ctrl.ProcessorRoutine(processed)
	m.Run()
}

func TestCreateGetAndProcess(t *testing.T) {
	now := uint64(time.Now().Unix())

	// Must be executed now
	t0, err := ctrl.Create(context.TODO(), &Transaction{
		Transaction: &pb.Transaction{
			Start: now - 3600,
			End: now,
			Exec: now,
			Instance: "111111-1111-1111-111111",
			Resource: "test",
			Total: 1,
			Meta: map[string]*structpb.Value{
				"tag": structpb.NewStringValue("exec-now"),
			},
		},
	})
	if err != nil {
		t.Fatalf("Error creating transaction: %v", err)
	}
	log.Info("Transaction created", zap.String("uuid", t0.Uuid))

	// Must not be created
	_, err = ctrl.Create(context.TODO(), &Transaction{
		Transaction: &pb.Transaction{
			Start: now - 1800,
			End: now + 1800,
			Exec: now + 1800,
			Instance: "111111-1111-1111-111111",
			Resource: "test",
			Total: 1,
			Meta: map[string]*structpb.Value{
				"tag": structpb.NewStringValue("shall-not-be-created"),
			},
		},
	})
	if err == nil {
		t.Error("Error transaction created")
	}

	// Must be executed 30 seconds later
	t1, err := ctrl.Create(context.TODO(), &Transaction{
		Transaction: &pb.Transaction{
			Start: now,
			End: now + 30,
			Exec: now + 30,
			Instance: "111111-1111-1111-111112",
			Resource: "test",
			Total: 1,
			Meta: map[string]*structpb.Value{
				"tag": structpb.NewStringValue("30s-later"),
			},
		},
	})
	if err != nil {
		t.Fatalf("Error creating transaction: %v", err)
	}
	log.Info("Transaction created", zap.String("uuid", t1.Uuid))

	for i := 0; i < 2; i++ {
		select {
		case tr := <- processed:
			if t0.Uuid == tr.Uuid || t1.Uuid == tr.Uuid {
				t.Logf("Transaction processed: %s", tr.Uuid)
			}
		case <- time.After(time.Second * 120):
			t.Error("Transactions processing timeout")
		}
	}

	t0, err = ctrl.Get(context.TODO(), &Transaction{Transaction: &pb.Transaction{Uuid: t0.Uuid}})
	if err != nil {
		t.Fatalf("Error retrieving transaction: %v", err)
	}
	if !t0.Processed {
		t.Fatalf("Error transaction not processed: %v", err)
	}

	t1, err = ctrl.Get(context.TODO(), &Transaction{Transaction: &pb.Transaction{Uuid: t1.Uuid}})
	if err != nil {
		t.Fatalf("Error retrieving transaction: %v", err)
	}
	if !t1.Processed {
		t.Fatalf("Error transaction not processed: %v", err)
	}
}
