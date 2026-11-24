package services_providers

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func val(t *testing.T, v any) *structpb.Value {
	t.Helper()
	pv, err := structpb.NewValue(v)
	if err != nil {
		t.Fatalf("structpb.NewValue(%v): %v", v, err)
	}
	return pv
}

func secretsOf(t *testing.T, m map[string]any) map[string]*structpb.Value {
	t.Helper()
	res := make(map[string]*structpb.Value, len(m))
	for k, v := range m {
		res[k] = val(t, v)
	}
	return res
}

func assertSecrets(t *testing.T, got map[string]*structpb.Value, want map[string]any) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("secrets: got %d keys %v, want %d keys %v", len(got), got, len(want), want)
	}
	for k, w := range want {
		g, ok := got[k]
		if !ok {
			t.Fatalf("secrets: key %q missing in %v", k, got)
		}
		if !proto.Equal(g, val(t, w)) {
			t.Fatalf("secrets[%q] = %v, want %v", k, g.AsInterface(), w)
		}
	}
}

func TestRedactSecrets(t *testing.T) {
	in := secretsOf(t, map[string]any{
		"pass":   "hunter2",
		"group":  float64(101),
		"tls":    true,
		"vlans":  map[string]any{"pool": map[string]any{"start": float64(10), "size": float64(100)}},
		"hosts":  []any{"a", "b"},
		"empty":  "",
		"absent": nil,
	})
	before := proto.Clone(structpb.NewStructValue(&structpb.Struct{Fields: in}))

	got := RedactSecrets(in)

	assertSecrets(t, got, map[string]any{
		"pass":   RedactedSecret,
		"group":  RedactedSecret,
		"tls":    RedactedSecret,
		"vlans":  RedactedSecret,
		"hosts":  RedactedSecret,
		"empty":  "",
		"absent": nil,
	})
	if !proto.Equal(before, structpb.NewStructValue(&structpb.Struct{Fields: in})) {
		t.Fatal("RedactSecrets changed its input")
	}
	if RedactSecrets(nil) != nil {
		t.Fatal("RedactSecrets(nil) should be nil")
	}
}

func TestMergeSecrets(t *testing.T) {
	stored := func() map[string]*structpb.Value {
		return secretsOf(t, map[string]any{
			"host":  "https://one.example",
			"user":  "admin",
			"pass":  "hunter2",
			"vlans": map[string]any{"pool": map[string]any{"start": float64(10)}},
		})
	}

	t.Run("nothing sent keeps everything", func(t *testing.T) {
		for _, incoming := range []map[string]*structpb.Value{nil, {}} {
			assertSecrets(t, MergeSecrets(stored(), incoming), map[string]any{
				"host": "https://one.example", "user": "admin", "pass": "hunter2",
				"vlans": map[string]any{"pool": map[string]any{"start": float64(10)}},
			})
		}
	})

	t.Run("redacted values keep the stored ones", func(t *testing.T) {
		incoming := RedactSecrets(stored())
		assertSecrets(t, MergeSecrets(stored(), incoming), map[string]any{
			"host": "https://one.example", "user": "admin", "pass": "hunter2",
			"vlans": map[string]any{"pool": map[string]any{"start": float64(10)}},
		})
	})

	t.Run("change, remove, add and leave out", func(t *testing.T) {
		s := stored()
		incoming := secretsOf(t, map[string]any{
			"host":    RedactedSecret,
			"pass":    "new-pass",
			"vlans":   RemoveSecret,
			"token":   "tok",
			"unknown": RedactedSecret, // nothing stored to keep: dropped
		})
		got := MergeSecrets(s, incoming)
		assertSecrets(t, got, map[string]any{
			"host":  "https://one.example",
			"user":  "admin", // left out: unchanged
			"pass":  "new-pass",
			"token": "tok",
		})
		assertSecrets(t, s, map[string]any{
			"host": "https://one.example", "user": "admin", "pass": "hunter2",
			"vlans": map[string]any{"pool": map[string]any{"start": float64(10)}},
		})
	})

	t.Run("empty string sets an empty value", func(t *testing.T) {
		got := MergeSecrets(stored(), secretsOf(t, map[string]any{"pass": ""}))
		if got["pass"].GetStringValue() != "" || got["pass"] == nil {
			t.Fatalf("pass = %v, want empty string", got["pass"])
		}
	})
}

func TestApplySecrets(t *testing.T) {
	stored := secretsOf(t, map[string]any{"host": "h", "pass": "p"})

	got, err := applySecrets(true, stored, secretsOf(t, map[string]any{"host": "h2", "pass": RedactedSecret}))
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, got, map[string]any{"host": "h2", "pass": "p"})

	got, err = applySecrets(false, stored, secretsOf(t, map[string]any{"pass": RedactedSecret}))
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, got, map[string]any{"host": "h", "pass": "p"})

	got, err = applySecrets(false, stored, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, got, map[string]any{"host": "h", "pass": "p"})

	got, err = applySecrets(false, stored, secretsOf(t, map[string]any{"host": "h2"}))
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, got, map[string]any{"host": "h2"})

	// empty values come back from the redacted view as they are: unchanged
	withEmpty := secretsOf(t, map[string]any{"host": "h", "pass": "p", "ca": ""})
	got, err = applySecrets(false, withEmpty, secretsOf(t, map[string]any{"host": RedactedSecret, "pass": RedactedSecret, "ca": ""}))
	if err != nil {
		t.Fatal(err)
	}
	assertSecrets(t, got, map[string]any{"host": "h", "pass": "p", "ca": ""})

	for _, incoming := range []map[string]any{
		{"host": "h2", "pass": RedactedSecret},
		{"pass": RemoveSecret},
		// a correct guess of a stored value is still a change, or Update
		// would tell an admin whether the guess was right
		{"host": RedactedSecret, "pass": "p"},
	} {
		if _, err := applySecrets(false, stored, secretsOf(t, incoming)); status.Code(err) != codes.PermissionDenied {
			t.Errorf("non-root %v: got %v, want PermissionDenied", incoming, err)
		}
	}
}

func TestDropSecretSentinels(t *testing.T) {
	got := dropSecretSentinels(secretsOf(t, map[string]any{
		"pass": RedactedSecret, "user": RemoveSecret, "host": "h",
	}))
	assertSecrets(t, got, map[string]any{"host": "h"})
}

func signed(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

// serviceCtx authenticates like a driver or core service: a token from
// auth.MakeToken, run through the real auth middleware.
func serviceCtx(t *testing.T) context.Context {
	t.Helper()
	auth.SetContext(zap.NewNop(), nil, []byte("test-key"))
	token, err := auth.MakeToken("0")
	if err != nil {
		t.Fatal(err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "bearer "+token))
	ctx, err = auth.JWT_AUTH_MIDDLEWARE(ctx)
	if err != nil {
		t.Fatalf("auth middleware: %v", err)
	}
	return ctx
}

// userCtx is what the auth middleware makes of a login token (bound to a
// session): the admin UI, the customer app, API scripts.
func userCtx(t *testing.T, account string) context.Context {
	t.Helper()
	token := signed(t, jwt.MapClaims{
		nocloud.NOCLOUD_ACCOUNT_CLAIM: account,
		nocloud.NOCLOUD_SESSION_CLAIM: "sid",
		nocloud.NOCLOUD_ROOT_CLAIM:    4,
		"expires":                     4102444800,
	})
	ctx := context.WithValue(context.Background(), nocloud.NoCloudSession, "sid")
	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, account)
	return context.WithValue(ctx, nocloud.NoCloudToken, token)
}

func oauthCtx(t *testing.T, account string) context.Context {
	t.Helper()
	token := signed(t, jwt.MapClaims{
		nocloud.NOCLOUD_ACCOUNT_CLAIM:   account,
		nocloud.NOCLOUD_NOSESSION_CLAIM: true,
		"client_id":                     "some-app",
		"scopes":                        []string{"openid"},
	})
	ctx := context.WithValue(context.Background(), nocloud.NoCloudAccount, account)
	return context.WithValue(ctx, nocloud.NoCloudToken, token)
}

func withReveal(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	md = metadata.Join(md, metadata.Pairs(nocloud.NOCLOUD_REVEAL_SECRETS_METADATA, "true"))
	return metadata.NewIncomingContext(ctx, md)
}

func TestIsServiceCall(t *testing.T) {
	if !isServiceCall(serviceCtx(t)) {
		t.Error("a token from auth.MakeToken should count as a service call")
	}
	if isServiceCall(userCtx(t, "0")) {
		t.Error("a root login token (with a session) must not count as a service call")
	}
	if isServiceCall(oauthCtx(t, "0")) {
		t.Error("an OAuth access token must not count as a service call")
	}
	if isServiceCall(context.Background()) {
		t.Error("no token must not count as a service call")
	}
}

func TestRevealRequested(t *testing.T) {
	if revealRequested(userCtx(t, "0")) {
		t.Error("no metadata: no reveal")
	}
	if !revealRequested(withReveal(userCtx(t, "0"))) {
		t.Error("reveal metadata not seen")
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(nocloud.NOCLOUD_REVEAL_SECRETS_METADATA, "no"))
	if revealRequested(ctx) {
		t.Error(`"no" is not a reveal request`)
	}
}
