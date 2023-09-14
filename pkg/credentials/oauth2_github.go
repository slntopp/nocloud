package credentials

import (
	"context"
	"errors"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type OAuth2GithubCredentials struct {
	Login string `json:"login"`

	log *zap.Logger
	driver.DocumentMeta
}

func NewOAuth2GithubCredentials(data []string) (Credentials, error) {
	if len(data) < 1 {
		return nil, errors.New("empty Credentials")
	}
	login := data[0]
	return &OAuth2GithubCredentials{Login: login}, nil
}

func (*OAuth2GithubCredentials) Type() string {
	return "oauth2-github"
}

// Authorize method for StandardCredentials assumes that args consist of username and password stored at 0 and 1 accordingly
func (cred *OAuth2GithubCredentials) Authorize(args ...string) bool {
	return cred.Login == args[0]
}

func (cred *OAuth2GithubCredentials) SetLogger(log *zap.Logger) {
	cred.log = log.Named("OAuth2Github")
	cred.log.Debug("Logger is now set")
}

func (cred *OAuth2GithubCredentials) Find(ctx context.Context, db driver.Database) bool {
	query := `FOR cred IN @@credentials FILTER cred.login == @login RETURN cred`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"login":        cred.Login,
		"@credentials": schema.CREDENTIALS_COL,
	})
	if err != nil {
		return false
	}
	defer c.Close()

	_, err = c.ReadDocument(ctx, &cred)
	return err == nil
}

func (cred *OAuth2GithubCredentials) FindByKey(ctx context.Context, col driver.Collection, key string) error {
	_, err := col.ReadDocument(ctx, key, cred)
	return err
}
