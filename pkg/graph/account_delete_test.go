package graph

import (
	"context"
	"testing"

	"github.com/slntopp/nocloud/pkg/instances/proto"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/access"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	pb "github.com/slntopp/nocloud/pkg/services/proto"
	spb "github.com/slntopp/nocloud/pkg/services_providers/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	arangodbHost string
	arangodbCred string
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
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred)
	db.Info(ctx)

	ac := NewAccountsController(log, db)
	nsc := NewNamespacesController(log, db)
	instc := NewInstancesController(log, db)
	igc := NewInstancesGroupsController(log, db)
	spc := NewServicesProvidersController(log, db)
	srvc := NewServicesController(log, db)

	account, err := ac.Create(ctx, "test_user")
	if err != nil {
		t.Error("Can't create account")
	}

	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, account.ID.Key())

	namespace, err := nsc.Create(ctx, "test_namespace")
	if err != nil {
		t.Error("Can't create namespace")
	}

	if err := nsc.Join(ctx, account, namespace, access.ADMIN, roles.OWNER); err != nil {
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

	if err := srvc.Join(ctx, service, &namespace, access.ADMIN, roles.OWNER); err != nil {
		t.Error("Can't join service")
	}

	ac.Delete(ctx, account.Uuid)

	ig := service.GetInstancesGroups()[0]
	inst := ig.GetInstances()[0]
	if _, err := igc.col.ReadDocument(ctx, ig.GetUuid(), &proto.InstancesGroup{}); err == nil {
		t.Error("Found orphan instances groups")
	}
	if _, err := instc.col.ReadDocument(ctx, inst.GetUuid(), &proto.Instance{}); err == nil {
		t.Error("Found orphan instance")
	}
}
