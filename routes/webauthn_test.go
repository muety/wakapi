package routes

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/descope/virtualwebauthn"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/securecookie"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type WebAuthnRegisterRequest struct {
	Action          string `json:"action"`
	AttestationJSON string `json:"attestation_json"`
}

var (
	rp = virtualwebauthn.RelyingParty{
		ID:     "example.com",
		Name:   "Wakapi",
		Origin: "https://example.com",
	}
	settingsHandler *SettingsHandler
	loginHandler    *LoginHandler
	userService     *mocks.UserServiceMock
	webauthnService *mocks.WebAuthnServiceMock
	lastCookies     []*http.Cookie
)

func TestWebauthn_RegisterAndAuth(t *testing.T) {
	createAndLoadConfig()

	userNonLocal := createUser()
	userNonLocal.AuthType = "not-local-here"
	userA := createUser()
	userB := createUser()

	userService = new(mocks.UserServiceMock)
	userService.On("GetUserById", userA.ID).Return(userA, nil)
	userService.On("GetUserById", userB.ID).Return(userB, nil)
	userService.On("GetUserById", userNonLocal.ID).Return(userNonLocal, nil)
	userService.On("GetUserByWebAuthnID", userA.WebauthnID).Return(userA, nil)
	userService.On("GetUserByWebAuthnID", userB.WebauthnID).Return(userB, nil)
	userService.On("Update", mock.MatchedBy(func(user *models.User) bool { return user.ID == userA.ID })).Return(userA, nil)
	userService.On("Update", mock.MatchedBy(func(user *models.User) bool { return user.ID == userB.ID })).Return(userB, nil)
	userService.On("Count").Return(0, nil)

	webauthnService = new(mocks.WebAuthnServiceMock)
	webauthnService.On("LoadCredentialIntoUser", userA).Return(nil)
	webauthnService.On("LoadCredentialIntoUser", userB).Return(nil)
	webauthnService.On("CreateCredential", mock.Anything, userA, "authenticator 1").Return(&models.WebAuthnCredential{}, nil)
	webauthnService.On("CreateCredential", mock.Anything, userB, "authenticator 2").Return(&models.WebAuthnCredential{}, nil)
	webauthnService.On("UpdateCredential", mock.Anything).Return(nil)
	webauthnService.On("GetCredentialByUserAndName", mock.Anything, mock.Anything).Return(&models.WebAuthnCredential{}, nil)
	webauthnService.On("DeleteCredential", mock.Anything).Return(nil)

	settingsHandler = NewSettingsHandler(userService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, webauthnService)
	loginHandler = NewLoginHandler(userService, nil, nil, webauthnService)

	authenticator1 := virtualwebauthn.NewAuthenticator()
	authenticator2 := virtualwebauthn.NewAuthenticator()
	credential1 := virtualwebauthn.Credential{}
	credential2 := virtualwebauthn.Credential{}

	t.Run("register denied for non-local users", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/settings/webauthn/options", nil)
		req = req.WithContext(context.WithValue(req.Context(), config.KeySharedData, config.NewSharedData()))
		routeutils.SetPrincipal(req, userNonLocal)
		settingsHandler.GetWebAuthnOptions(rec, req)

		res := rec.Result()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("register 1 user 1", func(t *testing.T) {
		registerAuthenticator(t, userA, "authenticator 1", &authenticator1, &credential1)
	})
	t.Run("login 1 user 1", func(t *testing.T) {
		tryLogin(t, &authenticator1, &credential1, userA, true)
	})
	t.Run("register 2 user 2", func(t *testing.T) {
		registerAuthenticator(t, userB, "authenticator 2", &authenticator2, &credential2)
	})
	t.Run("login 1 user 1 still works", func(t *testing.T) {
		tryLogin(t, &authenticator1, &credential1, userA, true)
	})
	t.Run("login 1 user 2 fails", func(t *testing.T) {
		tryLogin(t, &authenticator1, &credential1, userB, false)
	})
	t.Run("login 2 user 2 works", func(t *testing.T) {
		tryLogin(t, &authenticator2, &credential2, userB, true)
	})
	t.Run("login 2 user 1 fails", func(t *testing.T) {
		tryLogin(t, &authenticator2, &credential2, userA, false)
	})
	t.Run("delete authenticator 2", func(t *testing.T) {
		deleteAuthenticator(t, userB, "authenticator 2")
	})
	t.Run("login 2 user 2 fails after delete", func(t *testing.T) {
		// we need to remove the credential from the mock user's list as well
		userB.Credentials = []*models.WebAuthnCredential{}
		tryLogin(t, &authenticator2, &credential2, userB, false)
	})
}

func getRegisterOption(t *testing.T, user *models.User) string {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings/webauthn/options", nil)
	req = req.WithContext(context.WithValue(req.Context(), config.KeySharedData, config.NewSharedData()))
	routeutils.SetPrincipal(req, user)
	settingsHandler.GetWebAuthnOptions(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	lastCookies = res.Cookies()
	return string(body)
}

func registerAuthenticator(t *testing.T, user *models.User, name string, authenticator *virtualwebauthn.Authenticator, credential *virtualwebauthn.Credential) {
	optionsJSON := getRegisterOption(t, user)
	parsedAttestationOptions, err := virtualwebauthn.ParseAttestationOptions(optionsJSON)
	assert.NoError(t, err)

	*credential = virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)
	attestationResponse := virtualwebauthn.CreateAttestationResponse(rp, *authenticator, *credential, *parsedAttestationOptions)
	authenticator.AddCredential(*credential)

	reqBody := url.Values{
		"action":             {"webauthn_add"},
		"credential_json":    {attestationResponse},
		"authenticator_name": {name},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), config.KeySharedData, config.NewSharedData()))
	routeutils.SetPrincipal(req, user)
	for _, c := range lastCookies {
		req.AddCookie(c)
	}

	res := settingsHandler.actionWebAuthnAdd(rec, req)
	assert.Equal(t, http.StatusOK, res.code)

	// add to user for server-side validation during login (since we use mocks and don't have a real DB)
	user.Credentials = append(user.Credentials, &models.WebAuthnCredential{
		ID:        credential.ID,
		PublicKey: credential.Key.AttestationData(),
		Name:      name,
	})
}

func deleteAuthenticator(t *testing.T, user *models.User, name string) {
	reqBody := url.Values{
		"action":          {"webauthn_delete"},
		"credential_name": {name},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), config.KeySharedData, config.NewSharedData()))
	routeutils.SetPrincipal(req, user)

	res := settingsHandler.actionWebAuthnDelete(rec, req)
	assert.Equal(t, http.StatusOK, res.code)
}

func getLoginOption(t *testing.T) string {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/webauthn/options", nil)
	loginHandler.GetWebAuthnOptions(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	lastCookies = res.Cookies()
	return string(body)
}

func tryLogin(t *testing.T, authenticator *virtualwebauthn.Authenticator, credential *virtualwebauthn.Credential, expectedUser *models.User, shouldLogin bool) {
	optionsJSON := getLoginOption(t)
	parsedAssertionOptions, err := virtualwebauthn.ParseAssertionOptions(optionsJSON)
	assert.NoError(t, err)

	authenticator.Options.UserHandle = []byte(expectedUser.WebauthnID) // in a real scenario, the client would store the user handle and send it back during login, but since we don't have a real client here, we need to set it manually for the assertion to be created correctly
	assertionResponse := virtualwebauthn.CreateAssertionResponse(rp, *authenticator, *credential, *parsedAssertionOptions)

	reqBody := url.Values{
		"assertion_json": {assertionResponse},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webauthn/login", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range lastCookies {
		req.AddCookie(c)
	}

	loginHandler.PostLoginWebAuthn(rec, req)

	res := rec.Result()
	if shouldLogin {
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Contains(t, res.Header.Get("Location"), "/summary")
	} else {
		assert.NotEqual(t, http.StatusFound, res.StatusCode)
	}
}

func createUser() *models.User {
	return &models.User{
		ID:          uuid.Must(uuid.NewV4()).String(),
		WebauthnID:  uuid.Must(uuid.NewV4()).String(),
		AuthType:    "local",
		Credentials: []*models.WebAuthnCredential{},
	}
}

func createAndLoadConfig() {
	cfg := config.Empty()
	hashKey := securecookie.GenerateRandomKey(64)
	blockKey := securecookie.GenerateRandomKey(32)
	sessionKey := securecookie.GenerateRandomKey(32)
	cfg.Security.SecureCookie = securecookie.New(hashKey, blockKey)
	cfg.Security.SessionKey = sessionKey
	cfg.Server.PublicUrl = "https://example.com"
	config.Set(cfg)
	config.InitWebAuthn(cfg)
	config.ResetSessionStore()

	// Initialize templates to avoid panic during failed logins
	templates = make(map[string]*template.Template)
	templates[config.LoginTemplate] = template.Must(template.New("login").Parse("{{.}}"))
}
