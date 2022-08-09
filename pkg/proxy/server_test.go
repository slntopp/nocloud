package proxy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/proxy"
	pb "github.com/slntopp/nocloud/pkg/services_providers/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	arangodbHost string
	arangodbCred string

	ctrl graph.ServicesProvidersController
)

func init() {
	viper.AddConfigPath("../../")
	viper.SetConfigType("env")

	viper.SetDefault("DB_HOST", "db.nocloud.local:80")
	viper.SetDefault("DB_CRED", "root:openSesame")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")

	log := zap.NewExample()

	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred)
	ctrl = graph.NewServicesProvidersController(log, db)

	proxy.Setup(log, ctrl)
}

func TestResolve(t *testing.T) {
	ctx := context.Background()
	sp := &graph.ServicesProvider{
		ServicesProvider: &pb.ServicesProvider{},
	}

	err := ctrl.Create(ctx, sp)
	if err != nil {
		t.Fatalf("Error creating sample ServicesProvider: %v", err)
	}

	defer func() {
		ctrl.Delete(ctx, sp.Uuid)
	}()

	sample_host := fmt.Sprintf("%s.proxy.nocloud.zone", sp.Uuid)
	r, err := proxy.Resolve(sample_host)
	if err == nil {
		t.Fatalf("Test must have failed(proxy unset), but yet resolved something: %s", r)
	}
	if err.Error() != "proxy is not defined" {
		t.Fatalf("Unexpected error while resolving: %v", err)
	}

	sp.ServicesProvider.Proxy = &pb.ProxyConf{}
	err = ctrl.Update(ctx, sp.ServicesProvider)
	if err != nil {
		t.Fatalf("Unexpected error while Updating ServicesProvider: %v", err)
	}

	r, err = proxy.Resolve(sample_host)
	if err == nil {
		t.Fatalf("Test must have failed(proxy set, but socket isn't), but yet resolved something: %s", r)
	}
	if err.Error() != "proxy is not defined" {
		t.Fatalf("Unexpected error while resolving: %v", err)
	}

	socket := "wss://sample.example.com/socket"
	sp.ServicesProvider.Proxy.Socket = &socket
	err = ctrl.Update(ctx, sp.ServicesProvider)
	if err != nil {
		t.Fatalf("Unexpected error while Updating ServicesProvider: %v", err)
	}

	r, err = proxy.Resolve(sample_host)
	if err != nil {
		t.Fatalf("Got error while resolving proxy host: %v", err)
	}
	if r != socket {
		t.Fatalf("Resolved Proxy host isn't matching defined one: %s != %s", socket, r)
	}
}
