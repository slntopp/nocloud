package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakeSPClient struct {
	sppb.ServicesProvidersServiceClient
	md  metadata.MD
	req *sppb.GetRequest
	err error
}

func (f *fakeSPClient) Get(ctx context.Context, in *sppb.GetRequest, _ ...grpc.CallOption) (*sppb.ServicesProvider, error) {
	f.md, _ = metadata.FromOutgoingContext(ctx)
	f.req = in
	if f.err != nil {
		return nil, f.err
	}
	return &sppb.ServicesProvider{
		Uuid:    in.GetUuid(),
		Title:   "One",
		Type:    "ione",
		Secrets: map[string]*structpb.Value{"pass": structpb.NewStringValue("hunter2")},
		Vars:    map[string]*sppb.Var{"cpu": {}},
	}, nil
}

func newTestMux(t *testing.T, client sppb.ServicesProvidersServiceClient) *runtime.ServeMux {
	t.Helper()
	mux := runtime.NewServeMux()
	// the generated routes as in main(), plus the secrets route
	if err := sppb.RegisterServicesProvidersServiceHandlerClient(context.Background(), mux, client); err != nil {
		t.Fatal(err)
	}
	if err := registerSpSecretsRoute(mux, client); err != nil {
		t.Fatal(err)
	}
	return mux
}

func TestSpSecretsRoute(t *testing.T) {
	client := &fakeSPClient{}
	mux := newTestMux(t, client)

	req := httptest.NewRequest(http.MethodGet, "/sp/sp-1/secrets", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if got := client.req.GetUuid(); got != "sp-1" {
		t.Errorf("Get called for %q", got)
	}
	if got := client.md.Get(nocloud.NOCLOUD_REVEAL_SECRETS_METADATA); len(got) != 1 || got[0] != "true" {
		t.Errorf("reveal metadata = %v", got)
	}
	if got := client.md.Get("authorization"); len(got) != 1 || got[0] != "Bearer user-token" {
		t.Errorf("authorization not forwarded: %v", got)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q", cc)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 2 || body["uuid"] != "sp-1" {
		t.Fatalf("body should hold only uuid and secrets: %v", body)
	}
	if secrets, _ := body["secrets"].(map[string]any); secrets["pass"] != "hunter2" {
		t.Fatalf("secrets = %v", body["secrets"])
	}
}

func TestSpSecretsRouteDenied(t *testing.T) {
	client := &fakeSPClient{err: status.Error(codes.PermissionDenied, "Not enough access rights to read ServicesProvider secrets")}
	mux := newTestMux(t, client)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/sp/sp-1/secrets", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d, want 403: %s", rec.Code, rec.Body)
	}
}

func TestSpGetRouteDoesNotAskForSecrets(t *testing.T) {
	client := &fakeSPClient{}
	mux := newTestMux(t, client)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/sp/sp-1", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if got := client.md.Get(nocloud.NOCLOUD_REVEAL_SECRETS_METADATA); len(got) != 0 {
		t.Fatalf("GET /sp/{uuid} asked for secrets: %v", got)
	}
}
