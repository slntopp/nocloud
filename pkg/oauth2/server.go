package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Config struct {
	Issuer string `yaml:"issuer"`

	AuthorizePath  string `yaml:"authorize_path"`
	TokenPath      string `yaml:"token_path"`
	IntrospectPath string `yaml:"introspect_path"`
	RevokePath     string `yaml:"revoke_path"`

	LoginURL                string `yaml:"login_url"`
	ConsentURL              string `yaml:"consent_url"`
	LoginReturnParam        string `yaml:"login_return_param"`
	ConsentInteractionParam string `yaml:"consent_interaction_param"`

	InteractionPath        string        `yaml:"interaction_path"`
	InteractionConfirmPath string        `yaml:"interaction_confirm_path"`
	InteractionTTL         time.Duration `yaml:"interaction_ttl"`

	AuthorizationCodeTTL time.Duration `yaml:"authorization_code_ttl"`
	AccessTokenTTL       time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL      time.Duration `yaml:"refresh_token_ttl"`

	IssueRefreshToken     *bool `yaml:"issue_refresh_token"`
	RotateRefreshTokens   *bool `yaml:"rotate_refresh_tokens"`
	RevokeOldRefreshToken *bool `yaml:"revoke_old_refresh_token"`

	MaxBodyBytes int64 `yaml:"max_body_bytes"`

	AllowPublicClientsOnTokenEndpoint *bool `yaml:"allow_public_clients_on_token_endpoint"`
}

type Dependencies struct {
	Clients       ClientStore
	Codes         AuthorizationCodeStore
	Tokens        TokenStore
	Authorization AuthorizationService
	Interactions  InteractionStore
}

type Server struct {
	cfg  Config
	log  *zap.Logger
	deps Dependencies

	router *mux.Router
}

func NewServer(router *mux.Router, cfg Config, deps Dependencies, logger *zap.Logger) (*Server, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	if deps.Clients == nil || deps.Codes == nil || deps.Tokens == nil || deps.Authorization == nil {
		return nil, errors.New("dependencies Clients, Codes, Tokens, Authorization are required")
	}

	applyDefaults(&cfg)

	s := &Server{
		cfg:    cfg,
		log:    logger.Named("oauth2"),
		deps:   deps,
		router: router,
	}

	s.registerRoutes(s.router, "")
	return s, nil
}

type ClientDisplay struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
	WebsiteURL  string `json:"website_url"`
}

type Client struct {
	ID            string
	Secret        string
	RedirectURIs  []string
	AllowedGrants []string // authorization_code, refresh_token, client_credentials
	AllowedScopes []string
	Public        bool          // Should be false by default
	Display       ClientDisplay `json:"display"`
}

type ClientStore interface {
	GetClient(ctx context.Context, clientID string) (Client, error)
	ValidateClientSecret(ctx context.Context, clientID, clientSecret string) (bool, error)
}

type AuthorizationCode struct {
	Code        string
	ClientID    string
	RedirectURI string
	Subject     string
	Scopes      []string
	IssuedAt    time.Time
	ExpiresAt   time.Time
	Consumed    bool
}

type AuthorizationCodeStore interface {
	Create(ctx context.Context, code AuthorizationCode) error
	Consume(ctx context.Context, code string) (AuthorizationCode, error)
}

type AccessToken struct {
	Token     string
	ClientID  string
	Subject   string
	Scopes    []string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Revoked   bool
}

type RefreshToken struct {
	Token             string
	ClientID          string
	Subject           string
	Scopes            []string
	IssuedAt          time.Time
	ExpiresAt         time.Time
	Revoked           bool
	AuthorizationCode string
}

type TokenStore interface {
	SaveAccessToken(ctx context.Context, t AccessToken) error
	SaveRefreshToken(ctx context.Context, t RefreshToken) error

	LookupAccessToken(ctx context.Context, token string) (AccessToken, error)
	LookupRefreshToken(ctx context.Context, token string) (RefreshToken, error)

	RevokeAccessToken(ctx context.Context, token string) error
	RevokeRefreshToken(ctx context.Context, token string) error
}

type AuthorizationService interface {
	Subject(ctx context.Context, r *http.Request) (subject string, ok bool, err error)
	ConsentedScopes(ctx context.Context, subject, clientID string) ([]string, error)
	SaveConsent(ctx context.Context, subject, clientID string, scopes []string) error
	IssueAccessToken(ttl time.Duration, clientID, subject string, scopes []string) (AccessToken, error)
}

type Interaction struct {
	ID              string
	ClientID        string
	RedirectURI     string
	State           string
	Subject         string
	RequestedScopes []string
	ExistingScopes  []string
	CreatedAt       time.Time
	ExpiresAt       time.Time
	Consumed        bool
}

type InteractionStore interface {
	CreateInteraction(ctx context.Context, it Interaction) error
	GetInteraction(ctx context.Context, id string) (Interaction, error)
	ConsumeInteraction(ctx context.Context, id string) (Interaction, error)
}

type AuthorizeRequest struct {
	ResponseType string
	ClientID     string
	RedirectURI  string
	Scopes       []string
	State        string
}

type AuthorizeResult struct {
	Subject string
	Scopes  []string
}

type OAuthError struct {
	Code        string
	Description string
	HTTPStatus  int
}

func (e *OAuthError) Error() string {
	if e == nil {
		return ""
	}
	if e.Description == "" {
		return e.Code
	}
	return e.Code + ": " + e.Description
}

func oauthErr(code, desc string, status int) *OAuthError {
	return &OAuthError{Code: code, Description: desc, HTTPStatus: status}
}

func asOAuthError(err error) *OAuthError {
	if err == nil {
		return nil
	}
	var oe *OAuthError
	if errors.As(err, &oe) && oe != nil {
		if oe.HTTPStatus == 0 {
			oe.HTTPStatus = http.StatusBadRequest
		}
		return oe
	}
	return oauthErr("server_error", "internal error", http.StatusInternalServerError)
}

func (s *Server) registerRoutes(r *mux.Router, base string) {
	r.HandleFunc(base+s.cfg.AuthorizePath, s.handleAuthorize).Methods(http.MethodGet)
	r.HandleFunc(base+s.cfg.TokenPath, s.handleToken).Methods(http.MethodPost)
	r.HandleFunc(base+s.cfg.IntrospectPath, s.handleIntrospect).Methods(http.MethodPost)
	r.HandleFunc(base+s.cfg.RevokePath, s.handleRevoke).Methods(http.MethodPost)
	r.HandleFunc(base+s.cfg.InteractionPath+"/{id}", s.handleGetInteraction).Methods(http.MethodGet)
	r.HandleFunc(base+s.cfg.InteractionPath+"/{id}"+s.cfg.InteractionConfirmPath, s.handleConfirmInteraction).Methods(http.MethodPost)
}

func (s *Server) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, raw, err := parseAuthorizeRequest(r.URL.Query())
	if err != nil {
		s.writeAuthorizeError(w, r, raw, oauthErr("invalid_request", err.Error(), http.StatusBadRequest))
		return
	}

	client, err := s.deps.Clients.GetClient(ctx, req.ClientID)
	if err != nil {
		s.writeAuthorizeError(w, r, raw, oauthErr("unauthorized_client", "unknown client", http.StatusBadRequest))
		return
	}

	if !grantAllowed(client.AllowedGrants, "authorization_code") {
		s.redirectAuthorizeError(w, r, req.RedirectURI, req.State,
			oauthErr("unauthorized_client", "client is not allowed to use authorization_code", http.StatusBadRequest))
		return
	}

	redirectURI, ok := resolveAndValidateRedirectURI(client.RedirectURIs, req.RedirectURI)
	if !ok {
		writePlainOAuthError(w, oauthErr("invalid_request", "invalid redirect_uri", http.StatusBadRequest))
		return
	}
	req.RedirectURI = redirectURI

	if req.ResponseType != "code" {
		s.redirectAuthorizeError(w, r, req.RedirectURI, req.State,
			oauthErr("unsupported_response_type", "only response_type=code is supported", http.StatusBadRequest))
		return
	}

	if !scopesAllowed(client.AllowedScopes, req.Scopes) {
		s.redirectAuthorizeError(w, r, req.RedirectURI, req.State,
			oauthErr("invalid_scope", "requested scope is not allowed for this client", http.StatusBadRequest))
		return
	}

	subject, loggedIn, err := s.deps.Authorization.Subject(ctx, r)
	if err != nil {
		s.log.Error("failed to resolve subject", zap.Error(err))
		s.redirectAuthorizeError(w, r, req.RedirectURI, req.State,
			oauthErr("server_error", "failed to resolve subject", http.StatusInternalServerError))
		return
	}
	if !loggedIn || subject == "" {
		if s.cfg.LoginURL == "" {
			writePlainOAuthError(w, oauthErr("server_error", "login_url is not configured", http.StatusInternalServerError))
			return
		}
		loginURL := s.buildLoginRedirect(r)
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}

	existing, err := s.deps.Authorization.ConsentedScopes(ctx, subject, client.ID)
	if err != nil {
		s.log.Error("failed to resolve consented scopes", zap.Error(err))
		s.redirectAuthorizeError(w, r, req.RedirectURI, req.State,
			oauthErr("server_error", "failed to resolve consent", http.StatusInternalServerError))
		return
	}

	missing := diffStrings(req.Scopes, existing)
	if len(missing) > 0 {
		if s.cfg.ConsentURL == "" {
			s.redirectAuthorizeError(w, r, req.RedirectURI, req.State,
				oauthErr("server_error", "consent_url is not configured", http.StatusInternalServerError))
			return
		}
		if s.deps.Interactions == nil {
			s.redirectAuthorizeError(w, r, req.RedirectURI, req.State,
				oauthErr("server_error", "interaction store is not configured", http.StatusInternalServerError))
			return
		}
		itID, err := randomURLSafeString(32)
		if err != nil {
			s.log.Error("failed to generate interaction id", zap.Error(err))
			s.redirectAuthorizeError(w, r, req.RedirectURI, req.State,
				oauthErr("server_error", "failed to start interaction", http.StatusInternalServerError))
			return
		}
		now := time.Now().UTC()
		it := Interaction{
			ID:              itID,
			ClientID:        client.ID,
			RedirectURI:     req.RedirectURI,
			State:           req.State,
			Subject:         subject,
			RequestedScopes: cloneStrings(req.Scopes),
			ExistingScopes:  cloneStrings(existing),
			CreatedAt:       now,
			ExpiresAt:       now.Add(s.interactionTTL()),
			Consumed:        false,
		}
		if err := s.deps.Interactions.CreateInteraction(ctx, it); err != nil {
			s.log.Error("failed to store interaction", zap.Error(err))
			s.redirectAuthorizeError(w, r, req.RedirectURI, req.State,
				oauthErr("server_error", "failed to start interaction", http.StatusInternalServerError))
			return
		}
		consentURL, err := s.buildConsentRedirect(itID)
		if err != nil {
			s.log.Error("failed to build consent redirect", zap.Error(err))
			s.redirectAuthorizeError(w, r, req.RedirectURI, req.State,
				oauthErr("server_error", "failed to start interaction", http.StatusInternalServerError))
			return
		}
		http.Redirect(w, r, consentURL, http.StatusFound)
		return
	}

	if err := s.issueCodeAndRedirect(ctx, w, r, client.ID, req.RedirectURI, subject, req.Scopes, req.State); err != nil {
		oe := asOAuthError(err)
		if oe == nil {
			oe = oauthErr("server_error", "failed to issue authorization code", http.StatusInternalServerError)
		}
		s.redirectAuthorizeError(w, r, req.RedirectURI, req.State, oe)
		return
	}
}

func (s *Server) handleGetInteraction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.setNoStoreHeaders(w)

	if s.deps.Interactions == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "server_error", "error_description": "interaction store is not configured",
		})
		return
	}

	id := mux.Vars(r)["id"]
	if strings.TrimSpace(id) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid_request", "error_description": "missing interaction id",
		})
		return
	}

	it, err := s.deps.Interactions.GetInteraction(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "not_found", "error_description": "interaction not found",
		})
		return
	}

	now := time.Now().UTC()
	if it.Consumed || it.ExpiresAt.UTC().Before(now) {
		writeJSON(w, http.StatusGone, map[string]any{
			"error": "interaction_expired", "error_description": "interaction expired or consumed",
		})
		return
	}

	client, err := s.deps.Clients.GetClient(ctx, it.ClientID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "unauthorized_client", "error_description": "unknown client",
		})
		return
	}

	requested := uniqueStrings(it.RequestedScopes)
	existing := uniqueStrings(it.ExistingScopes)

	resp := map[string]any{
		"interaction_id": id,
		"client": map[string]any{
			"id":          client.ID,
			"name":        client.Display.Name,
			"description": client.Display.Description,
			"logo_url":    client.Display.LogoURL,
			"website_url": client.Display.WebsiteURL,
		},
		"account": map[string]any{
			"subject": it.Subject,
		},
		"requested_scopes": requested,
		"existing_scopes":  existing,
		"missing_scopes":   diffStrings(requested, existing),
		"created_at":       it.CreatedAt.Format(time.RFC3339Nano),
		"expires_at":       it.ExpiresAt.Format(time.RFC3339Nano),
	}

	writeJSON(w, http.StatusOK, resp)
}

type confirmInteractionRequest struct {
	Approve bool     `json:"approve"`
	Scopes  []string `json:"scopes,omitempty"`
}

func (s *Server) handleConfirmInteraction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.setNoStoreHeaders(w)

	if s.deps.Interactions == nil {
		s.log.Error("interaction store is not configured")
		writePlainOAuthError(w, oauthErr("server_error", "interaction store is not configured", http.StatusInternalServerError))
		return
	}

	id := mux.Vars(r)["id"]
	if strings.TrimSpace(id) == "" {
		s.log.Warn("missing interaction id in request")
		writePlainOAuthError(w, oauthErr("invalid_request", "missing interaction id", http.StatusBadRequest))
		return
	}

	var body confirmInteractionRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		s.log.Warn("Failed to decode confirm interaction request body", zap.Error(err))
		writePlainOAuthError(w, oauthErr("invalid_request", "failed to parse json body", http.StatusBadRequest))
		return
	}

	it, err := s.deps.Interactions.ConsumeInteraction(ctx, id)
	if err != nil {
		s.log.Error("Failed to consume interaction", zap.Error(err), zap.String("interaction_id", id))
		writePlainOAuthError(w, oauthErr("invalid_request", "interaction not found", http.StatusBadRequest))
		return
	}

	now := time.Now().UTC()
	if it.Consumed || it.ExpiresAt.UTC().Before(now) {
		s.log.Warn("Interaction expired or already consumed", zap.String("interaction_id", id))
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State,
			oauthErr("access_denied", "interaction expired", http.StatusForbidden))
		return
	}

	subject, ok, err := s.deps.Authorization.Subject(ctx, r)
	if err != nil {
		s.log.Warn("failed to resolve subject", zap.Error(err))
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State,
			oauthErr("server_error", "failed to resolve subject", http.StatusInternalServerError))
		return
	}
	if !ok || subject == "" {
		s.log.Warn("User not logged in", zap.String("interaction_id", id))
		loginURL := s.buildLoginRedirectToConsent(r, id)
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}
	if subject != it.Subject {
		s.log.Warn("Subject mismatch", zap.String("interaction_id", id), zap.String("expected_subject", it.Subject), zap.String("actual_subject", subject))
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State,
			oauthErr("access_denied", "subject mismatch", http.StatusForbidden))
		return
	}

	client, err := s.deps.Clients.GetClient(ctx, it.ClientID)
	if err != nil {
		s.log.Warn("Unknown client", zap.String("interaction_id", id), zap.String("client_id", it.ClientID))
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State,
			oauthErr("unauthorized_client", "unknown client", http.StatusBadRequest))
		return
	}
	if !grantAllowed(client.AllowedGrants, "authorization_code") {
		s.log.Warn("Client not allowed to use authorization_code grant", zap.String("interaction_id", id), zap.String("client_id", it.ClientID))
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State,
			oauthErr("unauthorized_client", "client is not allowed to use authorization_code", http.StatusBadRequest))
		return
	}

	ru, ok := resolveAndValidateRedirectURI(client.RedirectURIs, it.RedirectURI)
	if !ok || ru != it.RedirectURI {
		s.log.Warn("redirect_uri mismatch", zap.String("interaction_id", id), zap.String("client_id", it.ClientID), zap.String("expected_redirect_uri", it.RedirectURI), zap.String("allowed_redirect_uris", strings.Join(client.RedirectURIs, ",")))
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State,
			oauthErr("invalid_request", "redirect_uri mismatch", http.StatusBadRequest))
		return
	}

	if !body.Approve {
		s.log.Info("User denied consent", zap.String("interaction_id", id), zap.String("client_id", it.ClientID), zap.String("subject", subject))
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State,
			oauthErr("access_denied", "user denied consent", http.StatusForbidden))
		return
	}

	approved := uniqueStrings(body.Scopes)
	if len(approved) == 0 {
		approved = uniqueStrings(it.RequestedScopes)
	}

	if !isSubset(approved, it.RequestedScopes) {
		s.log.Warn("Approved scopes are not subset of requested scopes", zap.String("interaction_id", id), zap.String("client_id", it.ClientID), zap.String("subject", subject), zap.Strings("requested_scopes", it.RequestedScopes), zap.Strings("approved_scopes", approved))
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State,
			oauthErr("invalid_scope", "approved scopes must be subset of requested scopes", http.StatusBadRequest))
		return
	}
	if !scopesAllowed(client.AllowedScopes, approved) {
		s.log.Warn("Approved scopes are not allowed for this client", zap.String("interaction_id", id), zap.String("client_id", it.ClientID), zap.String("subject", subject), zap.Strings("approved_scopes", approved), zap.Strings("client_allowed_scopes", client.AllowedScopes))
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State,
			oauthErr("invalid_scope", "approved scope is not allowed for this client", http.StatusBadRequest))
		return
	}

	newGranted := unionStrings(it.ExistingScopes, approved)
	if err := s.deps.Authorization.SaveConsent(ctx, it.Subject, it.ClientID, newGranted); err != nil {
		s.log.Error("Failed to save consent", zap.Error(err), zap.String("interaction_id", id), zap.String("client_id", it.ClientID), zap.String("subject", subject), zap.Strings("approved_scopes", approved))
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State,
			oauthErr("server_error", "failed to save consent", http.StatusInternalServerError))
		return
	}

	if redirectUrl, err := s.issueCodeWithRedirectURL(ctx, client.ID, it.RedirectURI, it.Subject, approved, it.State); err != nil {
		s.log.Error("Failed to issue authorization code", zap.Error(err), zap.String("interaction_id", id), zap.String("client_id", it.ClientID), zap.String("subject", subject), zap.Strings("approved_scopes", approved))
		oe := asOAuthError(err)
		if oe == nil {
			oe = oauthErr("server_error", "failed to issue authorization code", http.StatusInternalServerError)
		}
		s.redirectAuthorizeError(w, r, it.RedirectURI, it.State, oe)
		return
	} else {
		writeJSON(w, http.StatusOK, map[string]any{
			"redirect_to": redirectUrl,
		})
		return
	}
}

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	s.setNoStoreHeaders(w)

	if err := s.parseFormWithLimit(w, r); err != nil {
		return
	}

	grantType := strings.TrimSpace(r.Form.Get("grant_type"))
	switch grantType {
	case "authorization_code":
		s.handleTokenAuthorizationCode(w, r)
	case "refresh_token":
		s.handleTokenRefreshToken(w, r)
	case "client_credentials":
		s.handleTokenClientCredentials(w, r)
	default:
		writeJSONOAuthError(w, oauthErr("unsupported_grant_type", "unsupported grant_type", http.StatusBadRequest))
	}
}

func (s *Server) handleIntrospect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.setNoStoreHeaders(w)

	if err := s.parseFormWithLimit(w, r); err != nil {
		return
	}

	client, err := s.authenticateClientForTokenLikeEndpoints(ctx, r, authClientOpts{
		AllowPublic: false,
	})
	if err != nil {
		writeJSONOAuthError(w, err)
		return
	}

	token := strings.TrimSpace(r.Form.Get("token"))
	if token == "" {
		writeJSONOAuthError(w, oauthErr("invalid_request", "missing token", http.StatusBadRequest))
		return
	}

	hint := strings.TrimSpace(r.Form.Get("token_type_hint"))
	now := time.Now().UTC()

	var (
		active bool
		resp   any
	)
	if hint == "refresh_token" {
		rt, e := s.deps.Tokens.LookupRefreshToken(ctx, token)
		active, resp = introspectRefresh(client.ID, rt, e, now, s.cfg.Issuer)
	} else {
		at, e := s.deps.Tokens.LookupAccessToken(ctx, token)
		active, resp = introspectAccess(client.ID, at, e, now, s.cfg.Issuer)
	}

	if !active && hint != "refresh_token" {
		rt, e := s.deps.Tokens.LookupRefreshToken(ctx, token)
		active, resp = introspectRefresh(client.ID, rt, e, now, s.cfg.Issuer)
	}
	if !active && hint == "refresh_token" {
		at, e := s.deps.Tokens.LookupAccessToken(ctx, token)
		active, resp = introspectAccess(client.ID, at, e, now, s.cfg.Issuer)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRevoke(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.setNoStoreHeaders(w)

	if err := s.parseFormWithLimit(w, r); err != nil {
		return
	}

	_, err := s.authenticateClientForTokenLikeEndpoints(ctx, r, authClientOpts{
		AllowPublic: false,
	})
	if err != nil {
		writeJSONOAuthError(w, err)
		return
	}

	token := strings.TrimSpace(r.Form.Get("token"))
	if token == "" {
		writeJSONOAuthError(w, oauthErr("invalid_request", "missing token", http.StatusBadRequest))
		return
	}

	hint := strings.TrimSpace(r.Form.Get("token_type_hint"))
	switch hint {
	case "access_token":
		_ = s.deps.Tokens.RevokeAccessToken(ctx, token)
	case "refresh_token":
		_ = s.deps.Tokens.RevokeRefreshToken(ctx, token)
	default:
		_ = s.deps.Tokens.RevokeAccessToken(ctx, token)
		_ = s.deps.Tokens.RevokeRefreshToken(ctx, token)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "")
}

func (s *Server) handleTokenAuthorizationCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code := strings.TrimSpace(r.Form.Get("code"))
	if code == "" {
		writeJSONOAuthError(w, oauthErr("invalid_request", "missing code", http.StatusBadRequest))
		return
	}

	redirectURI := strings.TrimSpace(r.Form.Get("redirect_uri"))

	client, err := s.authenticateClientForTokenLikeEndpoints(ctx, r, authClientOpts{
		AllowPublic: s.allowPublicClientsOnTokenEndpoint(),
	})
	if err != nil {
		writeJSONOAuthError(w, err)
		return
	}

	if !grantAllowed(client.AllowedGrants, "authorization_code") {
		writeJSONOAuthError(w, oauthErr("unauthorized_client", "client is not allowed to use authorization_code", http.StatusBadRequest))
		return
	}

	stored, err := s.deps.Codes.Consume(ctx, code)
	if err != nil {
		writeJSONOAuthError(w, oauthErr("invalid_grant", "invalid authorization code", http.StatusBadRequest))
		return
	}

	now := time.Now().UTC()
	if stored.ExpiresAt.UTC().Before(now) {
		writeJSONOAuthError(w, oauthErr("invalid_grant", "authorization code expired", http.StatusBadRequest))
		return
	}

	if stored.ClientID != client.ID {
		writeJSONOAuthError(w, oauthErr("invalid_grant", "authorization code was not issued to this client", http.StatusBadRequest))
		return
	}

	if stored.RedirectURI == "" || redirectURI == "" || stored.RedirectURI != redirectURI {
		writeJSONOAuthError(w, oauthErr("invalid_grant", "redirect_uri mismatch", http.StatusBadRequest))
		return
	}

	access, err := s.deps.Authorization.IssueAccessToken(s.cfg.AccessTokenTTL, client.ID, stored.Subject, stored.Scopes)
	if err != nil {
		s.log.Error("failed to issue access token", zap.Error(err))
		writeJSONOAuthError(w, oauthErr("server_error", "failed to issue access token", http.StatusInternalServerError))
		return
	}
	if err := s.deps.Tokens.SaveAccessToken(ctx, access); err != nil {
		s.log.Error("failed to store access token", zap.Error(err))
		writeJSONOAuthError(w, oauthErr("server_error", "failed to store access token", http.StatusInternalServerError))
		return
	}

	var refresh *RefreshToken
	if s.issueRefreshToken() {
		rt, e := s.issueRefreshTokenValue(now, client.ID, stored.Subject, stored.Scopes, stored.Code)
		if e != nil {
			s.log.Error("failed to issue refresh token", zap.Error(e))
			writeJSONOAuthError(w, oauthErr("server_error", "failed to issue refresh token", http.StatusInternalServerError))
			return
		}
		if e := s.deps.Tokens.SaveRefreshToken(ctx, rt); e != nil {
			s.log.Error("failed to store refresh token", zap.Error(e))
			writeJSONOAuthError(w, oauthErr("server_error", "failed to store refresh token", http.StatusInternalServerError))
			return
		}
		refresh = &rt
	}

	writeTokenResponse(w, access, refresh)
}

func (s *Server) handleTokenRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	refreshStr := strings.TrimSpace(r.Form.Get("refresh_token"))
	if refreshStr == "" {
		writeJSONOAuthError(w, oauthErr("invalid_request", "missing refresh_token", http.StatusBadRequest))
		return
	}

	client, err := s.authenticateClientForTokenLikeEndpoints(ctx, r, authClientOpts{
		AllowPublic: s.allowPublicClientsOnTokenEndpoint(),
	})
	if err != nil {
		writeJSONOAuthError(w, err)
		return
	}

	if !grantAllowed(client.AllowedGrants, "refresh_token") {
		writeJSONOAuthError(w, oauthErr("unauthorized_client", "client is not allowed to use refresh_token", http.StatusBadRequest))
		return
	}

	rt, err := s.deps.Tokens.LookupRefreshToken(ctx, refreshStr)
	if err != nil {
		writeJSONOAuthError(w, oauthErr("invalid_grant", "invalid refresh_token", http.StatusBadRequest))
		return
	}

	now := time.Now().UTC()
	if rt.Revoked || rt.ExpiresAt.UTC().Before(now) {
		writeJSONOAuthError(w, oauthErr("invalid_grant", "refresh_token is expired or revoked", http.StatusBadRequest))
		return
	}
	if rt.ClientID != client.ID {
		writeJSONOAuthError(w, oauthErr("invalid_grant", "refresh_token was not issued to this client", http.StatusBadRequest))
		return
	}

	reqScope := parseScopes(r.Form.Get("scope"))
	scopes := rt.Scopes
	if len(reqScope) > 0 {
		if !isSubset(reqScope, rt.Scopes) {
			writeJSONOAuthError(w, oauthErr("invalid_scope", "requested scope exceeds originally granted scope", http.StatusBadRequest))
			return
		}
		scopes = reqScope
	}

	access, err := s.deps.Authorization.IssueAccessToken(s.cfg.AccessTokenTTL, client.ID, rt.Subject, scopes)
	if err != nil {
		s.log.Error("failed to issue access token", zap.Error(err))
		writeJSONOAuthError(w, oauthErr("server_error", "failed to issue access token", http.StatusInternalServerError))
		return
	}
	if err := s.deps.Tokens.SaveAccessToken(ctx, access); err != nil {
		s.log.Error("failed to store access token", zap.Error(err))
		writeJSONOAuthError(w, oauthErr("server_error", "failed to store access token", http.StatusInternalServerError))
		return
	}

	var newRefresh *RefreshToken
	if s.rotateRefreshTokens() {
		rot, e := s.issueRefreshTokenValue(now, client.ID, rt.Subject, rt.Scopes, rt.AuthorizationCode)
		if e != nil {
			s.log.Error("failed to issue refresh token", zap.Error(e))
			writeJSONOAuthError(w, oauthErr("server_error", "failed to issue refresh token", http.StatusInternalServerError))
			return
		}
		if e := s.deps.Tokens.SaveRefreshToken(ctx, rot); e != nil {
			s.log.Error("failed to store refresh token", zap.Error(e))
			writeJSONOAuthError(w, oauthErr("server_error", "failed to store refresh token", http.StatusInternalServerError))
			return
		}
		newRefresh = &rot

		if s.revokeOldRefreshToken() {
			_ = s.deps.Tokens.RevokeRefreshToken(ctx, rt.Token)
		}
	}

	writeTokenResponse(w, access, newRefresh)
}

func (s *Server) handleTokenClientCredentials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	client, err := s.authenticateClientForTokenLikeEndpoints(ctx, r, authClientOpts{
		AllowPublic: false,
	})
	if err != nil {
		writeJSONOAuthError(w, err)
		return
	}

	if !grantAllowed(client.AllowedGrants, "client_credentials") {
		writeJSONOAuthError(w, oauthErr("unauthorized_client", "client is not allowed to use client_credentials", http.StatusBadRequest))
		return
	}

	reqScopes := parseScopes(r.Form.Get("scope"))
	if !scopesAllowed(client.AllowedScopes, reqScopes) {
		writeJSONOAuthError(w, oauthErr("invalid_scope", "requested scope is not allowed for this client", http.StatusBadRequest))
		return
	}

	access, err := s.deps.Authorization.IssueAccessToken(s.cfg.AccessTokenTTL, client.ID, client.ID, reqScopes)
	if err != nil {
		s.log.Error("failed to issue access token", zap.Error(err))
		writeJSONOAuthError(w, oauthErr("server_error", "failed to issue access token", http.StatusInternalServerError))
		return
	}
	if err := s.deps.Tokens.SaveAccessToken(ctx, access); err != nil {
		s.log.Error("failed to store access token", zap.Error(err))
		writeJSONOAuthError(w, oauthErr("server_error", "failed to store access token", http.StatusInternalServerError))
		return
	}

	writeTokenResponse(w, access, nil)
}

func parseAuthorizeRequest(q url.Values) (AuthorizeRequest, url.Values, error) {
	raw := cloneValues(q)

	responseType := strings.TrimSpace(q.Get("response_type"))
	clientID := strings.TrimSpace(q.Get("client_id"))
	redirectURI := strings.TrimSpace(q.Get("redirect_uri"))
	state := q.Get("state")

	scopeStr := q.Get("scope")
	scopes := parseScopes(scopeStr)

	if responseType == "" {
		return AuthorizeRequest{}, raw, errors.New("missing response_type")
	}
	if clientID == "" {
		return AuthorizeRequest{}, raw, errors.New("missing client_id")
	}

	return AuthorizeRequest{
		ResponseType: responseType,
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		Scopes:       scopes,
		State:        state,
	}, raw, nil
}

func resolveAndValidateRedirectURI(registered []string, requested string) (string, bool) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		if len(registered) == 1 {
			return registered[0], true
		}
		return "", false
	}
	for _, ru := range registered {
		if ru == requested {
			return requested, true
		}
	}
	return "", false
}

func parseScopes(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Fields(s)
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func scopesAllowed(allowed []string, requested []string) bool {
	if len(requested) == 0 {
		return true
	}
	if len(allowed) == 0 {
		return true
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		allowedSet[a] = struct{}{}
	}
	for _, r := range requested {
		if _, ok := allowedSet[r]; !ok {
			return false
		}
	}
	return true
}

func isSubset(a, b []string) bool {
	if len(a) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(b))
	for _, x := range b {
		set[x] = struct{}{}
	}
	for _, x := range a {
		if _, ok := set[x]; !ok {
			return false
		}
	}
	return true
}

func grantAllowed(allowed []string, grant string) bool {
	if len(allowed) == 0 {
		return false
	}
	for _, g := range allowed {
		if g == grant {
			return true
		}
	}
	return false
}

type authClientOpts struct {
	AllowPublic bool
}

func (s *Server) authenticateClientForTokenLikeEndpoints(ctx context.Context, r *http.Request, opts authClientOpts) (Client, error) {
	idFromBasic, secretFromBasic, hasBasic := parseBasicAuth(r)

	idFromBody := strings.TrimSpace(r.Form.Get("client_id"))
	secretFromBody := r.Form.Get("client_secret")

	clientID := ""
	clientSecret := ""

	if hasBasic {
		clientID = idFromBasic
		clientSecret = secretFromBasic
		if idFromBody != "" && idFromBody != clientID {
			return Client{}, oauthErr("invalid_client", "client_id mismatch between Authorization header and body", http.StatusUnauthorized)
		}
	} else {
		clientID = idFromBody
		clientSecret = secretFromBody
	}

	if clientID == "" {
		return Client{}, oauthErr("invalid_client", "missing client authentication", http.StatusUnauthorized)
	}

	client, err := s.deps.Clients.GetClient(ctx, clientID)
	if err != nil {
		return Client{}, oauthErr("invalid_client", "unknown client", http.StatusUnauthorized)
	}

	if client.Public || client.Secret == "" {
		if !opts.AllowPublic {
			return Client{}, oauthErr("invalid_client", "public clients are not allowed on this endpoint", http.StatusUnauthorized)
		}
		if clientSecret != "" {
			ok, e := s.deps.Clients.ValidateClientSecret(ctx, clientID, clientSecret)
			if e != nil {
				s.log.Warn("client secret validation error", zap.Error(e), zap.String("client_id", clientID))
				return Client{}, oauthErr("server_error", "client validation error", http.StatusInternalServerError)
			}
			if !ok {
				return Client{}, oauthErr("invalid_client", "invalid client secret", http.StatusUnauthorized)
			}
		}
		return client, nil
	}

	if clientSecret == "" {
		return Client{}, oauthErr("invalid_client", "missing client secret", http.StatusUnauthorized)
	}

	ok, e := s.deps.Clients.ValidateClientSecret(ctx, clientID, clientSecret)
	if e != nil {
		s.log.Warn("client secret validation error", zap.Error(e), zap.String("client_id", clientID))
		return Client{}, oauthErr("server_error", "client validation error", http.StatusInternalServerError)
	}
	if !ok {
		return Client{}, oauthErr("invalid_client", "invalid client secret", http.StatusUnauthorized)
	}

	return client, nil
}

func parseBasicAuth(r *http.Request) (clientID, clientSecret string, ok bool) {
	id, secret, ok := r.BasicAuth()
	if !ok {
		return "", "", false
	}
	return id, secret, true
}

func (s *Server) issueRefreshTokenValue(now time.Time, clientID, subject string, scopes []string, authCode string) (RefreshToken, error) {
	token, err := randomURLSafeString(48)
	if err != nil {
		return RefreshToken{}, err
	}
	return RefreshToken{
		Token:             token,
		ClientID:          clientID,
		Subject:           subject,
		Scopes:            cloneStrings(scopes),
		IssuedAt:          now,
		ExpiresAt:         now.Add(s.cfg.RefreshTokenTTL),
		Revoked:           false,
		AuthorizationCode: authCode,
	}, nil
}

func writeTokenResponse(w http.ResponseWriter, access AccessToken, refresh *RefreshToken) {
	resp := map[string]any{
		"access_token": access.Token,
		"token_type":   "Bearer",
		"expires_in":   int64(time.Until(access.ExpiresAt).Seconds()),
	}
	if refresh != nil {
		resp["refresh_token"] = refresh.Token
	}
	if len(access.Scopes) > 0 {
		resp["scope"] = strings.Join(access.Scopes, " ")
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) setNoStoreHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}

func writeJSONOAuthError(w http.ResponseWriter, err error) {
	oe := asOAuthError(err)
	if oe == nil {
		oe = oauthErr("server_error", "internal error", http.StatusInternalServerError)
	}
	if oe.HTTPStatus == 0 {
		oe.HTTPStatus = http.StatusBadRequest
	}

	if oe.Code == "invalid_client" && oe.HTTPStatus == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", `Basic realm="oauth2"`)
	}

	writeJSON(w, oe.HTTPStatus, map[string]any{
		"error":             oe.Code,
		"error_description": oe.Description,
	})
}

func writePlainOAuthError(w http.ResponseWriter, err error) {
	oe := asOAuthError(err)
	if oe == nil {
		oe = oauthErr("server_error", "internal error", http.StatusInternalServerError)
	}
	if oe.HTTPStatus == 0 {
		oe.HTTPStatus = http.StatusBadRequest
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(oe.HTTPStatus)
	_, _ = io.WriteString(w, fmt.Sprintf("%s: %s", oe.Code, oe.Description))
}

func (s *Server) redirectAuthorizeError(w http.ResponseWriter, r *http.Request, redirectURI, state string, err error) {
	oe := asOAuthError(err)
	if oe == nil {
		oe = oauthErr("server_error", "internal error", http.StatusInternalServerError)
	}

	u, parseErr := url.Parse(redirectURI)
	if parseErr != nil {
		writePlainOAuthError(w, oauthErr("server_error", "failed to parse redirect_uri", http.StatusInternalServerError))
		return
	}

	q := u.Query()
	q.Set("error", oe.Code)
	if oe.Description != "" {
		q.Set("error_description", oe.Description)
	}
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (s *Server) writeAuthorizeError(w http.ResponseWriter, r *http.Request, raw url.Values, err error) {
	clientID := strings.TrimSpace(raw.Get("client_id"))
	redirectURI := strings.TrimSpace(raw.Get("redirect_uri"))
	state := raw.Get("state")

	if clientID == "" || redirectURI == "" {
		writePlainOAuthError(w, err)
		return
	}

	client, e := s.deps.Clients.GetClient(r.Context(), clientID)
	if e != nil {
		writePlainOAuthError(w, err)
		return
	}

	ru, ok := resolveAndValidateRedirectURI(client.RedirectURIs, redirectURI)
	if !ok {
		writePlainOAuthError(w, err)
		return
	}
	s.redirectAuthorizeError(w, r, ru, state, err)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}

func introspectAccess(requestingClientID string, at AccessToken, err error, now time.Time, issuer string) (active bool, resp any) {
	if err != nil {
		return false, map[string]any{"active": false}
	}
	if at.ClientID != requestingClientID {
		return false, map[string]any{"active": false}
	}
	if at.Revoked || at.ExpiresAt.UTC().Before(now) {
		return false, map[string]any{"active": false}
	}

	m := map[string]any{
		"active":     true,
		"client_id":  at.ClientID,
		"sub":        at.Subject,
		"token_type": "access_token",
		"exp":        at.ExpiresAt.Unix(),
		"iat":        at.IssuedAt.Unix(),
	}
	if len(at.Scopes) > 0 {
		m["scope"] = strings.Join(at.Scopes, " ")
	}
	if issuer != "" {
		m["iss"] = issuer
	}
	return true, m
}

func introspectRefresh(requestingClientID string, rt RefreshToken, err error, now time.Time, issuer string) (active bool, resp any) {
	if err != nil {
		return false, map[string]any{"active": false}
	}
	if rt.ClientID != requestingClientID {
		return false, map[string]any{"active": false}
	}
	if rt.Revoked || rt.ExpiresAt.UTC().Before(now) {
		return false, map[string]any{"active": false}
	}

	m := map[string]any{
		"active":     true,
		"client_id":  rt.ClientID,
		"sub":        rt.Subject,
		"token_type": "refresh_token",
		"exp":        rt.ExpiresAt.Unix(),
		"iat":        rt.IssuedAt.Unix(),
	}
	if len(rt.Scopes) > 0 {
		m["scope"] = strings.Join(rt.Scopes, " ")
	}
	if issuer != "" {
		m["iss"] = issuer
	}
	return true, m
}

func (s *Server) parseFormWithLimit(w http.ResponseWriter, r *http.Request) error {
	limit := s.cfg.MaxBodyBytes
	if limit <= 0 {
		limit = 1 << 20
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)

	ct := r.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		writeJSONOAuthError(w, oauthErr("invalid_request", "content-type must be application/x-www-form-urlencoded", http.StatusBadRequest))
		return errors.New("invalid content-type")
	}

	if err := r.ParseForm(); err != nil {
		writeJSONOAuthError(w, oauthErr("invalid_request", "failed to parse form", http.StatusBadRequest))
		return err
	}
	return nil
}

func randomURLSafeString(nBytes int) (string, error) {
	if nBytes <= 0 {
		nBytes = 32
	}
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func cloneStrings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneValues(v url.Values) url.Values {
	out := make(url.Values, len(v))
	for k, arr := range v {
		cp := make([]string, len(arr))
		copy(cp, arr)
		out[k] = cp
	}
	return out
}

func applyDefaults(cfg *Config) {
	if cfg.AuthorizePath == "" {
		cfg.AuthorizePath = "/authorize"
	}
	if cfg.TokenPath == "" {
		cfg.TokenPath = "/token"
	}
	if cfg.IntrospectPath == "" {
		cfg.IntrospectPath = "/introspect"
	}
	if cfg.RevokePath == "" {
		cfg.RevokePath = "/revoke"
	}

	if cfg.AuthorizationCodeTTL <= 0 {
		cfg.AuthorizationCodeTTL = 5 * time.Minute
	}
	if cfg.AccessTokenTTL <= 0 {
		cfg.AccessTokenTTL = 15 * time.Minute
	}
	if cfg.RefreshTokenTTL <= 0 {
		cfg.RefreshTokenTTL = 30 * 24 * time.Hour
	}

	if cfg.MaxBodyBytes == 0 {
		cfg.MaxBodyBytes = 1 << 20
	}

	if cfg.IssueRefreshToken == nil {
		v := true
		cfg.IssueRefreshToken = &v
	}
	if cfg.RotateRefreshTokens == nil {
		v := true
		cfg.RotateRefreshTokens = &v
	}
	if cfg.RevokeOldRefreshToken == nil {
		v := true
		cfg.RevokeOldRefreshToken = &v
	}
	if cfg.AllowPublicClientsOnTokenEndpoint == nil {
		v := false
		cfg.AllowPublicClientsOnTokenEndpoint = &v
	}

	if cfg.LoginReturnParam == "" {
		cfg.LoginReturnParam = "return_to"
	}
	if cfg.ConsentInteractionParam == "" {
		cfg.ConsentInteractionParam = "interaction_id"
	}
	if cfg.InteractionPath == "" {
		cfg.InteractionPath = "/interaction"
	}
	if cfg.InteractionConfirmPath == "" {
		cfg.InteractionConfirmPath = "/confirm"
	}
	if cfg.InteractionTTL <= 0 {
		cfg.InteractionTTL = 5 * time.Minute
	}
}

func (s *Server) issueRefreshToken() bool {
	if s.cfg.IssueRefreshToken == nil {
		return true
	}
	return *s.cfg.IssueRefreshToken
}
func (s *Server) rotateRefreshTokens() bool {
	if s.cfg.RotateRefreshTokens == nil {
		return true
	}
	return *s.cfg.RotateRefreshTokens
}
func (s *Server) revokeOldRefreshToken() bool {
	if s.cfg.RevokeOldRefreshToken == nil {
		return true
	}
	return *s.cfg.RevokeOldRefreshToken
}
func (s *Server) allowPublicClientsOnTokenEndpoint() bool {
	if s.cfg.AllowPublicClientsOnTokenEndpoint == nil {
		return false
	}
	return *s.cfg.AllowPublicClientsOnTokenEndpoint
}

func (s *Server) issueCodeAndRedirect(ctx context.Context, w http.ResponseWriter, r *http.Request,
	clientID, redirectURI, subject string, scopes []string, state string,
) error {
	codeStr, err := randomURLSafeString(32)
	if err != nil {
		s.log.Error("failed to generate authorization code", zap.Error(err))
		return oauthErr("server_error", "failed to generate authorization code", http.StatusInternalServerError)
	}

	now := time.Now().UTC()
	ac := AuthorizationCode{
		Code:        codeStr,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Subject:     subject,
		Scopes:      cloneStrings(scopes),
		IssuedAt:    now,
		ExpiresAt:   now.Add(s.cfg.AuthorizationCodeTTL),
		Consumed:    false,
	}

	if err := s.deps.Codes.Create(ctx, ac); err != nil {
		s.log.Error("failed to store authorization code", zap.Error(err))
		return oauthErr("server_error", "failed to store authorization code", http.StatusInternalServerError)
	}

	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("code", codeStr)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
	return nil
}

func (s *Server) issueCodeWithRedirectURL(ctx context.Context, clientID, redirectURI, subject string, scopes []string, state string) (string, error) {
	codeStr, err := randomURLSafeString(32)
	if err != nil {
		s.log.Error("failed to generate authorization code", zap.Error(err))
		return "", oauthErr("server_error", "failed to generate authorization code", http.StatusInternalServerError)
	}

	now := time.Now().UTC()
	ac := AuthorizationCode{
		Code:        codeStr,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Subject:     subject,
		Scopes:      cloneStrings(scopes),
		IssuedAt:    now,
		ExpiresAt:   now.Add(s.cfg.AuthorizationCodeTTL),
		Consumed:    false,
	}

	if err := s.deps.Codes.Create(ctx, ac); err != nil {
		s.log.Error("failed to store authorization code", zap.Error(err))
		return "", oauthErr("server_error", "failed to store authorization code", http.StatusInternalServerError)
	}

	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("code", codeStr)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (s *Server) interactionTTL() time.Duration {
	if s.cfg.InteractionTTL > 0 {
		return s.cfg.InteractionTTL
	}
	return 5 * time.Minute
}

func (s *Server) buildLoginRedirect(r *http.Request) string {
	returnParam := s.cfg.LoginReturnParam
	if strings.TrimSpace(returnParam) == "" {
		returnParam = "return_to"
	}
	returnTo := externalRequestURL(r)
	u, _ := url.Parse(s.cfg.LoginURL)
	q := u.Query()
	q.Set(returnParam, returnTo)
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *Server) buildLoginRedirectToConsent(_ *http.Request, interactionID string) string {
	returnParam := s.cfg.LoginReturnParam
	if strings.TrimSpace(returnParam) == "" {
		returnParam = "return_to"
	}
	consentURL, _ := url.Parse(s.cfg.ConsentURL)
	consentParam := s.cfg.ConsentInteractionParam
	if strings.TrimSpace(consentParam) == "" {
		consentParam = "interaction_id"
	}
	consentQ := url.Values{}
	consentQ.Set(consentParam, interactionID)
	consentURL.RawQuery = consentQ.Encode()
	u, _ := url.Parse(s.cfg.LoginURL)
	q := u.Query()
	q.Set(returnParam, consentURL.String())
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *Server) buildConsentRedirect(interactionID string) (string, error) {
	u, err := url.Parse(s.cfg.ConsentURL)
	if err != nil {
		return "", err
	}
	param := s.cfg.ConsentInteractionParam
	if strings.TrimSpace(param) == "" {
		param = "interaction_id"
	}
	q := url.Values{}
	q.Set(param, interactionID)
	u.RawQuery = q.Encode()
	u.Fragment = ""
	return u.String(), nil
}

func externalRequestURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xfProto := r.Header.Get("X-Forwarded-Proto"); xfProto != "" {
		parts := strings.Split(xfProto, ",")
		scheme = strings.TrimSpace(parts[0])
	}
	host := r.Host
	if xfHost := r.Header.Get("X-Forwarded-Host"); xfHost != "" {
		parts := strings.Split(xfHost, ",")
		host = strings.TrimSpace(parts[0])
	}
	uri := r.URL.RequestURI()
	return scheme + "://" + host + uri
}

func uniqueStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func unionStrings(a, b []string) []string {
	set := make(map[string]struct{}, len(a)+len(b))
	for _, x := range a {
		x = strings.TrimSpace(x)
		if x == "" {
			continue
		}
		set[x] = struct{}{}
	}
	for _, x := range b {
		x = strings.TrimSpace(x)
		if x == "" {
			continue
		}
		set[x] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

func diffStrings(requested, existing []string) []string {
	req := uniqueStrings(requested)
	ex := make(map[string]struct{}, len(existing))
	for _, x := range existing {
		x = strings.TrimSpace(x)
		if x == "" {
			continue
		}
		ex[x] = struct{}{}
	}
	out := make([]string, 0, len(req))
	for _, s := range req {
		if _, ok := ex[s]; !ok {
			out = append(out, s)
		}
	}
	return out
}
