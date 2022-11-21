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
	"github.com/slntopp/nocloud/pkg/access"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type Node struct {
	Collection string `json:"collection"`
	Key        string `json:"key"`
}

type Accessible interface {
	GetAccess() *access.Access
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
    LET ns = (FOR ns IN Namespaces FILTER ns._id == path.edges[0]._to RETURN ns)[0]
	RETURN MERGE(path.vertices[-1], {
		uuid: path.vertices[-1]._key,
	    access: {level: path.edges[0].level ? : 0, role: path.edges[0].role ? : "none", namespace: ns._key }
	})
`

func GetWithAccess[T Accessible](ctx context.Context, db driver.Database, id driver.DocumentID) (T, error) {
	var o T
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	vars := map[string]interface{}{
		"account":     driver.NewDocumentID(schema.ACCOUNTS_COL, requestor),
		"node":        id,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
	}
	c, err := db.Query(ctx, getWithAccessLevel, vars)
	if err != nil {
		return o, err
	}
	defer c.Close()

	_, err = c.ReadDocument(ctx, &o)
	if err != nil {
		return o, err
	}

	return o, nil
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

const edgeExistQuery = `
FOR edge IN @@collection
    FILTER edge._from == @fromDocID && edge._to == @toDocID
    LIMIT 1
    RETURN edge._key
`

func EdgeExist(ctx context.Context, db driver.Database, fromCollection, toCollection, fromKey, toKey string) (bool, error) {
	fromDocID := driver.NewDocumentID(fromCollection, fromKey)
	toDocID := driver.NewDocumentID(toCollection, toKey)
	collection := fromCollection + "2" + toCollection

	c, err := db.Query(ctx, edgeExistQuery, map[string]interface{}{
		"@collection": collection,
		"fromDocID":   fromDocID,
		"toDocID":     toDocID,
	})
	if err != nil {
		return false, err
	}
	defer c.Close()

	var key string
	_, err = c.ReadDocument(ctx, &key)
	if driver.IsNoMoreDocuments(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

const listObjectsOfKind = `
FOR node, edge, path IN 0..@depth OUTBOUND @from
GRAPH @permissions_graph
OPTIONS {order: "bfs", uniqueVertices: "global"}
FILTER IS_SAME_COLLECTION(@@kind, node)
// FILTER edge.level > 0 // TODO: ensure all edges have level
    LET perm = path.edges[0]
	RETURN MERGE(node, { uuid: node._key, access: { level: perm.level, role: perm.role, namespace: path.vertices[-2]._key } })
`

func ListWithAccess[T Accessible](
	ctx context.Context,
	log *zap.Logger,
	db driver.Database,
	fromDocument driver.DocumentID,
	collectionName string,
	depth int32,
) ([]T, error) {
	var list []T

	bindVars := map[string]interface{}{
		"depth":             depth,
		"from":              fromDocument,
		"permissions_graph": schema.PERMISSIONS_GRAPH.Name,
		"@kind":             collectionName,
	}

	log.Debug("ListWithAccess", zap.Any("vars", bindVars))
	c, err := db.Query(ctx, listObjectsOfKind, bindVars)
	if err != nil {
		return list, err
	}

	for c.HasMore() {
		var o T
		_, err := c.ReadDocument(ctx, &o)
		if err != nil {
			log.Warn("Could not append entity to query results", zap.Any("object", &o))
		}
		list = append(list, o)
	}

	return list, nil
}
