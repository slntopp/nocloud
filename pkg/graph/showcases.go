package graph

import (
	"context"
	"github.com/arangodb/go-driver"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"go.uber.org/zap"
)

type ShowcasesController struct {
	log *zap.Logger
	col driver.Collection

	db driver.Database
}

func (ctrl *ShowcasesController) Create(ctx context.Context, showcase *sppb.Showcase) (*sppb.Showcase, error) {
	return nil, nil
}

func (ctrl *ShowcasesController) Update(ctx context.Context, showcase *sppb.Showcase) (*sppb.Showcase, error) {
	return nil, nil
}

func (ctrl *ShowcasesController) List(ctx context.Context) ([]*sppb.Showcase, error) {
	return nil, nil
}

func (ctrl *ShowcasesController) Get(ctx context.Context, uuid string) (*sppb.Showcase, error) {
	return nil, nil
}

func (ctrl *ShowcasesController) Delete(ctx context.Context, uuid string) error {
	return nil
}
