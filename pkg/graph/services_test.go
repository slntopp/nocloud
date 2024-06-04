package graph_test

import (
	"context"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/arangodb/go-driver"
	ipb "github.com/slntopp/nocloud-proto/instances"
	pb "github.com/slntopp/nocloud-proto/services"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	log  *zap.Logger
	db   driver.Database
	ctrl graph.ServicesController
)

func init() {
	viper.SetDefault("DB_HOST", "localhost:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	arangodbHost := viper.GetString("DB_HOST")
	arangodbCred := viper.GetString("DB_CRED")

	viper.Set("LOG_LEVEL", -1)
	log = nocloud.NewLogger()
	log.Info("Setting up DB Connection")
	db = connectdb.MakeDBConnection(log, arangodbHost, arangodbCred)
	log.Info("DB connection established")

	ctrl = graph.NewServicesController(log, db, nil)
}

func TestCreate(t *testing.T) {
	service := &pb.Service{
		Version: "1",
		Title:   randomdata.SillyName(),
		InstancesGroups: []*ipb.InstancesGroup{
			{
				Type:  "ione",
				Title: randomdata.SillyName(),
				Instances: []*ipb.Instance{
					{
						Title: randomdata.SillyName(),
					},
					{
						Title: randomdata.SillyName(),
					},
				},
			},
		},
	}

	srv, err := ctrl.Create(context.Background(), service)
	if err != nil {
		t.Fatalf("Error creating service: %v", err)
	}
	log.Info("Result Service", zap.Any("service", srv))
}
