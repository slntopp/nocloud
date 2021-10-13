package graph

import (
	"context"

	"github.com/arangodb/go-driver"
)

var (
	NAMESPACES_COL = "Namespaces"
)

type Namespace struct {
	Title string `json:"title"`
	driver.DocumentMeta
}

type NamespacesController struct {
	col driver.Collection
}

// func Create(title string) (string, error) {}

func (ctrl *NamespacesController) Get(ctx context.Context, id string) (Namespace, error) {
	var r Namespace
	_, err := ctrl.col.ReadDocument(nil, id, &r)
	return r, err
}

func (ctrl *NamespacesController) Create(ctx context.Context, title string) (Namespace, error) {
	acc := Namespace{
		Title: title,
	}
	meta, err := ctrl.col.CreateDocument(ctx, acc)
	acc.DocumentMeta = meta
	return acc, err
}