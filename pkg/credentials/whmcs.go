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
package credentials

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/slntopp/nocloud-proto/access"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"

	"github.com/arangodb/go-driver"
	sc "github.com/slntopp/nocloud/pkg/settings/client"
)

type WHMCSCredentials struct {
	Email string `json:"email"`

	log *zap.Logger
	driver.DocumentMeta
}

type WHMCSConfig struct {
	Api        string `json:"api"`
	User       string `json:"user"`
	Pass       string `json:"pass_hash"`
	DangerMode bool   `json:"danger_mode"`
}

func NewWHMCSCredentials(data []string) (Credentials, error) {
	if len(data) < 1 {
		return nil, errors.New("empty Credentials")
	}
	email := data[0]
	return &WHMCSCredentials{Email: email}, nil
}

func (*WHMCSCredentials) Type() string {
	return "whmcs"
}

func (c *WHMCSCredentials) SetLogger(log *zap.Logger) {
	c.log = log.Named("WHMCS")
	c.log.Debug("Logger is now set")
}

func (c *WHMCSCredentials) Authorize(args ...string) bool {
	log := c.log.Named("Authorize")
	log.Debug("Authorize Attempt", zap.String("login", args[0]), zap.String("email", c.Email))

	conf := &WHMCSConfig{}
	err := _GetWHMCSConfig(conf)
	if err != nil {
		c.log.Error("Error getting settings", zap.Error(err))
		return false
	}
	log.Debug("Attempt request to WHMCS", zap.Any("conf", conf))

	if conf.Api == "" || conf.User == "" || conf.Pass == "" {
		log.Error("Some settings are empty", zap.Any("conf", conf))
		return false
	}

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("email", c.Email)
	_ = writer.WriteField("password2", args[1])
	_ = writer.WriteField("username", conf.User)
	_ = writer.WriteField("password", conf.Pass)
	_ = writer.WriteField("action", "ValidateLogin")
	err = writer.Close()
	if err != nil {
		log.Error("Error writing FormData", zap.Error(err))
		return false
	}

	client := &http.Client{}
	if conf.DangerMode {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	req, err := http.NewRequest("POST", conf.Api, payload)
	if err != nil {
		log.Error("Error making Request", zap.Error(err))
		return false
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		log.Error("Error performing HTTP request", zap.Error(err))
		return false
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error("Error reading Response Body", zap.Error(err))
		return false
	}

	log.Debug("Response", zap.Int("code", res.StatusCode), zap.ByteString("body", body))
	for _, el := range strings.Split(string(body), ";") {
		data := strings.Split(el, "=")
		if data[0] == "result" {
			return data[1] == "success"
		}
	}

	log.Debug("No result found", zap.String("body", string(body)))
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

func _GetWHMCSConfig(conf *WHMCSConfig) error {
	return sc.Fetch("whmcs", conf, &sc.Setting[WHMCSConfig]{
		Value:       *conf,
		Description: "WHMCS Credentials Settings (API Endpoint, username, password)",
		Level:       access.Level_ADMIN,
	})
}
