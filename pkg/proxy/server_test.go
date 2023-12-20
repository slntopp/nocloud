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
package proxy_test

import (
	"context"
	"fmt"
	"testing"

	pb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/proxy"
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
		ctrl.Delete(ctx, sp.ServicesProvider)
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
