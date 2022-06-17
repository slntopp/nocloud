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
	"fmt"

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

func DeleteByDocID(ctx context.Context, db driver.Database, id driver.DocumentID) error {
	col, err := db.Collection(ctx, id.Collection())
	if err != nil {
		return fmt.Errorf("error while extracting collection: %v, DocID: %s", err, id)
	}

	_, err = col.RemoveDocument(ctx, id.Key())
	if err != nil {
		return fmt.Errorf("error while deleting by DocID: %v, DocID: %s", err, id)
	}

	return nil
}

func DeleteInBoundEdges(ctx context.Context, db driver.Database, id driver.DocumentID, graph string) error {
	query := `FOR node, edge IN INBOUND @node GRAPH @graph RETURN edge._id`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"node":  id,
		"graph": graph,
	})
	if err != nil {
		return err
	}
	defer c.Close()

	for {
		var id driver.DocumentID
		_, err := c.ReadDocument(ctx, &id)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}

		err = DeleteByDocID(ctx, db, id)
		if err != nil {
			return fmt.Errorf("error while deleting inbound edge: %v, id: %s", err, id)
		}
	}

	return nil
}

func DeleteRecursive(ctx context.Context, db driver.Database, id driver.DocumentID, graph string) error {
	query := `FOR node, edge IN 1..100 OUTBOUND @node GRAPH @graph FILTER edge.role == "owner" RETURN {edge: edge._id, to: edge._to}`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"node":  id,
		"graph": graph,
	})
	if err != nil {
		return err
	}
	defer c.Close()

	for {
		var obj map[string]driver.DocumentID
		_, err := c.ReadDocument(ctx, &obj)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}

		err = DeleteByDocID(ctx, db, obj["edge"])
		if err != nil {
			return fmt.Errorf("error while deleting 'edge': %v, obj: %v", err, obj)
		}
		if obj["to"].Collection() == schema.ACCOUNTS_COL {
			err = DeleteRecursive(ctx, db, id, schema.CREDENTIALS_GRAPH.Name)
			if err != nil {
				return fmt.Errorf("error while deleting 'to': %v, obj: %v", err, obj)
			}
		} else {
			err = DeleteInBoundEdges(ctx, db, obj["to"], schema.PERMISSIONS_GRAPH.Name)
			if err != nil {
				return fmt.Errorf("error while deleting 'to' inbound edges: %v, obj: %v", err, obj)
			}
			err = DeleteByDocID(ctx, db, obj["to"])
			if err != nil {
				return fmt.Errorf("error while deleting 'to': %v, obj: %v", err, obj)
			}
		}
	}

	err = DeleteInBoundEdges(ctx, db, id, schema.PERMISSIONS_GRAPH.Name)
	if err != nil {
		return fmt.Errorf("error while deleting node's inbound edges: %v, node: %v", err, id)
	}
	err = DeleteByDocID(ctx, db, id)
	if err != nil {
		return fmt.Errorf("error while deleting node: %v, DocID: %s", err, id)
	}

	return nil
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
