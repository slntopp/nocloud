package graph

import "github.com/arangodb/go-driver"

type Access struct {
	From driver.DocumentID `json:"_from"`
	To driver.DocumentID `json:"_to"`
	Level int8 `json:"level"`

	driver.DocumentMeta
}