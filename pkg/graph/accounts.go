package graph

import (
	"github.com/arangodb/go-driver"
)

type Account struct {
	Title string `json:"title"`
	driver.DocumentMeta
}

// func Create(title string) (string, error) {}

func Get(col driver.Collection, id string) (Account, error) {
	var r Account
	_, err := col.ReadDocument(nil, id, &r)
	return r, err
}