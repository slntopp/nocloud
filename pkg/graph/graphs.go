/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package graph

import (
	"context"
	"errors"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

type Node struct {
	Collection string `json:"collection"`
	Key        string `json:"key"`
}

type Deletable interface {
	Delete(context.Context, driver.Database) error
}

func DeleteNodeChildren(ctx context.Context, db driver.Database, node string) error {
	query := `FOR node, edge, path IN OUTBOUND @node GRAPH Permissions FILTER edge.role == "owner" RETURN PARSE_IDENTIFIER(node._id)`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"node": node,
	})
	if err != nil {
		return errors.New("Error executing find children query")
	}
	defer c.Close()

	for {
		var node Node
		_, err := c.ReadDocument(ctx, &node)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return errors.New("Error reading Document")
		}

		// fmt.Println("Node must be deleted", node)
		next, err := MakeDeletable(ctx, db, node)
		if err != nil {
			return errors.New("Error making node deletable")
		}

		err = next.Delete(ctx, db)
		if err != nil {
			return errors.New("Error deleting child node")
		}
	}

	return nil
}

func MakeDeletable(ctx context.Context, db driver.Database, node Node) (Deletable, error) {
	var result Deletable
	var err error
	col, _ := db.Collection(ctx, node.Collection)

	switch node.Collection {
	case schema.ACCOUNTS_COL:
		var acc Account
		_, err = col.ReadDocument(ctx, node.Key, &acc)
		result = &acc
	case schema.NAMESPACES_COL:
		var ns Namespace
		_, err = col.ReadDocument(ctx, node.Key, &ns)
		result = &ns
	default:
		return nil, errors.New("Error making node deletable: Can't define node type")
	}

	if err != nil {
		return nil, errors.New("Error making node deletable: Can't get node from Collection")
	}
	return result, nil
}

const getWithAccessLevel = `
FOR path IN OUTBOUND K_SHORTEST_PATHS @account TO @node
GRAPH @permissions SORT path.edges[0].level
	RETURN MERGE(path.vertices[-1], {
	    access_level: path.edges[0].level ? : 0, uuid: path.vertices[-1]._key
	})
`

func GetWithAccess(ctx context.Context, db driver.Database, acc, id driver.DocumentID, node interface{}) error {
	vars := map[string]interface{}{
		"account":     acc,
		"node":        id,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
	}
	c, err := db.Query(ctx, getWithAccessLevel, vars)
	if err != nil {
		return err
	}
	defer c.Close()

	_, err = c.ReadDocument(ctx, node)
	if err != nil {
		return err
	}

	return nil
}

const deleteEdgeQuery = `
FOR edge IN @@collection
    FILTER edge._from == @fromDocID && edge._to == @toDocID
    REMOVE edge._key IN @@collection
`

func DeleteEdge(ctx context.Context, db driver.Database, fromCollection, toCollection, fromKey, toKey string) error {
	fromDocID := driver.NewDocumentID(fromCollection, fromKey)
	toDocID := driver.NewDocumentID(toCollection, toKey)
	collection := fromCollection + "2" + toCollection

	c, err := db.Query(ctx, deleteEdgeQuery, map[string]interface{}{
		"@collection": collection,
		"fromDocID":   fromDocID,
		"toDocID":     toDocID,
	})
	if err != nil {
		return err
	}
	defer c.Close()

	return nil
}
