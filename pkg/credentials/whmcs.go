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
package credentials

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud/pkg/settings/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var settingsClient pb.SettingsServiceClient

type WHMCSCredentials struct {
	Email string `json:"email"`

	log *zap.Logger
	driver.DocumentMeta
}

func NewWHMCSCredentials(email string) (Credentials, error) {
	return &WHMCSCredentials{Email: email}, nil
}

func (*WHMCSCredentials) Type() string {
	return "whmcs"
}

func (c *WHMCSCredentials) SetLogger(log *zap.Logger) {
	c.log = log.Named("WHMCS Auth")
}

func (c *WHMCSCredentials) Authorize(args ...string) bool {
	vars, err := settingsClient.Get(context.Background(), &pb.GetRequest{Keys: []string{
		"whmcs:api", "whmcs:user", "whmcs:pass_hash",
	}})
	if err != nil {
		c.log.Error("Error getting settings", zap.Error(err))
		return false
	}

	api := vars.Fields["whmcs:api"].GetStringValue()
	user := vars.Fields["whmcs:user"].GetStringValue()
	pass := vars.Fields["whmcs:pass_hash"].GetStringValue()
	if api == "" || user == "" || pass == "" {
		c.log.Error("Some settings are empty", zap.Strings("vars", []string{api, user, pass}))
		return false
	}

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("email", c.Email)
	_ = writer.WriteField("password2", args[1])
	_ = writer.WriteField("username", user)
	_ = writer.WriteField("password", pass)
	_ = writer.WriteField("action", "ValidateLogin")
	err = writer.Close()
	if err != nil {
		c.log.Error("Error writing FormData", zap.Error(err))
		return false
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", api, payload)
	if err != nil {
		c.log.Error("Error making Request", zap.Error(err))
		return false
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		c.log.Error("Error performing HTTP request", zap.Error(err))
		return false
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	for _, el := range strings.Split(string(body), ";") {
		data := strings.Split(el, "=")
		if data[0] == "result" {
			return data[1] == "success"
		}
	}

	c.log.Debug("No result found", zap.String("body", string(body)))
	return false
}

func (cred *WHMCSCredentials) Find(ctx context.Context, db driver.Database) bool {
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

func (cred *WHMCSCredentials) FindByKey(ctx context.Context, col driver.Collection, key string) error {
	_, err := col.ReadDocument(ctx, key, cred)
	return err
}

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("SETTINGS_HOST", "settings:8000")
	host := viper.GetString("SETTINGS_HOST")

	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	settingsClient = pb.NewSettingsServiceClient(conn)
}
