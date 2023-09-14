package credentials

import (
	"context"
	"errors"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type OAuth2GoogleCredentials struct {
	Email string `json:"email"`

	log *zap.Logger
	driver.DocumentMeta
}

func NewOAuth2GoogleCredentials(data []string) (Credentials, error) {
	if len(data) < 1 {
		return nil, errors.New("empty Credentials")
	}
	email := data[0]
	return &OAuth2GoogleCredentials{Email: email}, nil
}

func (*OAuth2GoogleCredentials) Type() string {
	return "oauth2-google"
}

// Authorize method for StandardCredentials assumes that args consist of username and password stored at 0 and 1 accordingly
func (cred *OAuth2GoogleCredentials) Authorize(args ...string) bool {
	return cred.Email == args[0]
}

func (cred *OAuth2GoogleCredentials) SetLogger(log *zap.Logger) {
	cred.log = log.Named("OAuth2Google")
	cred.log.Debug("Logger is now set")
}

func (cred *OAuth2GoogleCredentials) Find(ctx context.Context, db driver.Database) bool {
	query := `FOR cred IN @@credentials FILTER cred.email == @email RETURN cred`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"email":        cred.Email,
		"@credentials": schema.CREDENTIALS_COL,
	})
	if err != nil {
		return false
	}
	defer c.Close()

	_, err = c.ReadDocument(ctx, &cred)
	return err == nil
}

func (cred *OAuth2GoogleCredentials) FindByKey(ctx context.Context, col driver.Collection, key string) error {
	_, err := col.ReadDocument(ctx, key, cred)
	return err
}
