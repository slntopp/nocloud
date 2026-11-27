package services_providers

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// How ServicesProvider secrets travel through the API:
//
//   - Users (root included) get the keys of a provider's secrets with every
//     non-empty value replaced by RedactedSecret, from Get, Create and Update.
//     Below admin level they get no secrets at all.
//   - Root reads the real values with an explicit request: GET
//     /sp/{uuid}/secrets, or Get with the gRPC metadata
//     nocloud.NOCLOUD_REVEAL_SECRETS_METADATA set to "true".
//   - Drivers and other core services keep getting the full provider: core
//     passes it to them as before (Up, Down, Monitoring, Invoke, Test, ...),
//     and service tokens (see isServiceCall) still read real values from Get.
//
// Update (and Test, for a saved provider) apply the secrets a root user sends
// to the stored ones:
//   - no secrets (or an empty map): the stored secrets stay as they are;
//   - RedactedSecret as a value: that key keeps its stored value;
//   - RemoveSecret as a value: that key is deleted;
//   - any other value: that key is set to it;
//   - keys left out: unchanged.
//
// Admins below root cannot read secrets, so they may not combine kept values
// with new ones either (that would let them point stored credentials at a
// host of their choosing): they keep the secrets as they are (none sent, or
// only RedactedSecret values) or replace them all, as before.
//
// Service tokens (drivers, core services) keep the old behaviour: secrets
// they send in Update replace the stored ones.

// RedactedSecret stands in for a stored secret value in API responses. Sent
// back as a value, it keeps the stored one.
const RedactedSecret = "_nocloud.redacted"

// RemoveSecret, sent as a secret's value in Update, deletes that key.
const RemoveSecret = "_nocloud.remove"

// isEmptyValue reports whether v is null or an empty string: values that
// RedactSecrets passes through as they are.
func isEmptyValue(v *structpb.Value) bool {
	switch k := v.GetKind().(type) {
	case nil, *structpb.Value_NullValue:
		return true
	case *structpb.Value_StringValue:
		return k.StringValue == ""
	}
	return false
}

func isSentinel(v *structpb.Value, sentinel string) bool {
	s, ok := v.GetKind().(*structpb.Value_StringValue)
	return ok && s.StringValue == sentinel
}

// RedactSecrets returns the secrets with every value that is set replaced by
// RedactedSecret. Nulls and empty strings carry no secret and are kept, so
// clients can tell a key that is set from one that is not.
func RedactSecrets(secrets map[string]*structpb.Value) map[string]*structpb.Value {
	if secrets == nil {
		return nil
	}
	redacted := make(map[string]*structpb.Value, len(secrets))
	for key, value := range secrets {
		switch v := value.GetKind().(type) {
		case nil, *structpb.Value_NullValue:
			redacted[key] = structpb.NewNullValue()
		case *structpb.Value_StringValue:
			if v.StringValue == "" {
				redacted[key] = structpb.NewStringValue("")
				continue
			}
			redacted[key] = structpb.NewStringValue(RedactedSecret)
		default:
			redacted[key] = structpb.NewStringValue(RedactedSecret)
		}
	}
	return redacted
}

// MergeSecrets applies the secrets a client sent to the stored ones, as
// described at the top of this file, and returns the result. Neither map is
// changed.
func MergeSecrets(stored, incoming map[string]*structpb.Value) map[string]*structpb.Value {
	merged := make(map[string]*structpb.Value, len(stored)+len(incoming))
	for key, value := range stored {
		merged[key] = value
	}
	for key, value := range incoming {
		switch {
		case isSentinel(value, RedactedSecret):
			// keep the stored value (if there is none, there is nothing to keep)
		case isSentinel(value, RemoveSecret):
			delete(merged, key)
		default:
			merged[key] = value
		}
	}
	if stored == nil && len(merged) == 0 {
		return nil
	}
	return merged
}

// applySecrets returns the secrets a provider ends up with when a user sends
// incoming for it, as described at the top of this file. root tells whether
// the user has root access (and may read the secrets anyway).
func applySecrets(root bool, stored, incoming map[string]*structpb.Value) (map[string]*structpb.Value, error) {
	if root {
		return MergeSecrets(stored, incoming), nil
	}
	keepsAll := true
	hasSentinel := false
	for key, value := range incoming {
		switch {
		case isSentinel(value, RedactedSecret):
			hasSentinel = true
		case isSentinel(value, RemoveSecret):
			hasSentinel = true
			keepsAll = false
		case isEmptyValue(value) && proto.Equal(value, stored[key]):
			// an empty value the redacted view showed as it is, sent back
			// unchanged (only empty ones: comparing others would let an
			// admin guess stored values)
		default:
			keepsAll = false
		}
	}
	if keepsAll {
		return MergeSecrets(stored, nil), nil
	}
	if hasSentinel {
		return nil, status.Error(codes.PermissionDenied,
			"Only root can change some ServicesProvider secrets and keep the others: send all of them, or none")
	}
	return incoming, nil
}

// dropSecretSentinels removes keys whose value is RedactedSecret or
// RemoveSecret: on Create there is no stored value for them to refer to.
func dropSecretSentinels(secrets map[string]*structpb.Value) map[string]*structpb.Value {
	for key, value := range secrets {
		if isSentinel(value, RedactedSecret) || isSentinel(value, RemoveSecret) {
			delete(secrets, key)
		}
	}
	return secrets
}

// isServiceCall reports whether the request is made with a service token: one
// that core services and drivers mint with the signing key (auth.MakeToken).
// Such tokens carry the nosession claim; users' login tokens are bound to a
// session, and OAuth access tokens carry a client_id. Service callers keep
// getting real secret values, since drivers (nocloud-ansible, for one) read
// providers through Get.
func isServiceCall(ctx context.Context) bool {
	if _, ok := ctx.Value(nocloud.NoCloudSession).(string); ok {
		return false
	}
	token, _ := ctx.Value(nocloud.NoCloudToken).(string)
	if token == "" {
		return false
	}
	// The auth interceptor has verified this token before putting it into
	// the context; here it is only read.
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(token, claims); err != nil {
		return false
	}
	if nosession, _ := claims[nocloud.NOCLOUD_NOSESSION_CLAIM].(bool); !nosession {
		return false
	}
	if _, oauth := claims["client_id"]; oauth {
		return false
	}
	return true
}

// revealRequested reports whether the caller asked for the real secret values.
func revealRequested(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	for _, v := range md.Get(nocloud.NOCLOUD_REVEAL_SECRETS_METADATA) {
		if v == "true" || v == "1" {
			return true
		}
	}
	return false
}

func cloneSP(sp *sppb.ServicesProvider) *sppb.ServicesProvider {
	return proto.Clone(sp).(*sppb.ServicesProvider)
}

// forResponse returns the provider as an API response may carry it: as is to
// service callers, with redacted secrets to everyone else.
func forResponse(ctx context.Context, sp *sppb.ServicesProvider) *sppb.ServicesProvider {
	if sp == nil || isServiceCall(ctx) {
		return sp
	}
	res := cloneSP(sp)
	res.Secrets = RedactSecrets(res.Secrets)
	return res
}

// forLog returns a copy of the provider that is safe to write to logs.
func forLog(sp *sppb.ServicesProvider) *sppb.ServicesProvider {
	if sp == nil {
		return nil
	}
	res := cloneSP(sp)
	res.Secrets = RedactSecrets(res.Secrets)
	return res
}
