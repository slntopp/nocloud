package graph

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/slntopp/nocloud/pkg/oauth2"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
)

type OAuthController struct {
	log *zap.Logger
	db  driver.Database

	clientsCol      driver.Collection
	codesCol        driver.Collection
	accessCol       driver.Collection
	refreshCol      driver.Collection
	interactionsCol driver.Collection
}

type ArangoRepoOptions struct {
	ClientsCollection      string
	CodesCollection        string
	AccessCollection       string
	RefreshCollection      string
	InteractionsCollection string
}

func DefaultArangoRepoOptions() ArangoRepoOptions {
	return ArangoRepoOptions{
		ClientsCollection:      "OAuth2Clients",
		CodesCollection:        "OAuth2AuthorizationCodes",
		AccessCollection:       "OAuth2AccessTokens",
		RefreshCollection:      "OAuth2RefreshTokens",
		InteractionsCollection: "OAuth2Interactions",
	}
}

func NewOAuthController(logger *zap.Logger, db driver.Database, opts *ArangoRepoOptions) *OAuthController {
	if opts == nil {
		o := DefaultArangoRepoOptions()
		opts = &o
	}

	log := logger.Named("OAuthController")
	ctx := context.TODO()

	clients, err := ensureCollection(ctx, db, opts.ClientsCollection)
	if err != nil {
		log.Fatal("ensure clients collection failed", zap.Error(err))
	}
	codes, err := ensureCollection(ctx, db, opts.CodesCollection)
	if err != nil {
		log.Fatal("ensure codes collection failed", zap.Error(err))
	}
	access, err := ensureCollection(ctx, db, opts.AccessCollection)
	if err != nil {
		log.Fatal("ensure access collection failed", zap.Error(err))
	}
	refresh, err := ensureCollection(ctx, db, opts.RefreshCollection)
	if err != nil {
		log.Fatal("ensure refresh collection failed", zap.Error(err))
	}
	interaction, err := ensureCollection(ctx, db, opts.InteractionsCollection)
	if err != nil {
		log.Fatal("ensure interactions collection failed", zap.Error(err))
	}

	return &OAuthController{
		log:             log,
		db:              db,
		clientsCol:      clients,
		codesCol:        codes,
		accessCol:       access,
		refreshCol:      refresh,
		interactionsCol: interaction,
	}
}

var _ oauth2.ClientStore = (*OAuthController)(nil)
var _ oauth2.AuthorizationCodeStore = (*OAuthController)(nil)
var _ oauth2.TokenStore = (*OAuthController)(nil)
var _ oauth2.InteractionStore = (*OAuthController)(nil)

type clientDoc struct {
	Key          string   `json:"_key,omitempty"`
	Secret       string   `json:"secret,omitempty"`
	RedirectURIs []string `json:"redirect_uris,omitempty"`

	AllowedGrants []string `json:"allowed_grants,omitempty"`
	AllowedScopes []string `json:"allowed_scopes,omitempty"`

	Public bool `json:"public,omitempty"`

	Display map[string]any `json:"display,omitempty"`
}

func (d clientDoc) toClient() oauth2.Client {
	displayBytes, _ := json.Marshal(d.Display)
	var display oauth2.ClientDisplay
	_ = json.Unmarshal(displayBytes, &display)
	return oauth2.Client{
		ID:            d.Key,
		Secret:        d.Secret,
		RedirectURIs:  d.RedirectURIs,
		AllowedGrants: d.AllowedGrants,
		AllowedScopes: d.AllowedScopes,
		Public:        d.Public,
		Display:       display,
	}
}

func (r *OAuthController) GetClient(ctx context.Context, clientID string) (oauth2.Client, error) {
	log := r.log.Named("GetClient")
	var d clientDoc
	meta, err := r.clientsCol.ReadDocument(ctx, clientID, &d)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return oauth2.Client{}, err
		}
		log.Error("read client failed", zap.Error(err))
		return oauth2.Client{}, err
	}
	d.Key = meta.Key
	return d.toClient(), nil
}

func (r *OAuthController) ValidateClientSecret(ctx context.Context, clientID, clientSecret string) (bool, error) {
	c, err := r.GetClient(ctx, clientID)
	if err != nil {
		return false, err
	}
	if c.Public || c.Secret == "" {
		return clientSecret == "", nil
	}
	ok := subtle.ConstantTimeCompare([]byte(c.Secret), []byte(clientSecret)) == 1
	return ok, nil
}

type authCodeDoc struct {
	Key string `json:"_key,omitempty"`

	Code string `json:"code"`

	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`

	Subject string   `json:"subject"`
	Scopes  []string `json:"scopes,omitempty"`

	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Consumed  bool      `json:"consumed"`
}

func toAuthCodeDoc(c oauth2.AuthorizationCode) authCodeDoc {
	return authCodeDoc{
		Key:         opaqueKey(c.Code),
		Code:        c.Code,
		ClientID:    c.ClientID,
		RedirectURI: c.RedirectURI,
		Subject:     c.Subject,
		Scopes:      c.Scopes,
		IssuedAt:    c.IssuedAt.UTC(),
		ExpiresAt:   c.ExpiresAt.UTC(),
		Consumed:    c.Consumed,
	}
}

func (d authCodeDoc) toAuthorizationCode() oauth2.AuthorizationCode {
	return oauth2.AuthorizationCode{
		Code:        d.Code,
		ClientID:    d.ClientID,
		RedirectURI: d.RedirectURI,
		Subject:     d.Subject,
		Scopes:      d.Scopes,
		IssuedAt:    d.IssuedAt.UTC(),
		ExpiresAt:   d.ExpiresAt.UTC(),
		Consumed:    d.Consumed,
	}
}

func (r *OAuthController) Create(ctx context.Context, code oauth2.AuthorizationCode) error {
	log := r.log.Named("AuthorizationCode.Create")
	if code.Code == "" {
		return fmt.Errorf("missing authorization code")
	}
	doc := toAuthCodeDoc(code)
	_, err := r.codesCol.CreateDocument(ctx, doc)
	if err != nil {
		log.Error("create authorization code failed", zap.Error(err))
		return err
	}
	return nil
}

const consumeCodeAQL = `
LET nowTs = DATE_TIMESTAMP(@now)
LET doc = FIRST(
  FOR c IN @@codes
    FILTER c._key == @key
    FILTER c.consumed != true
    FILTER DATE_TIMESTAMP(c.expires_at) > nowTs
    UPDATE c WITH { consumed: true } IN @@codes OPTIONS { ignoreRevs: false }
    RETURN NEW
)
RETURN doc
`

func (r *OAuthController) Consume(ctx context.Context, code string) (oauth2.AuthorizationCode, error) {
	log := r.log.Named("AuthorizationCode.Consume")

	if code == "" {
		return oauth2.AuthorizationCode{}, fmt.Errorf("missing authorization code")
	}

	key := opaqueKey(code)

	cur, err := r.db.Query(ctx, consumeCodeAQL, map[string]any{
		"@codes": r.codesCol.Name(),
		"key":    key,
		"now":    time.Now().UTC(),
	})
	if err != nil {
		log.Error("consume authorization code query failed", zap.Error(err))
		return oauth2.AuthorizationCode{}, err
	}
	defer cur.Close()

	var out *authCodeDoc
	_, err = cur.ReadDocument(ctx, &out)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			return oauth2.AuthorizationCode{}, fmt.Errorf("invalid authorization code")
		}
		log.Error("consume authorization code read failed", zap.Error(err))
		return oauth2.AuthorizationCode{}, err
	}
	if out == nil || out.Code != code {
		return oauth2.AuthorizationCode{}, fmt.Errorf("invalid authorization code")
	}

	return out.toAuthorizationCode(), nil
}

type accessTokenDoc struct {
	Key string `json:"_key,omitempty"`

	Token string `json:"token"`

	ClientID string   `json:"client_id"`
	Subject  string   `json:"subject"`
	Scopes   []string `json:"scopes,omitempty"`

	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`

	Revoked bool `json:"revoked"`
}

func toAccessTokenDoc(t oauth2.AccessToken) accessTokenDoc {
	return accessTokenDoc{
		Key:       opaqueKey(t.Token),
		Token:     t.Token,
		ClientID:  t.ClientID,
		Subject:   t.Subject,
		Scopes:    t.Scopes,
		IssuedAt:  t.IssuedAt.UTC(),
		ExpiresAt: t.ExpiresAt.UTC(),
		Revoked:   t.Revoked,
	}
}

func (d accessTokenDoc) toAccessToken() oauth2.AccessToken {
	return oauth2.AccessToken{
		Token:     d.Token,
		ClientID:  d.ClientID,
		Subject:   d.Subject,
		Scopes:    d.Scopes,
		IssuedAt:  d.IssuedAt.UTC(),
		ExpiresAt: d.ExpiresAt.UTC(),
		Revoked:   d.Revoked,
	}
}

type refreshTokenDoc struct {
	Key string `json:"_key,omitempty"`

	Token string `json:"token"`

	ClientID string   `json:"client_id"`
	Subject  string   `json:"subject"`
	Scopes   []string `json:"scopes,omitempty"`

	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`

	Revoked bool `json:"revoked"`

	AuthorizationCode string `json:"authorization_code,omitempty"`
}

func toRefreshTokenDoc(t oauth2.RefreshToken) refreshTokenDoc {
	return refreshTokenDoc{
		Key:               opaqueKey(t.Token),
		Token:             t.Token,
		ClientID:          t.ClientID,
		Subject:           t.Subject,
		Scopes:            t.Scopes,
		IssuedAt:          t.IssuedAt.UTC(),
		ExpiresAt:         t.ExpiresAt.UTC(),
		Revoked:           t.Revoked,
		AuthorizationCode: t.AuthorizationCode,
	}
}

func (d refreshTokenDoc) toRefreshToken() oauth2.RefreshToken {
	return oauth2.RefreshToken{
		Token:             d.Token,
		ClientID:          d.ClientID,
		Subject:           d.Subject,
		Scopes:            d.Scopes,
		IssuedAt:          d.IssuedAt.UTC(),
		ExpiresAt:         d.ExpiresAt.UTC(),
		Revoked:           d.Revoked,
		AuthorizationCode: d.AuthorizationCode,
	}
}

const upsertReplaceAQL = `
UPSERT { _key: @key }
INSERT MERGE(@doc, { _key: @key })
REPLACE MERGE(@doc, { _key: @key }) IN @@col
RETURN 1
`

func (r *OAuthController) SaveAccessToken(ctx context.Context, t oauth2.AccessToken) error {
	log := r.log.Named("SaveAccessToken")

	if t.Token == "" {
		return fmt.Errorf("missing access token")
	}

	key := opaqueKey(t.Token)
	doc := map[string]any{
		"token":      t.Token,
		"client_id":  t.ClientID,
		"subject":    t.Subject,
		"scopes":     t.Scopes,
		"issued_at":  t.IssuedAt.UTC(),
		"expires_at": t.ExpiresAt.UTC(),
		"revoked":    t.Revoked,
	}

	cur, err := r.db.Query(ctx, upsertReplaceAQL, map[string]any{
		"@col": r.accessCol.Name(),
		"key":  key,
		"doc":  doc,
	})
	if err != nil {
		log.Error("upsert access token failed", zap.Error(err))
		return err
	}
	defer cur.Close()

	var dummy int
	_, _ = cur.ReadDocument(ctx, &dummy)
	return nil
}

func (r *OAuthController) SaveRefreshToken(ctx context.Context, t oauth2.RefreshToken) error {
	log := r.log.Named("SaveRefreshToken")

	if t.Token == "" {
		return fmt.Errorf("missing refresh token")
	}

	key := opaqueKey(t.Token)
	doc := map[string]any{
		"token":              t.Token,
		"client_id":          t.ClientID,
		"subject":            t.Subject,
		"scopes":             t.Scopes,
		"issued_at":          t.IssuedAt.UTC(),
		"expires_at":         t.ExpiresAt.UTC(),
		"revoked":            t.Revoked,
		"authorization_code": t.AuthorizationCode,
	}

	cur, err := r.db.Query(ctx, upsertReplaceAQL, map[string]any{
		"@col": r.refreshCol.Name(),
		"key":  key,
		"doc":  doc,
	})
	if err != nil {
		log.Error("upsert refresh token failed", zap.Error(err))
		return err
	}
	defer cur.Close()

	var dummy int
	_, _ = cur.ReadDocument(ctx, &dummy)
	return nil
}

func (r *OAuthController) LookupAccessToken(ctx context.Context, token string) (oauth2.AccessToken, error) {
	log := r.log.Named("LookupAccessToken")

	if token == "" {
		return oauth2.AccessToken{}, fmt.Errorf("missing access token")
	}

	key := opaqueKey(token)
	var d accessTokenDoc
	_, err := r.accessCol.ReadDocument(ctx, key, &d)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return oauth2.AccessToken{}, fmt.Errorf("token not found")
		}
		log.Error("read access token failed", zap.Error(err))
		return oauth2.AccessToken{}, err
	}

	if d.Token != token {
		return oauth2.AccessToken{}, fmt.Errorf("token not found")
	}

	now := time.Now().UTC()
	if d.Revoked || (!d.ExpiresAt.UTC().IsZero() && !d.ExpiresAt.UTC().After(now)) {
		return oauth2.AccessToken{}, fmt.Errorf("token is invalid")
	}

	return d.toAccessToken(), nil
}

func (r *OAuthController) LookupRefreshToken(ctx context.Context, token string) (oauth2.RefreshToken, error) {
	log := r.log.Named("LookupRefreshToken")

	if token == "" {
		return oauth2.RefreshToken{}, fmt.Errorf("missing refresh token")
	}

	key := opaqueKey(token)
	var d refreshTokenDoc
	_, err := r.refreshCol.ReadDocument(ctx, key, &d)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return oauth2.RefreshToken{}, fmt.Errorf("invalid refresh token")
		}
		log.Error("read refresh token failed", zap.Error(err))
		return oauth2.RefreshToken{}, err
	}

	if d.Token != token {
		return oauth2.RefreshToken{}, fmt.Errorf("invalid refresh token")
	}

	now := time.Now().UTC()
	if d.Revoked || (!d.ExpiresAt.UTC().IsZero() && !d.ExpiresAt.UTC().After(now)) {
		return oauth2.RefreshToken{}, fmt.Errorf("invalid refresh token")
	}

	return d.toRefreshToken(), nil
}

const revokeAQL = `
UPDATE { _key: @key } WITH { revoked: true } IN @@col OPTIONS { ignoreErrors: true }
RETURN 1
`

func (r *OAuthController) RevokeAccessToken(ctx context.Context, token string) error {
	log := r.log.Named("RevokeAccessToken")

	if token == "" {
		return nil
	}

	cur, err := r.db.Query(ctx, revokeAQL, map[string]any{
		"@col": r.accessCol.Name(),
		"key":  opaqueKey(token),
	})
	if err != nil {
		log.Error("revoke access token failed", zap.Error(err))
		return err
	}
	defer cur.Close()

	var dummy int
	_, _ = cur.ReadDocument(ctx, &dummy)
	return nil
}

func (r *OAuthController) RevokeRefreshToken(ctx context.Context, token string) error {
	log := r.log.Named("RevokeRefreshToken")

	if token == "" {
		return nil
	}

	cur, err := r.db.Query(ctx, revokeAQL, map[string]any{
		"@col": r.refreshCol.Name(),
		"key":  opaqueKey(token),
	})
	if err != nil {
		log.Error("revoke refresh token failed", zap.Error(err))
		return err
	}
	defer cur.Close()

	var dummy int
	_, _ = cur.ReadDocument(ctx, &dummy)
	return nil
}

type interactionDoc struct {
	Key string `json:"_key,omitempty"`

	ID          string `json:"id"`
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
	State       string `json:"state"`

	Subject         string   `json:"subject,omitempty"`
	RequestedScopes []string `json:"requested_scopes,omitempty"`
	ExistingScopes  []string `json:"existing_scopes,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Consumed  bool      `json:"consumed"`
}

func toInteractionDoc(it oauth2.Interaction) interactionDoc {
	return interactionDoc{
		Key:             opaqueKey(it.ID),
		ID:              it.ID,
		ClientID:        it.ClientID,
		RedirectURI:     it.RedirectURI,
		State:           it.State,
		Subject:         it.Subject,
		RequestedScopes: it.RequestedScopes,
		ExistingScopes:  it.ExistingScopes,
		CreatedAt:       it.CreatedAt.UTC(),
		ExpiresAt:       it.ExpiresAt.UTC(),
		Consumed:        it.Consumed,
	}
}

func (d interactionDoc) toInteraction() oauth2.Interaction {
	return oauth2.Interaction{
		ID:              d.ID,
		ClientID:        d.ClientID,
		RedirectURI:     d.RedirectURI,
		State:           d.State,
		Subject:         d.Subject,
		RequestedScopes: d.RequestedScopes,
		ExistingScopes:  d.ExistingScopes,
		CreatedAt:       d.CreatedAt.UTC(),
		ExpiresAt:       d.ExpiresAt.UTC(),
		Consumed:        d.Consumed,
	}
}

func (r *OAuthController) CreateInteraction(ctx context.Context, it oauth2.Interaction) error {
	log := r.log.Named("Interaction.Create")

	if it.ID == "" {
		return fmt.Errorf("missing interaction id")
	}
	if it.ClientID == "" {
		return fmt.Errorf("missing interaction client id")
	}
	if it.CreatedAt.IsZero() {
		it.CreatedAt = time.Now().UTC()
	}

	doc := toInteractionDoc(it)
	_, err := r.interactionsCol.CreateDocument(ctx, doc)
	if err != nil {
		log.Error("create interaction failed", zap.Error(err))
		return err
	}
	return nil
}

func (r *OAuthController) GetInteraction(ctx context.Context, id string) (oauth2.Interaction, error) {
	log := r.log.Named("Interaction.Get")

	if id == "" {
		return oauth2.Interaction{}, fmt.Errorf("missing interaction id")
	}

	key := opaqueKey(id)
	var d interactionDoc
	_, err := r.interactionsCol.ReadDocument(ctx, key, &d)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return oauth2.Interaction{}, fmt.Errorf("interaction not found")
		}
		log.Error("read interaction failed", zap.Error(err))
		return oauth2.Interaction{}, err
	}

	if d.ID != id {
		return oauth2.Interaction{}, fmt.Errorf("interaction not found")
	}

	now := time.Now().UTC()
	if d.Consumed || (!d.ExpiresAt.UTC().IsZero() && !d.ExpiresAt.UTC().After(now)) {
		return oauth2.Interaction{}, fmt.Errorf("interaction is invalid")
	}

	return d.toInteraction(), nil
}

func (r *OAuthController) ConsumeInteraction(ctx context.Context, id string) (oauth2.Interaction, error) {
	log := r.log.Named("Interaction.Consume")

	if id == "" {
		return oauth2.Interaction{}, fmt.Errorf("missing interaction id")
	}

	key := opaqueKey(id)
	now := time.Now().UTC()

	tid, err := r.db.BeginTransaction(
		ctx,
		driver.TransactionCollections{
			Write: []string{r.interactionsCol.Name()},
		},
		&driver.BeginTransactionOptions{
			AllowImplicit: false,
		},
	)
	if err != nil {
		log.Error("Begin transaction failed", zap.Error(err))
		return oauth2.Interaction{}, err
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		abortCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if abortErr := r.db.AbortTransaction(abortCtx, tid, nil); abortErr != nil {
			log.Warn("Abort transaction failed", zap.Error(abortErr))
		}
	}()
	trxCtx := driver.WithTransactionID(ctx, tid)

	const aql = `
LET doc = DOCUMENT(@@its, @key)
LET _valid = (doc != null && doc.ID == @id) ? 1 : FAIL("interaction_not_found")
LET _once  = (!doc.consumed) ? 1 : FAIL("interaction_already_consumed")
UPDATE doc WITH { consumed: true } IN @@its OPTIONS { keepNull: false }
RETURN NEW
`
	cur, err := r.db.Query(trxCtx, aql, map[string]any{
		"@its": r.interactionsCol.Name(),
		"key":  key,
		"id":   id,
		"now":  now,
	})
	if err != nil {
		if ae, ok := driver.AsArangoError(err); ok && ae.ErrorNum == 1569 {
			switch {
			case strings.Contains(ae.ErrorMessage, "interaction_already_consumed"):
				return oauth2.Interaction{}, fmt.Errorf("interaction already consumed")
			case strings.Contains(ae.ErrorMessage, "interaction_not_found"):
				return oauth2.Interaction{}, fmt.Errorf("invalid interaction. not found")
			}
		}
		log.Error("Consume interaction query failed", zap.Error(err))
		return oauth2.Interaction{}, err
	}
	defer cur.Close()

	var out *interactionDoc
	_, err = cur.ReadDocument(trxCtx, &out)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			return oauth2.Interaction{}, fmt.Errorf("invalid interaction. not found")
		}
		log.Error("Consume interaction read failed", zap.Error(err))
		return oauth2.Interaction{}, err
	}
	if out == nil || out.ID != id {
		return oauth2.Interaction{}, fmt.Errorf("invalid interaction. not found")
	}

	_ = cur.Close()
	if err := r.db.CommitTransaction(ctx, tid, nil); err != nil {
		log.Error("Consume interaction commit failed", zap.Error(err))
		return oauth2.Interaction{}, err
	}

	committed = true
	return out.toInteraction(), nil
}

func ensureCollection(ctx context.Context, db driver.Database, name string) (driver.Collection, error) {
	col, err := db.Collection(ctx, name)
	if err == nil {
		return col, nil
	}
	if driver.IsNotFoundGeneral(err) {
		return db.CreateCollection(ctx, name, nil)
	}
	return nil, err
}

func opaqueKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
