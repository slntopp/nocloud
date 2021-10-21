package accounting

import (
	"context"

	"github.com/slntopp/nocloud/pkg/nocloud"

	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ValidateMetadata(ctx context.Context, log *zap.Logger) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Error("Failed to get metadata from context")
		return ctx, status.Error(codes.Aborted, "Failed to get metadata from context")
	}

	//Check for Authentication
	requestor := md.Get(nocloud.NOCLOUD_ACCOUNT_CLAIM)
	if requestor == nil {
		log.Error("Failed to authenticate account")
		return ctx, status.Error(codes.Unauthenticated, "Not authenticated")
	}
	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, requestor[0])

	return ctx, nil
}
