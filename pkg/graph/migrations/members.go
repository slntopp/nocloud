package migrations

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

// Subaccounts used to get an edge from their namespace to the mother account. AccessLevel takes
// the first edge's level, so that gave them ADMIN over the mother and everything she owns.
const unlinkMembersFromOwners = `
FOR a IN @@accounts
    FILTER a.account_owner != null AND a.account_owner != ""
    FOR ns IN 1 OUTBOUND a @@acc2ns
        FOR e IN @@ns2acc
            FILTER e._from == ns._id AND e._to == CONCAT(@accounts, "/", a.account_owner)
            REMOVE e IN @@ns2acc
            LET owner = DOCUMENT(@@accounts, a.account_owner)
            RETURN DISTINCT { member: a._key, resume: a.suspended == true AND owner.suspended == true }
`

const resumeMember = `UPDATE @member WITH { suspended: false } IN @@accounts`

// UnlinkMembersFromOwners removes those edges and signs the subaccounts out: their tokens have no
// member claim yet, so the member gate would not see them. Billing no longer suspends subaccounts
// with their owner, nor resumes them, so a subaccount suspended along with its suspended owner is
// resumed here: the owner did not switch it off. Each subaccount is migrated once: unlinked, it is
// not found again.
func UnlinkMembersFromOwners(log *zap.Logger, db driver.Database, rdb redisdb.Client) {
	log = log.Named("UnlinkMembersFromOwners")
	ctx := context.Background()

	c, err := db.Query(ctx, unlinkMembersFromOwners, map[string]interface{}{
		"@accounts": schema.ACCOUNTS_COL,
		"@acc2ns":   schema.ACC2NS,
		"@ns2acc":   schema.NS2ACC,
		"accounts":  schema.ACCOUNTS_COL,
	})
	if err != nil {
		log.Error("Failed to unlink members from owners", zap.Error(err))
		return
	}
	defer c.Close()

	for {
		var m struct {
			Member string `json:"member"`
			Resume bool   `json:"resume"`
		}
		if _, err := c.ReadDocument(ctx, &m); driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Error("Failed to read unlinked member", zap.Error(err))
			continue
		}
		if m.Resume {
			if _, err := db.Query(ctx, resumeMember, map[string]interface{}{
				"@accounts": schema.ACCOUNTS_COL,
				"member":    m.Member,
			}); err != nil {
				log.Error("Failed to resume the member", zap.String("member", m.Member), zap.Error(err))
			}
		}
		keys, err := rdb.Keys(ctx, fmt.Sprintf("sessions:%s:*", m.Member)).Result()
		if err == nil && len(keys) > 0 {
			err = rdb.Del(ctx, keys...).Err()
		}
		if err != nil {
			log.Error("Failed to sign the member out", zap.String("member", m.Member), zap.Error(err))
		}
		log.Info("Member unlinked from its owner", zap.String("member", m.Member),
			zap.Bool("resumed", m.Resume), zap.Int("sessions", len(keys)))
	}
}
