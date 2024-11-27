package graph

import (
	"context"
	"testing"

	"github.com/slntopp/nocloud/pkg/credentials"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	proto "github.com/slntopp/nocloud-proto/instances"
	"github.com/slntopp/nocloud-proto/registry/accounts"
	pb "github.com/slntopp/nocloud-proto/services"
	spb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	arangodbHost string
	arangodbCred string
	redisHost    string
	log          *zap.Logger
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("DB_HOST", "localhost:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
}

func TestDeleteAccount(t *testing.T) {
	ctx := context.TODO()
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred, schema.DB_NAME)
	db.Info(ctx)

	ac := NewAccountsController(log, db)
	nsc := NewNamespacesController(log, db)
	//instc := NewInstancesController(log, db, nil)
	//igc := NewInstancesGroupsController(log, db, nil)
	spc := NewServicesProvidersController(log, db)
	srvc := NewServicesController(log, db, nil)

	acc, err := ac.Create(ctx, accounts.Account{Title: "test_user"})
	if err != nil {
		t.Error("Can't create account")
	}

	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, acc.ID.Key())

	namespace, err := nsc.Create(ctx, "test_namespace")
	if err != nil {
		t.Error("Can't create namespace")
	}

	if err := nsc.Join(ctx, acc, namespace, access.Level_ADMIN, roles.OWNER); err != nil {
		t.Error("Can't join namespace")
	}

	sp := &ServicesProvider{
		ServicesProvider: &spb.ServicesProvider{},
	}
	err = spc.Create(ctx, sp)
	if err != nil {
		t.Error("Can't create sp")
	}

	service, err := srvc.Create(ctx, &pb.Service{
		InstancesGroups: []*proto.InstancesGroup{
			{
				Title: "test",
				Sp:    &sp.Uuid,
				Instances: []*proto.Instance{
					{
						Title: "test",
					},
				},
			},
		},
	})
	if err != nil {
		t.Error("Can't create service")
	}

	if err := srvc.Join(ctx, service, &namespace, access.Level_ADMIN, roles.OWNER); err != nil {
		t.Error("Can't join service")
	}

	ac.Delete(ctx, acc.Uuid)

	//ig := service.GetInstancesGroups()[0]
	//inst := ig.GetInstances()[0]
	//if _, err := igc.col.ReadDocument(ctx, ig.GetUuid(), &proto.InstancesGroup{}); err == nil {
	//	t.Error("Found orphan instances groups")
	//}
	//if _, err := instc.col.ReadDocument(ctx, inst.GetUuid(), &proto.Instance{}); err == nil {
	//	t.Error("Found orphan instance")
	//}
}

func TestDeleteAccountCredentials(t *testing.T) {
	ctx := context.TODO()
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred, schema.DB_NAME)
	db.Info(ctx)

	ac := NewAccountsController(log, db)
	nsc := NewNamespacesController(log, db)
	spc := NewServicesProvidersController(log, db)
	srvc := NewServicesController(log, db, nil)

	acc, err := ac.Create(ctx, accounts.Account{Title: "test_user"})
	if err != nil {
		t.Error("Can't create account")
	}

	cred_edge_col, _ := db.Collection(context.TODO(), schema.ACC2CRED)
	cred, _ := credentials.NewStandardCredentials([]string{"test_user", "test_user"})

	err = ac.SetCredentials(ctx, acc, cred_edge_col, cred, roles.OWNER)
	if err != nil {
		t.Error("Can't set credentials")
	}

	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, acc.ID.Key())

	namespace, err := nsc.Create(ctx, "test_namespace")
	if err != nil {
		t.Error("Can't create namespace")
	}

	if err := nsc.Join(ctx, acc, namespace, access.Level_ADMIN, roles.OWNER); err != nil {
		t.Error("Can't join namespace")
	}

	sp := &ServicesProvider{
		ServicesProvider: &spb.ServicesProvider{},
	}
	err = spc.Create(ctx, sp)
	if err != nil {
		t.Error("Can't create sp")
	}

	service, err := srvc.Create(ctx, &pb.Service{
		InstancesGroups: []*proto.InstancesGroup{
			{
				Title: "test",
				Sp:    &sp.Uuid,
				Instances: []*proto.Instance{
					{
						Title: "test",
					},
				},
			},
		},
	})
	if err != nil {
		t.Error("Can't create service")
	}

	if err := srvc.Join(ctx, service, &namespace, access.Level_ADMIN, roles.OWNER); err != nil {
		t.Error("Can't join service")
	}

	ac.Delete(ctx, acc.Uuid)

	// Check if Credentials graph is empty as well

	graph := GraphGetEnsure(log, ctx, db, schema.CREDENTIALS_GRAPH.Name)
	acc2cred := GraphGetEdgeEnsure(log, ctx, graph, schema.ACC2CRED, schema.ACCOUNTS_COL, schema.CREDENTIALS_COL)

	cursor, err := acc2cred.Database().Query(ctx,
		`FOR edge in @@col 
			FILTER edge._from == @acc
			RETURN edge`,
		map[string]interface{}{
			"@col": schema.ACC2CRED,
			"acc":  driver.NewDocumentID(schema.ACCOUNTS_COL, acc.Uuid),
		})
	if err != nil {
		t.Error("Unexpected error while credentials query")
	}
	if cursor.Count() != 0 {
		t.Error("Found orphan credential nodes")
	}
}
