package credentials

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type OAuth2Credentials struct {
	AuthField string `json:"auth_field"`
	AuthValue string `json:"auth_value"`

	log *zap.Logger
	driver.DocumentMeta
}

func NewOAuth2Credentials(data []string) (Credentials, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("some credentials data is missing, expected data length to be 2, got: %d", len(data))
	}
	field, value := data[0], data[1]
	return &OAuth2Credentials{AuthField: field, AuthValue: value}, nil
}

func (*OAuth2Credentials) Type() string {
	return "oauth2-github"
}

// Authorize method for StandardCredentials assumes that args consist of username and password stored at 0 and 1 accordingly
func (cred *OAuth2Credentials) Authorize(args ...string) bool {
	return cred.AuthField == args[0] && cred.AuthValue == args[1]
}

func (cred *OAuth2Credentials) SetLogger(log *zap.Logger) {
	cred.log = log.Named("OAuth2")
	cred.log.Debug("Logger is now set")
}

func (cred *OAuth2Credentials) Find(ctx context.Context, db driver.Database) bool {
	query := `FOR cred IN @@credentials FILTER cred[@field] == @value RETURN cred`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"field":        cred.AuthField,
		"value":        cred.AuthValue,
		"@credentials": schema.CREDENTIALS_COL,
	})
	if err != nil {
		return false
	}
	defer c.Close()

	_, err = c.ReadDocument(ctx, &cred)
	return err == nil
}

func (cred *OAuth2Credentials) FindByKey(ctx context.Context, col driver.Collection, key string) error {
	_, err := col.ReadDocument(ctx, key, cred)
	return err
}
