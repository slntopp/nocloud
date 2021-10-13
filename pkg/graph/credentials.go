package graph

import (
	"github.com/arangodb/go-driver"
	"golang.org/x/crypto/bcrypt"
)

var (
	CREDENTIALS_COL = "Credentials"
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
}

type StandardCredentials struct {
	PasswordHash string `json:"password_hash"`

	driver.DocumentMeta
}

func NewStandardCredentials(password string) (Credentials, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return &StandardCredentials{
		PasswordHash: string(bytes),
	}, err
}

// Authorize method for StandardCredentials assumes that args consist only of password stored at 0
func (c *StandardCredentials) Authorize(args ...string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(c.PasswordHash), []byte(args[0]))
    return err == nil
}

func (*StandardCredentials) Type() string {
	return "standard"
}