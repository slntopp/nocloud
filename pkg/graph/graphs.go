/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

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
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
	"math"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type ContextKey string

const (
	AQLTransactionContextKey ContextKey = "aql-transaction"
)

type TransactionsAPI interface {
	BeginTransaction(ctx context.Context) (context.Context, error)
	AbortTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
}

type Node struct {
	Collection string `json:"collection"`
	Key        string `json:"key"`
}

type Accessible interface {
	GetAccess() *access.Access
}

func Round(amount float64, precision int32, mode pb.Rounding) float64 {
	if mode == pb.Rounding_ROUND_HALF {
		return math.Round(amount*math.Pow10(int(precision))) / math.Pow10(int(precision))
	} else if mode == pb.Rounding_ROUND_UP {
		return math.Ceil(amount*math.Pow10(int(precision))) / math.Pow10(int(precision))
	} else {
		return math.Floor(amount*math.Pow10(int(precision))) / math.Pow10(int(precision))
	}
}

func deleteByDocID(ctx context.Context, db driver.Database, id driver.DocumentID) error {
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

var getEdge = `
LET inboundNode = DOCUMENT(@inboundNode)

LET inboundNode_edge = (
	FOR s, edge IN 1 INBOUND inboundNode
	GRAPH @permissions
	FILTER IS_SAME_COLLECTION(@collection, s)
	RETURN edge
)[0]

return inboundNode_edge._key
`

const listOwnedQuery = `
FOR node, edge IN 0..100
OUTBOUND @from
GRAPH Permissions
FILTER !edge || edge.role == "owner"
    RETURN MERGE({ node: node._id }, edge ? { edge: edge._id, parent: edge._from } : { edge: null, parent: null })
`

func listOwnedDeep(ctx context.Context, db driver.Database, id driver.DocumentID) (res *access.Nodes, err error) {
	c, err := db.Query(ctx, listOwnedQuery, map[string]interface{}{
		"from": id,
	})
	if err != nil {
		return res, err
	}
	defer c.Close()

	var nodes []*access.Node
	for {
		var node access.Node
		_, err := c.ReadDocument(ctx, &node)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return res, err
		}
		nodes = append(nodes, &node)
	}

	return &access.Nodes{Nodes: nodes}, nil
}

func deleteRecursive(ctx context.Context, db driver.Database, id driver.DocumentID) error {
	nodes, err := listOwnedDeep(ctx, db, id)
	if err != nil {
		return err
	}

	cols := make(map[string]driver.Collection)
	for i := len(nodes.Nodes) - 1; i >= 0; i-- {
		node := nodes.Nodes[i]

		if node.Node != "" {
			err := handleDeleteNodeInRecursion(ctx, db, node.Node, cols)
			if err != nil {
				if err.Error() == "ERR_ROOT_OBJECT_CANNOT_BE_DELETED" {
					continue
				}
				return err
			}
		}

		if node.Edge != "" {
			err := handleDeleteNodeInRecursion(ctx, db, node.Edge, cols)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func handleDeleteNodeInRecursion(ctx context.Context, db driver.Database, node string, cols map[string]driver.Collection) (err error) {
	id := strings.SplitN(node, "/", 2)
	col, ok := cols[id[0]]
	if !ok {
		col, err = db.Collection(ctx, id[0])
		if err != nil {
			return err
		}
		cols[id[0]] = col
	}

	if id[0] == schema.ACCOUNTS_COL {
		if id[1] == schema.ROOT_ACCOUNT_KEY {
			return errors.New("ERR_ROOT_OBJECT_CANNOT_BE_DELETED")
		}
		nodes, err := listCredentialsAndEdges(ctx, col.Database(), driver.DocumentID(node))
		if err != nil {
			return err
		}
		for _, node := range nodes {
			err = handleDeleteNodeInRecursion(ctx, col.Database(), node, cols)
			if err != nil {
				return err
			}
		}
	}
	if id[0] == schema.NAMESPACES_COL && id[1] == schema.ROOT_NAMESPACE_KEY {
		return errors.New("ERR_ROOT_OBJECT_CANNOT_BE_DELETED")
	}

	_, err = col.RemoveDocument(ctx, id[1])
	if e, ok := driver.AsArangoError(err); ok && e.Code == 404 {
		return nil
	}
	return err
}

func listCredentialsAndEdges(ctx context.Context, db driver.Database, account driver.DocumentID) (nodes []string, err error) {
	c, err := db.Query(ctx, listCredentialsAndEdgesQuery, map[string]interface{}{
		"account":     account,
		"credentials": schema.CREDENTIALS_COL,
	})
	if err != nil {
		return nil, err
	}
	defer c.Close()

	_, err = c.ReadDocument(ctx, &nodes)
	return nodes, err
}

const listCredentialsAndEdgesQuery = `
RETURN FLATTEN(
FOR node, edge IN 1 OUTBOUND @account
GRAPH @credentials
    RETURN [ node._id, edge._id ]
)
`

const getWithAccessLevel = `
FOR path IN OUTBOUND K_SHORTEST_PATHS @account TO @node
GRAPH @permissions SORT path.edges[0].level
    LET cred = LAST(
               FOR n, e IN 0..1 OUTBOUND @node GRAPH @credentials_graph 
               OPTIONS {order: "bfs", uniqueVertices: "global"}
                   RETURN n
        )
    RETURN MERGE(path.vertices[-1], {
        uuid: path.vertices[-1]._key,
        currency: DOCUMENT(@@currencies, TO_STRING(TO_NUMBER(path.vertices[-1].currency.id))),
	    access: {level: path.edges[0].level ? : 0, role: path.edges[0].role ? : "none", namespace: path.vertices[-2]._key, username: cred.username }
	})
`

func getWithAccess[T Accessible](ctx context.Context, db driver.Database, from driver.DocumentID, id driver.DocumentID) (T, error) {
	var o T

	vars := map[string]interface{}{
		"account":           from,
		"node":              id,
		"permissions":       schema.PERMISSIONS_GRAPH.Name,
		"credentials_graph": schema.CREDENTIALS_GRAPH.Name,
		"@currencies":       schema.CUR_COL,
	}
	c, err := db.Query(ctx, getWithAccessLevel, vars)
	if err != nil {
		return o, err
	}
	defer c.Close()

	meta, err := c.ReadDocument(ctx, &o)
	if err != nil {
		return o, err
	}

	if from.String() == meta.ID.String() {
		o.GetAccess().Level = access.Level_ROOT
	}

	return o, nil
}

func GetWithAccess[T Accessible](ctx context.Context, db driver.Database, from driver.DocumentID, id driver.DocumentID) (T, error) {
	return getWithAccess[T](ctx, db, from, id)
}

const deleteEdgeQuery = `
FOR edge IN @@collection
    FILTER edge._from == @fromDocID && edge._to == @toDocID
    REMOVE edge._key IN @@collection
`

func deleteEdge(ctx context.Context, db driver.Database, fromCollection, toCollection, fromKey, toKey string) error {
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

func edgeExist(ctx context.Context, db driver.Database, fromCollection, toCollection, fromKey, toKey string) (bool, error) {
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
    LET currency = DOCUMENT(@@currencies, TO_STRING(TO_NUMBER(node.currency.id)))
	RETURN MERGE(node, { uuid: node._key, currency: currency, access: { level: perm.level, role: perm.role, namespace: path.vertices[-2]._key } })
`

func listWithAccess[T Accessible](
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
		"@currencies":       schema.CUR_COL,
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

const listObjectsWithFiltersOfKind = `
LET list = (FOR node, edge, path IN 0..@depth OUTBOUND @from
	GRAPH @permissions_graph
	OPTIONS {order: "bfs", uniqueVertices: "global"}
	FILTER IS_SAME_COLLECTION(@@kind, node)
		LET perm = path.edges[0]
		%s
		RETURN MERGE(node, { uuid: node._key, access: { level: perm.level, role: perm.role, namespace: path.vertices[-2]._key } })
	)
	
	RETURN { 
		result: (@limit > 0) ? SLICE(list, @offset, @limit) : list,
		count: LENGTH(list)
	}
`

const listAccountsQuery = `
LET list = (FOR node, edge, path IN 0..@depth OUTBOUND @from
	GRAPH @permissions_graph
	OPTIONS {order: "bfs", uniqueVertices: "global"}
	FILTER IS_SAME_COLLECTION(@@kind, node)
		LET perm = path.edges[0]
		%s
		LET instances = (
			FOR subnode IN 0..@depth OUTBOUND node._id
				GRAPH @permissions_graph
				OPTIONS {order: "bfs", uniqueVertices: "global"}
				FILTER IS_SAME_COLLECTION(@@subkiund, subnode)
				RETURN subnode
			)

        LET cred = LAST(
               FOR n, e IN 0..1 OUTBOUND node GRAPH @credentials_graph 
               OPTIONS {order: "bfs", uniqueVertices: "global"}
                   RETURN n
        )
        
		RETURN MERGE(node, { uuid: node._key, currency: DOCUMENT(@@currencies, TO_STRING(TO_NUMBER(node.currency.id))), active: length(instances) != 0, access: { level: perm.level, role: perm.role, namespace: path.vertices[-2]._key, username: cred.username } })
	)
	
	LET active = LENGTH(
		FOR l in list
			FILTER l.active == true
			return l
	)
	
	RETURN { 
		result: (@limit > 0) ? SLICE(list, @offset, @limit) : list,
		count: LENGTH(list),
		active: active
	}
`

type ListQueryResult[T Accessible] struct {
	Result []T `json:"result"`
	Count  int `json:"count"`
	Active int `json:"active"`
}

// ListAccounts deprecated for import
func listAccounts[T Accessible](
	ctx context.Context,
	log *zap.Logger,
	db driver.Database,
	fromDocument driver.DocumentID,
	collectionName string,
	depth int32,
	offset, limit uint64,
	field, sort string,
	filters map[string]*structpb.Value,
) (*ListQueryResult[T], error) {
	var result ListQueryResult[T]

	bindVars := map[string]interface{}{
		"depth":             depth,
		"from":              fromDocument,
		"permissions_graph": schema.PERMISSIONS_GRAPH.Name,
		"credentials_graph": schema.CREDENTIALS_GRAPH.Name,
		"@kind":             collectionName,
		"@subkiund":         schema.INSTANCES_COL,
		"offset":            offset,
		"limit":             limit,
		"@currencies":       schema.CUR_COL,
	}

	var insert string

	if field != "" && sort != "" {
		insert += fmt.Sprintf("SORT node.%s %s\n", field, sort)
	}

	for key, val := range filters {
		if key == "search_param" && collectionName == schema.ACCOUNTS_COL {
			insert += fmt.Sprintf(` FILTER LOWER(node.title) LIKE LOWER("%s") || LOWER(node.data.email) LIKE LOWER("%s") || node._key LIKE "%s" || LOWER(node.data.company) LIKE LOWER("%s")`, "%"+val.GetStringValue()+"%", "%"+val.GetStringValue()+"%", "%"+val.GetStringValue()+"%", "%"+val.GetStringValue()+"%")
		} else if key == "namespace" && collectionName == schema.ACCOUNTS_COL {
			insert += fmt.Sprintf(` FILTER path.vertices[-2]._key == "%s"`, "%"+val.GetStringValue()+"%")
		} else if strings.HasPrefix(key, "data") {
			split := strings.Split(key, ".")
			if len(split) != 2 {
				continue
			}
			if split[1] == "address" || split[1] == "country" || split[1] == "email" || split[1] == "city" || split[1] == "company" || split[1] == "phone" {
				insert += fmt.Sprintf(` FILTER node.data["%s"] LIKE "%s"`, split[1], "%"+val.GetStringValue()+"%")
			} else if split[1] == "date_create" {
				values := val.GetStructValue().AsMap()
				if val, ok := values["from"]; ok {
					from := val.(float64)
					insert += fmt.Sprintf(` FILTER node.data["%s"] >= %f`, split[1], from)
				}

				if val, ok := values["to"]; ok {
					to := val.(float64)
					insert += fmt.Sprintf(` FILTER node.data["%s"] <= %f`, split[1], to)
				}
			} else if split[1] == "whmcs_id" {
				insert += fmt.Sprintf(` FILTER node.data["%s"] == %d`, split[1], int(val.GetNumberValue()))
			}
		} else if key == "uuid" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			insert += ` FILTER node._key in @uuids`
			bindVars["uuids"] = values
		} else if key == "access.level" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			insert += ` FILTER perm.level in @level`
			bindVars["level"] = values
		} else if key == "status" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			insert += ` FILTER TO_NUMBER(node.status) in @statuses`
			bindVars["statuses"] = values
		} else if key == "currency" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			insert += ` FILTER TO_NUMBER(node.currency.id) in @currencies`
			bindVars["currencies"] = values
		} else if key == "balance" {
			values := val.GetStructValue().AsMap()
			if val, ok := values["from"]; ok {
				from := val.(float64)
				insert += fmt.Sprintf(` FILTER node["%s"] >= %f`, key, from)
			}

			if val, ok := values["to"]; ok {
				to := val.(float64)
				insert += fmt.Sprintf(` FILTER node["%s"] <= %f`, key, to)
			}
		} else {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			insert += fmt.Sprintf(` FILTER node["%s"] in @%s`, key, key)
			bindVars[key] = values
		}
	}

	log.Debug("ListWithAccess", zap.Any("vars", bindVars))
	q := fmt.Sprintf(listAccountsQuery, insert)
	log.Debug("Query", zap.String("q", q))
	c, err := db.Query(ctx, q, bindVars)
	if err != nil {
		return nil, err
	}

	_, err = c.ReadDocument(ctx, &result)
	if err != nil {
		log.Warn("Could not append entity to query results", zap.Error(err))
	}

	return &result, nil
}

func listNamespaces[T Accessible](
	ctx context.Context,
	log *zap.Logger,
	db driver.Database,
	fromDocument driver.DocumentID,
	collectionName string,
	depth int32,
	offset, limit uint64,
	field, sort string,
	filters map[string]*structpb.Value,
) (*ListQueryResult[T], error) {
	var result ListQueryResult[T]

	bindVars := map[string]interface{}{
		"depth":             depth,
		"from":              fromDocument,
		"permissions_graph": schema.PERMISSIONS_GRAPH.Name,
		"@kind":             collectionName,
		"offset":            offset,
		"limit":             limit,
	}

	var insert string

	if field != "" && sort != "" {
		insert += fmt.Sprintf("SORT node.%s %s\n", field, sort)
	}

	for key, val := range filters {
		if key == "search_param" {
			insert += fmt.Sprintf(` FILTER LOWER(node.title) LIKE LOWER("%s")`, "%"+val.GetStringValue()+"%")
		} else if key == "access.level" {
			values := val.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			insert += ` FILTER perm.level in @levels`
			bindVars["levels"] = values
		} else if key == "account" {
			account := val.GetStringValue()
			insert += ` FILTER path.vertices[-2]._key == @account`
			bindVars["account"] = account
		}
	}

	log.Debug("ListWithAccess", zap.Any("vars", bindVars))
	q := fmt.Sprintf(listObjectsWithFiltersOfKind, insert)
	log.Debug("Query", zap.String("q", q))
	c, err := db.Query(ctx, q, bindVars)
	if err != nil {
		return nil, err
	}

	_, err = c.ReadDocument(ctx, &result)
	if err != nil {
		log.Warn("Could not append entity to query results", zap.Error(err))
	}

	return &result, nil
}
