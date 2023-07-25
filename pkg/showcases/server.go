package showcases

import (
	"context"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
)

type ShowcasesServer struct {
	sppb.UnimplementedShowcasesServiceServer
}

func (s *ShowcasesServer) Create(ctx context.Context, req *sppb.Showcase) (*sppb.Showcase, error) {
	return nil, nil
}

func (s *ShowcasesServer) Update(ctx context.Context, req *sppb.Showcase) (*sppb.Showcase, error) {
	return nil, nil
}

func (s *ShowcasesServer) Get(ctx context.Context, req *sppb.GetRequest) (*sppb.Showcase, error) {
	return nil, nil
}

func (s *ShowcasesServer) List(ctx context.Context, req *sppb.ListRequest) (*sppb.Showcases, error) {
	return nil, nil
}

func (s *ShowcasesServer) Delete(ctx context.Context, req *sppb.DeleteRequest) (*sppb.DeleteResponse, error) {
	return nil, nil
}
