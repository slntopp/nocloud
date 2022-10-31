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

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

type Access struct {
	From  driver.DocumentID `json:"_from"`
	To    driver.DocumentID `json:"_to"`
	Level int32             `json:"level"`
	Role  string            `json:"role"`

	driver.DocumentMeta
}

// account - Account Key, node - DocumentID
func HasAccess(ctx context.Context, db driver.Database, account string, node driver.DocumentID, level int32) bool {
	if (schema.ACCOUNTS_COL + "/" + account) == node.String() {
		return true
	}
	_, r := AccessLevel(ctx, db, account, node)
	return r >= level
}

// account - Account Key, node - DocumentID
func AccessLevel(ctx context.Context, db driver.Database, account string, node driver.DocumentID) (bool, int32) {
	query := `FOR path IN OUTBOUND K_SHORTEST_PATHS @account TO @node GRAPH @permissions RETURN path.edges[0].level`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"account":     schema.ACCOUNTS_COL + "/" + account,
		"node":        node,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
	})
	if err != nil {
		return false, 0
	}
	defer c.Close()

	var access int32 = 0
	for {
		var level int32
		_, err := c.ReadDocument(ctx, &level)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			continue
		}
		if level > access {
			access = level
		}
	}
	return access > 0, access
}
