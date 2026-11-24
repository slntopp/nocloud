package main

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

const spSecretsPath = "/sp/{uuid}/secrets"

// registerSpSecretsRoute serves GET /sp/{uuid}/secrets: a provider's real
// secret values, which GET /sp/{uuid} returns redacted. It asks
// ServicesProvidersService/Get for them with the reveal metadata; the service
// checks that the requestor has root access. The response is
// {"uuid": ..., "secrets": {...}} and is not cached.
func registerSpSecretsRoute(mux *runtime.ServeMux, client sppb.ServicesProvidersServiceClient) error {
	return mux.HandlePath(http.MethodGet, spSecretsPath, spSecretsHandler(mux, client))
}

func spSecretsHandler(mux *runtime.ServeMux, client sppb.ServicesProvidersServiceClient) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		_, outbound := runtime.MarshalerForRequest(mux, r)
		w.Header().Set("Cache-Control", "no-store")

		ctx, err := runtime.AnnotateContext(r.Context(), mux, r,
			sppb.ServicesProvidersService_Get_FullMethodName,
			runtime.WithHTTPPathPattern(spSecretsPath))
		if err != nil {
			runtime.HTTPError(r.Context(), mux, outbound, w, r, err)
			return
		}
		ctx = metadata.AppendToOutgoingContext(ctx, nocloud.NOCLOUD_REVEAL_SECRETS_METADATA, "true")

		var md runtime.ServerMetadata
		sp, err := client.Get(ctx, &sppb.GetRequest{Uuid: pathParams["uuid"]},
			grpc.Header(&md.HeaderMD), grpc.Trailer(&md.TrailerMD))
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outbound, w, r, err)
			return
		}

		body := &structpb.Struct{Fields: map[string]*structpb.Value{
			"uuid":    structpb.NewStringValue(sp.GetUuid()),
			"secrets": structpb.NewStructValue(&structpb.Struct{Fields: sp.GetSecrets()}),
		}}
		runtime.ForwardResponseMessage(ctx, mux, outbound, w, r, body)
	}
}
