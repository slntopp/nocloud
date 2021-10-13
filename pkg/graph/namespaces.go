package graph

import (
	"github.com/arangodb/go-driver"
)

type Namespace struct {
	Title string `json:"title"`
	driver.DocumentMeta
}

// func Create(title string) (string, error) {}

func Get(col driver.Collection, id string) (Namespace, error) {
	var r Namespace
	_, err := col.ReadDocument(nil, id, &r)
	return r, err
}