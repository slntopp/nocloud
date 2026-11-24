package services_providers

import (
	"context"
	"testing"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	"github.com/slntopp/nocloud-proto/registry/namespaces"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Accounts used in these tests and their access to the root namespace.
var levels = map[string]access.Level{
	"0":     access.Level_ROOT,  // root (the admin UI's "nocloud" user, and service tokens)
	"admin": access.Level_ADMIN, // an admin who is not root
	"user":  access.Level_READ,  // a plain account
}

type fakeSPCtrl struct {
	graph.ServicesProvidersController
	db map[string]*sppb.ServicesProvider
}

func (f *fakeSPCtrl) Get(_ context.Context, id string) (*graph.ServicesProvider, error) {
	sp, ok := f.db[id]
	if !ok {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return &graph.ServicesProvider{ServicesProvider: proto.Clone(sp).(*sppb.ServicesProvider)}, nil
}

func (f *fakeSPCtrl) Update(_ context.Context, sp *sppb.ServicesProvider) error {
	f.db[sp.GetUuid()] = proto.Clone(sp).(*sppb.ServicesProvider)
	return nil
}

func (f *fakeSPCtrl) Create(_ context.Context, sp *graph.ServicesProvider) error {
	sp.Uuid = "new-sp"
	f.db[sp.Uuid] = proto.Clone(sp.ServicesProvider).(*sppb.ServicesProvider)
	return nil
}

type fakeNS struct{ graph.NamespacesController }

func (fakeNS) Get(ctx context.Context, id string) (graph.Namespace, error) {
	requestor, _ := ctx.Value(nocloud.NoCloudAccount).(string)
	return graph.Namespace{
		Namespace:    &namespaces.Namespace{Uuid: id, Access: &access.Access{Level: levels[requestor]}},
		DocumentMeta: driver.DocumentMeta{ID: driver.NewDocumentID(schema.NAMESPACES_COL, id)},
	}, nil
}

type fakeCA struct{ graph.CommonActionsController }

func (fakeCA) HasAccess(_ context.Context, account string, _ driver.DocumentID, level access.Level) bool {
	return levels[account] >= level
}

// fakeDriver records the providers core hands to the driver.
type fakeDriver struct {
	driverpb.DriverServiceClient
	tested  []*sppb.ServicesProvider
	invoked []*sppb.ServicesProvider
}

func (d *fakeDriver) TestServiceProviderConfig(_ context.Context, in *driverpb.TestServiceProviderConfigRequest, _ ...grpc.CallOption) (*sppb.TestResponse, error) {
	d.tested = append(d.tested, proto.Clone(in.GetServicesProvider()).(*sppb.ServicesProvider))
	return &sppb.TestResponse{Result: true}, nil
}

func (d *fakeDriver) SpInvoke(_ context.Context, in *driverpb.SpInvokeRequest, _ ...grpc.CallOption) (*sppb.InvokeResponse, error) {
	d.invoked = append(d.invoked, proto.Clone(in.GetServicesProvider()).(*sppb.ServicesProvider))
	return &sppb.InvokeResponse{Result: true}, nil
}

func (d *fakeDriver) lastTested(t *testing.T) *sppb.ServicesProvider {
	t.Helper()
	if len(d.tested) == 0 {
		t.Fatal("driver was not asked to test the config")
	}
	return d.tested[len(d.tested)-1]
}

var storedSecrets = map[string]any{
	"host":  "https://one.example:2633/RPC2",
	"user":  "oneadmin",
	"pass":  "hunter2",
	"group": float64(101),
	"vlans": map[string]any{"pool": map[string]any{"start": float64(10), "size": float64(100)}},
	"empty": "",
}

var redactedSecrets = map[string]any{
	"host": RedactedSecret, "user": RedactedSecret, "pass": RedactedSecret,
	"group": RedactedSecret, "vlans": RedactedSecret, "empty": "",
}

func newTestServer(t *testing.T) (*ServicesProviderServer, *fakeSPCtrl, *fakeDriver) {
	t.Helper()
	ctrl := &fakeSPCtrl{db: map[string]*sppb.ServicesProvider{
		"sp-1": {
			Uuid:    "sp-1",
			Type:    "fake",
			Title:   "One",
			Secrets: secretsOf(t, storedSecrets),
			Vars:    map[string]*sppb.Var{"cpu": {Value: map[string]*structpb.Value{"default": structpb.NewNumberValue(1)}}},
		},
	}}
	drv := &fakeDriver{}
	s := &ServicesProviderServer{
		log:               zap.NewNop(),
		ctrl:              ctrl,
		ns_ctrl:           fakeNS{},
		ca:                fakeCA{},
		drivers:           map[string]driverpb.DriverServiceClient{"fake": drv},
		extention_servers: map[string]sppb.ServicesProvidersExtentionsServiceClient{},
	}
	return s, ctrl, drv
}

func TestGetRedactsSecretsForUsers(t *testing.T) {
	s, _, _ := newTestServer(t)

	t.Run("root user gets redacted values", func(t *testing.T) {
		sp, err := s.Get(userCtx(t, "0"), &sppb.GetRequest{Uuid: "sp-1"})
		if err != nil {
			t.Fatal(err)
		}
		assertSecrets(t, sp.GetSecrets(), redactedSecrets)
		if sp.GetVars() == nil {
			t.Error("root should still get vars")
		}
	})

	t.Run("OAuth token for root gets redacted values", func(t *testing.T) {
		sp, err := s.Get(oauthCtx(t, "0"), &sppb.GetRequest{Uuid: "sp-1"})
		if err != nil {
			t.Fatal(err)
		}
		assertSecrets(t, sp.GetSecrets(), redactedSecrets)
	})

	t.Run("admin gets redacted values", func(t *testing.T) {
		sp, err := s.Get(userCtx(t, "admin"), &sppb.GetRequest{Uuid: "sp-1"})
		if err != nil {
			t.Fatal(err)
		}
		assertSecrets(t, sp.GetSecrets(), redactedSecrets)
	})

	t.Run("plain user gets neither secrets nor vars", func(t *testing.T) {
		sp, err := s.Get(userCtx(t, "user"), &sppb.GetRequest{Uuid: "sp-1"})
		if err != nil {
			t.Fatal(err)
		}
		if sp.GetSecrets() != nil || sp.GetVars() != nil {
			t.Fatalf("plain user got secrets %v / vars %v", sp.GetSecrets(), sp.GetVars())
		}
	})
}

func TestGetRevealsSecretsToRootOnly(t *testing.T) {
	s, _, _ := newTestServer(t)

	sp, err := s.Get(withReveal(userCtx(t, "0")), &sppb.GetRequest{Uuid: "sp-1"})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, sp.GetSecrets(), storedSecrets)

	for _, account := range []string{"admin", "user"} {
		_, err := s.Get(withReveal(userCtx(t, account)), &sppb.GetRequest{Uuid: "sp-1"})
		if status.Code(err) != codes.PermissionDenied {
			t.Errorf("%s asking for secrets: got %v, want PermissionDenied", account, err)
		}
	}
}

func TestGetKeepsSecretsForServiceCalls(t *testing.T) {
	// Drivers such as nocloud-ansible read providers through Get with a
	// service token and need the real values.
	s, _, _ := newTestServer(t)
	sp, err := s.Get(serviceCtx(t), &sppb.GetRequest{Uuid: "sp-1"})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, sp.GetSecrets(), storedSecrets)
}

func TestUpdateKeepsRedactedSecrets(t *testing.T) {
	s, ctrl, drv := newTestServer(t)
	ctx := userCtx(t, "0")

	// What the admin UI does: take the provider from Get and save it back
	// with another title.
	sp, err := s.Get(ctx, &sppb.GetRequest{Uuid: "sp-1"})
	if err != nil {
		t.Fatal(err)
	}
	sp.Title = "Renamed"

	res, err := s.Update(ctx, sp)
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, ctrl.db["sp-1"].GetSecrets(), storedSecrets)
	if ctrl.db["sp-1"].GetTitle() != "Renamed" {
		t.Errorf("title not saved: %q", ctrl.db["sp-1"].GetTitle())
	}
	assertSecrets(t, drv.lastTested(t).GetSecrets(), storedSecrets) // the driver tests the real config
	assertSecrets(t, res.GetSecrets(), redactedSecrets)             // and the response is redacted
}

func TestUpdateWithoutSecretsKeepsThem(t *testing.T) {
	s, ctrl, _ := newTestServer(t)
	_, err := s.Update(userCtx(t, "0"), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One"})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, ctrl.db["sp-1"].GetSecrets(), storedSecrets)

	_, err = s.Update(userCtx(t, "0"), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One", Secrets: map[string]*structpb.Value{}})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, ctrl.db["sp-1"].GetSecrets(), storedSecrets)
}

func TestUpdateChangesAndRemovesSecrets(t *testing.T) {
	s, ctrl, drv := newTestServer(t)
	_, err := s.Update(userCtx(t, "0"), &sppb.ServicesProvider{
		Uuid: "sp-1", Type: "fake", Title: "One",
		Secrets: secretsOf(t, map[string]any{
			"pass":  "new-pass",
			"vlans": RemoveSecret,
			"token": "abc",
			"user":  RedactedSecret,
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"host":  "https://one.example:2633/RPC2",
		"user":  "oneadmin",
		"pass":  "new-pass",
		"group": float64(101),
		"token": "abc",
		"empty": "",
	}
	assertSecrets(t, ctrl.db["sp-1"].GetSecrets(), want)
	assertSecrets(t, drv.lastTested(t).GetSecrets(), want)
}

func TestUpdateNeedsAdmin(t *testing.T) {
	s, ctrl, _ := newTestServer(t)
	_, err := s.Update(userCtx(t, "user"), &sppb.ServicesProvider{
		Uuid: "sp-1", Type: "fake", Title: "One",
		Secrets: secretsOf(t, map[string]any{"pass": "pwned"}),
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("got %v, want PermissionDenied", err)
	}
	assertSecrets(t, ctrl.db["sp-1"].GetSecrets(), storedSecrets)
}

func TestUpdateResponseForServiceCallIsUnchanged(t *testing.T) {
	s, _, _ := newTestServer(t)
	res, err := s.Update(serviceCtx(t), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One"})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, res.GetSecrets(), storedSecrets)
}

func TestTestMergesStoredSecretsForAdmins(t *testing.T) {
	s, _, drv := newTestServer(t)
	redacted := secretsOf(t, redactedSecrets)

	// The "test config" flow with a config taken from Get.
	for _, account := range []string{"0", "admin"} {
		res, err := s.Test(userCtx(t, account), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One", Secrets: redacted})
		if err != nil || !res.GetResult() {
			t.Fatalf("%s: Test = %v, %v", account, res, err)
		}
		assertSecrets(t, drv.lastTested(t).GetSecrets(), storedSecrets)
	}

	// A changed value is tested as sent.
	_, err := s.Test(userCtx(t, "0"), &sppb.ServicesProvider{
		Uuid: "sp-1", Type: "fake", Title: "One",
		Secrets: secretsOf(t, map[string]any{"pass": "try-this", "host": RedactedSecret}),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := drv.lastTested(t).GetSecrets()
	if got["pass"].GetStringValue() != "try-this" || got["host"].GetStringValue() != "https://one.example:2633/RPC2" {
		t.Fatalf("tested secrets = %v", got)
	}
}

func TestTestDoesNotHandStoredSecretsToNonAdmins(t *testing.T) {
	s, _, drv := newTestServer(t)
	_, err := s.Test(userCtx(t, "user"), &sppb.ServicesProvider{
		Uuid: "sp-1", Type: "fake", Title: "One",
		Secrets: secretsOf(t, map[string]any{"pass": RedactedSecret, "host": "https://attacker.example"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := drv.lastTested(t).GetSecrets()
	if got["pass"].GetStringValue() != RedactedSecret {
		t.Fatalf("a non-admin's test got the stored password: %v", got)
	}
}

func TestCreateRedactsResponseAndStoresValues(t *testing.T) {
	s, ctrl, drv := newTestServer(t)
	res, err := s.Create(userCtx(t, "0"), &sppb.ServicesProvider{
		Type: "fake", Title: "Two",
		Secrets: secretsOf(t, map[string]any{"user": "u", "pass": "p", "stale": RedactedSecret}),
	})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, res.GetSecrets(), map[string]any{"user": RedactedSecret, "pass": RedactedSecret})
	assertSecrets(t, ctrl.db["new-sp"].GetSecrets(), map[string]any{"user": "u", "pass": "p"})
	assertSecrets(t, drv.lastTested(t).GetSecrets(), map[string]any{"user": "u", "pass": "p"})
}

func TestInvokeHandsDriverTheFullProvider(t *testing.T) {
	// Driver calls (here SpInvoke; Up, Down, Monitoring and instance Invoke
	// read the provider from the DB the same way) keep the real secrets.
	s, _, drv := newTestServer(t)
	if _, err := s.Invoke(userCtx(t, "0"), &sppb.InvokeRequest{Uuid: "sp-1", Method: "get_plans"}); err != nil {
		t.Fatal(err)
	}
	if len(drv.invoked) != 1 {
		t.Fatalf("driver invoked %d times", len(drv.invoked))
	}
	assertSecrets(t, drv.invoked[0].GetSecrets(), storedSecrets)
}

func TestAdminCannotPointStoredSecretsElsewhere(t *testing.T) {
	// An admin below root cannot read secrets, so it must not be able to keep
	// the stored password while changing the host it is sent to.
	s, ctrl, drv := newTestServer(t)
	attack := secretsOf(t, map[string]any{
		"host": "https://attacker.example", "user": RedactedSecret, "pass": RedactedSecret,
	})

	_, err := s.Update(userCtx(t, "admin"), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One", Secrets: attack})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("Update: got %v, want PermissionDenied", err)
	}
	_, err = s.Test(userCtx(t, "admin"), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One", Secrets: attack})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("Test: got %v, want PermissionDenied", err)
	}
	// Leaving the other keys out does not keep them either: for an admin,
	// sent secrets replace the stored ones, as before.
	_, err = s.Test(userCtx(t, "admin"), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One",
		Secrets: secretsOf(t, map[string]any{"host": "https://attacker.example"})})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, drv.lastTested(t).GetSecrets(), map[string]any{"host": "https://attacker.example"})

	for _, tested := range drv.tested {
		if tested.GetSecrets()["pass"].GetStringValue() == "hunter2" {
			t.Fatalf("stored password handed to the driver: %v", tested.GetSecrets())
		}
	}
	assertSecrets(t, ctrl.db["sp-1"].GetSecrets(), storedSecrets)
}

func TestAdminKeepsOrReplacesAllSecrets(t *testing.T) {
	s, ctrl, _ := newTestServer(t)

	// keep: what the admin UI sends back untouched
	_, err := s.Update(userCtx(t, "admin"), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One", Secrets: secretsOf(t, redactedSecrets)})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, ctrl.db["sp-1"].GetSecrets(), storedSecrets)

	// replace all
	_, err = s.Update(userCtx(t, "admin"), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One",
		Secrets: secretsOf(t, map[string]any{"host": "https://two.example", "user": "u2", "pass": "p2"})})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, ctrl.db["sp-1"].GetSecrets(), map[string]any{"host": "https://two.example", "user": "u2", "pass": "p2"})
}

func TestServiceCallsKeepOldSecretsSemantics(t *testing.T) {
	// Drivers and core services never see redacted values: Update replaces
	// the secrets with what they send and Test tests what they send.
	s, ctrl, drv := newTestServer(t)

	_, err := s.Test(serviceCtx(t), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One",
		Secrets: secretsOf(t, map[string]any{"pass": "p"})})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, drv.lastTested(t).GetSecrets(), map[string]any{"pass": "p"})

	_, err = s.Update(serviceCtx(t), &sppb.ServicesProvider{Uuid: "sp-1", Type: "fake", Title: "One",
		Secrets: secretsOf(t, map[string]any{"pass": "p"})})
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, ctrl.db["sp-1"].GetSecrets(), map[string]any{"pass": "p"})
}
