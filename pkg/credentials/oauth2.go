package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	oauth2_config "github.com/slntopp/nocloud/pkg/oauth2/config"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

var cfg map[string]oauth2_config.OAuth2Config

func init() {
	config, err := oauth2_config.Config()
	if err != nil {
		return
	}

	cfg = config
}

type OAuth2Credentials struct {
	AuthField string `json:"auth_field"`
	AuthValue string `json:"auth_value"`

	AuthType string `json:"auth_type"`

	log *zap.Logger
	driver.DocumentMeta
}

func NewOAuth2Credentials(data []string, credType string) (Credentials, error) {
	if len(data) != 1 {
		return nil, fmt.Errorf("some credentials data is missing, expected data length to be 1, got: %d", len(data))
	}
	token := data[0]

	oauth2Type := strings.Split(credType, "-")
	if len(oauth2Type) != 2 {
		return nil, fmt.Errorf("wrong oauth2 type, got: %s", credType)
	}

	oauth2TypeValue := oauth2Type[1]

	oauth2TypeConfig := cfg[oauth2TypeValue]

	var req *http.Request

	if oauth2TypeValue == "github" {
		request, err := http.NewRequest("GET", oauth2TypeConfig.UserInfoURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request, got: %s", err.Error())
		}
		request.Header.Set("Accept", "application/vnd.github+json")
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		req = request
	} else {
		request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", oauth2TypeConfig.UserInfoURL, token), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request, got: %s", err.Error())
		}
		req = request
	}

	client := http.Client{}

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request, got: %s", err.Error())
	}

	var bodyMap = map[string]any{}

	defer response.Body.Close()
	all, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to do read body, got: %s", err.Error())
	}

	fmt.Println(string(all))

	if oauth2TypeValue == "bitrix" {
		var responseBody = map[string]any{}
		err := json.Unmarshal(all, &responseBody)
		if err != nil {
			return nil, fmt.Errorf("failed to do unmarshal, got: %s", err.Error())
		}

		bodyMap = responseBody["result"].(map[string]any)
	} else {
		err := json.Unmarshal(all, &bodyMap)
		if err != nil {
			return nil, fmt.Errorf("failed to do unmarshal, got: %s", err.Error())
		}
	}

	authValue := bodyMap[oauth2TypeConfig.AuthField].(string)

	fmt.Println("authValue", authValue)
	fmt.Println("authField", oauth2TypeConfig.AuthField)

	return &OAuth2Credentials{AuthField: oauth2TypeConfig.AuthField, AuthValue: authValue, AuthType: credType}, nil
}

func (cred *OAuth2Credentials) Type() string {
	return cred.AuthType
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
	query := `FOR cred IN @@credentials FILTER cred.auth_field == @field AND cred.auth_value == @value && cred.auth_type == @type RETURN cred`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"field":        cred.AuthField,
		"value":        cred.AuthValue,
		"type":         cred.AuthType,
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
