package consent

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/consent"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type ConsentServer struct {
	pb.UnimplementedConsentServiceServer

	log *zap.Logger

	ctrl graph.ConsentsController
	ca   graph.CommonActionsController
}

func NewConsentServer(log *zap.Logger, db driver.Database) *ConsentServer {
	return &ConsentServer{
		log:  log.Named("ConsentServer"),
		ctrl: graph.NewConsentsController(log, db),
		ca:   graph.NewCommonActionsController(log, db),
	}
}

func (s *ConsentServer) Record(ctx context.Context, req *pb.RecordRequest) (*pb.RecordResponse, error) {
	log := s.log.Named("Record")

	if req.GetDecision() == pb.Decision_DECISION_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "decision must be ACCEPT or DECLINE")
	}

	record := &pb.Record{
		Decision:      req.GetDecision(),
		Categories:    req.GetCategories(),
		PolicyVersion: req.GetPolicyVersion(),
		Ip:            clientIP(ctx),
		UserAgent:     userAgent(ctx),
		CreatedAt:     time.Now().Unix(),
		Domain:        requestDomain(ctx),
	}
	if acc, ok := ctx.Value(nocloud.NoCloudAccount).(string); ok {
		record.Account = &acc
	}

	log.Debug("Recording consent", zap.Any("record", record))

	record, err := s.ctrl.Create(ctx, record)
	if err != nil {
		log.Error("Failed to store consent record", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to store consent record")
	}

	return &pb.RecordResponse{Uuid: record.GetUuid()}, nil
}

func (s *ConsentServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	log := s.log.Named("List")

	requester, ok := ctx.Value(nocloud.NoCloudAccount).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no requester")
	}

	root := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	if hasAccess := s.ca.HasAccess(ctx, requester, root, access.Level_ROOT); !hasAccess {
		return nil, status.Error(codes.PermissionDenied, "not enough access rights to perform List")
	}

	records, err := s.ctrl.List(ctx, req)
	if err != nil {
		log.Error("Failed to list consent records", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list consent records")
	}

	return &pb.ListResponse{Pool: records}, nil
}

func clientIP(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// Set by apiserver_web's HTTP-layer middleware for POST /consent, from the
	// original request, before it's re-forwarded as an internal gRPC call — takes
	// priority since the plain x-forwarded-for below reflects the internal hop too.
	if ip := md.Get("x-client-ip"); len(ip) > 0 && ip[0] != "" {
		return ip[0]
	}

	if fwd := md.Get("x-forwarded-for"); len(fwd) > 0 {
		// grpc-gateway joins the chain as "<original client>, <next hop>, ...";
		// the leftmost entry is the one closest to the actual visitor.
		if ip := strings.TrimSpace(strings.Split(fwd[0], ",")[0]); ip != "" {
			return ip
		}
	}

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		return p.Addr.String()
	}
	return ""
}

// grpc-gateway forwards standard HTTP headers into gRPC metadata prefixed with
// "grpcgateway-" (see runtime.DefaultHeaderMatcher), not under their raw HTTP names.
const gatewayHeaderPrefix = "grpcgateway-"

func userAgent(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ua := md.Get(gatewayHeaderPrefix + "user-agent"); len(ua) > 0 {
			return ua[0]
		}
	}
	return ""
}

// requestDomain identifies which frontend site sent the request, so that a single
// backend deployment can serve consent logs for multiple domains without mixing them up.
func requestDomain(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if origin := md.Get(gatewayHeaderPrefix + "origin"); len(origin) > 0 {
		if u, err := url.Parse(origin[0]); err == nil && u.Host != "" {
			return u.Host
		}
	}
	if referer := md.Get(gatewayHeaderPrefix + "referer"); len(referer) > 0 {
		if u, err := url.Parse(referer[0]); err == nil && u.Host != "" {
			return u.Host
		}
	}
	// x-forwarded-host is always populated by grpc-gateway (falls back to the
	// request's Host header), unlike the "grpcgateway-" prefixed headers above.
	if host := md.Get("x-forwarded-host"); len(host) > 0 {
		return strings.ToLower(host[0])
	}
	return ""
}
