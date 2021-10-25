/*
Copyright © 2021 Nikita Ivanovski info@slnt-opp.xyz

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
	"golang.org/x/crypto/bcrypt"
)

const (
	CREDENTIALS_COL = "Credentials"
	CREDENTIALS_EDGE_COL = ACCOUNTS_COL + "2" + CREDENTIALS_COL
)

type CredentialsLink struct {
	From driver.DocumentID `json:"_from"`
	To driver.DocumentID `json:"_to"`
	Type string `json:"type"`

	driver.DocumentMeta
}

type Credentials interface {
	Authorize(...string) bool;
	Type() string;

	Find(context.Context, driver.Database) bool;
	FindByKey(context.Context, driver.Collection, string) error;
	Account(context.Context, driver.Database) (Account, bool);

	Key() string;
}

type StandardCredentials struct {
	Username string `json:"username"`
	PasswordHash string `json:"password_hash"`

	driver.DocumentMeta
}

func NewStandardCredentials(username, password string) (Credentials, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return &StandardCredentials{
		Username: username,
		PasswordHash: string(bytes),
	}, err
}

// Authorize method for StandardCredentials assumes that args consist of username and password stored at 0 and 1 accordingly
func (c *StandardCredentials) Authorize(args ...string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(c.PasswordHash), []byte(args[1]))
    return err == nil
}

func (*StandardCredentials) Type() string {
	return "standard"
}

func (cred *StandardCredentials) Find(ctx context.Context, db driver.Database) (bool) {
	query := `FOR cred IN @@credentials FILTER cred.username == @username RETURN cred`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"username": cred.Username,
		"@credentials": CREDENTIALS_COL,
	})
	if err != nil {
		return false
	}
	defer c.Close()

	_, err = c.ReadDocument(ctx, &cred)
	return err == nil
}

func (cred *StandardCredentials) FindByKey(ctx context.Context, col driver.Collection, key string) (error) {
	_, err := col.ReadDocument(ctx, key, cred)
	return err
}

func (cred *StandardCredentials) Account(ctx context.Context, db driver.Database) (Account, bool) {
	query := `FOR account IN 1 INBOUND @credentials GRAPH @credentials_graph RETURN account`
	c, err := db.Query(ctx, query, map[string]interface{}{
		"credentials": cred.DocumentMeta.ID,
		"credentials_graph": CREDENTIALS_GRAPH.Name,
	})
	if err != nil {
		return Account{}, false
	}
	defer c.Close()

	var r Account
	_, err = c.ReadDocument(ctx, &r)
	return r, err == nil
}

func (cred *StandardCredentials) Key() (string) {
	return cred.Key()
}