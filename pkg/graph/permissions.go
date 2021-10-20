package graph

import (
	"context"

	"github.com/arangodb/go-driver"
)

type Access struct {
	From driver.DocumentID `json:"_from"`
	To driver.DocumentID `json:"_to"`
	Level int32 `json:"level"`

	driver.DocumentMeta
}

// account - Account Key, node - DocumentID
func HasAccess(ctx context.Context, db driver.Database, account string, node string, level int32) (bool) {
	if (ACCOUNTS_COL + "/" + account) == node {
		return true
	}
	_, r := AccessLevel(ctx, db, account, node)
	return r >= level
}

// account - Account Key, node - DocumentID
func AccessLevel(ctx context.Context, db driver.Database, account string, node string) (bool, int32) {
	query := `FOR path IN OUTBOUND K_SHORTEST_PATHS @account TO @node GRAPH @permissions RETURN path.edges[0].level`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"account": ACCOUNTS_COL + "/" + account,
		"node": node,
		"permissions": PERMISSIONS_GRAPH.Name,
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