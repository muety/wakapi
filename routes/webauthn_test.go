package routes

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/descope/virtualwebauthn"
	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/securecookie"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type WebAuthnTestSuite struct {
	suite.Suite
	Router                 *chi.Mux
	RP                     virtualwebauthn.RelyingParty
	SettingsHandler        *SettingsHandler
	LoginHandler           *LoginHandler
	AliasService           *mocks.AliasServiceMock
	UserService            *mocks.UserServiceMock
	WebauthnService        *mocks.WebAuthnServiceMock
	ProjectLabelService    *mocks.ProjectLabelServiceMock
	ApiKeyService          *mocks.MockApiKeyService
	HeartbeatService       *mocks.HeartbeatServiceMock
	LanguageMappingService *mocks.LanguageMappingServiceMock
	UserNonLocal           *models.User
	UserA                  *models.User
	UserB                  *models.User
	Authenticator1         virtualwebauthn.Authenticator
	Authenticator2         virtualwebauthn.Authenticator
	Credential1            *virtualwebauthn.Credential
	Credential2            *virtualwebauthn.Credential
}

func TestWebauthn(t *testing.T) {
	suite.Run(t, new(WebAuthnTestSuite))
}

func (suite *WebAuthnTestSuite) SetupSuite() {
	suite.RP = virtualwebauthn.RelyingParty{
		ID:     "example.com",
		Name:   "Wakapi",
		Origin: "https://example.com",
	}
}

func (suite *WebAuthnTestSuite) BeforeTest(suiteName, testName string) {
	createAndLoadConfig()
	suite.AliasService = new(mocks.AliasServiceMock)
	suite.UserService = new(mocks.UserServiceMock)
	suite.WebauthnService = new(mocks.WebAuthnServiceMock)
	suite.HeartbeatService = new(mocks.HeartbeatServiceMock)
	suite.ProjectLabelService = new(mocks.ProjectLabelServiceMock)
	suite.ApiKeyService = new(mocks.MockApiKeyService)
	suite.LanguageMappingService = new(mocks.LanguageMappingServiceMock)
	suite.SettingsHandler = NewSettingsHandler(suite.UserService, suite.HeartbeatService, nil, nil, suite.AliasService, nil, suite.LanguageMappingService, suite.ProjectLabelService, nil, nil, suite.ApiKeyService, suite.WebauthnService)
	suite.LoginHandler = NewLoginHandler(suite.UserService, nil, nil, suite.WebauthnService)
	Init() // load templates

	suite.mockSettingsViewDefaults()

	suite.UserNonLocal = createUser()
	suite.UserNonLocal.AuthType = "not-local-here"
	suite.UserA = createUser()
	suite.UserB = createUser()
	suite.Authenticator1 = virtualwebauthn.NewAuthenticator()
	suite.Authenticator2 = virtualwebauthn.NewAuthenticator()
	suite.Credential1 = &virtualwebauthn.Credential{}
	suite.Credential2 = &virtualwebauthn.Credential{}

	suite.Router = chi.NewRouter()
	suite.Router.Use(middlewares.NewSharedDataMiddleware())
	suite.SettingsHandler.RegisterRoutes(suite.Router)
	suite.LoginHandler.RegisterRoutes(suite.Router)
}

func (suite *WebAuthnTestSuite) mockSettingsViewDefaults() {
	suite.LanguageMappingService.On("GetByUser", mock.Anything).Return([]*models.LanguageMapping{}, nil).Maybe()
	suite.AliasService.On("GetByUser", mock.Anything).Return([]*models.Alias{}, nil).Maybe()
	suite.AliasService.On("GetByUserAndType", mock.Anything, mock.Anything).Return([]*models.Alias{}, nil).Maybe()
	suite.ProjectLabelService.On("GetByUserGroupedInverted", mock.Anything).Return(map[string][]*models.ProjectLabel{}, nil).Maybe()
	suite.HeartbeatService.On("GetEntitySetByUser", mock.Anything, mock.Anything).Return([]string{}, nil).Maybe()
	suite.HeartbeatService.On("GetFirstByUser", mock.Anything).Return(time.Time{}, nil).Maybe()
	suite.ApiKeyService.On("GetByUser", mock.Anything).Return([]*models.ApiKey{}, nil).Maybe()
	suite.WebauthnService.On("LoadCredentialIntoUser", mock.Anything).Return(nil).Maybe()
}

func (suite *WebAuthnTestSuite) loginAsUser(user *models.User) []*http.Cookie {
	suite.UserService.On("GetUserById", user.ID).Return(user, nil)
	suite.UserService.On("Update", mock.Anything).Return(user, nil)
	return suite.getLoginCookies(user.ID, user.ID+"_password")
}

func (suite *WebAuthnTestSuite) mockCreateCredential(user *models.User, name string, credential *virtualwebauthn.Credential) {
	createdCredential := &models.WebAuthnCredential{
		Name:   name,
		UserID: user.ID,
	}

	suite.WebauthnService.
		On("CreateCredential", mock.Anything, user, name).
		Run(func(args mock.Arguments) {
			c := args.Get(0).(*webauthn.Credential)
			u := args.Get(1).(*models.User)

			createdCredential.ID = c.ID
			createdCredential.PublicKey = c.PublicKey
			createdCredential.AttestationType = c.AttestationType
			createdCredential.Transport = c.Transport
			createdCredential.Flags = models.CredentialFlags(c.Flags)
			createdCredential.Authenticator = c.Authenticator
			createdCredential.Attestation = c.Attestation

			u.Credentials = append(u.Credentials, createdCredential)
		}).
		Return(createdCredential, nil)
}

func (suite *WebAuthnTestSuite) mockDeleteCredential(user *models.User, name string) {
	suite.WebauthnService.On("GetCredentialByUserAndName", user, name).Return(&models.WebAuthnCredential{Name: name}, nil)
	suite.WebauthnService.
		On("DeleteCredential", mock.Anything).
		Run(func(args mock.Arguments) {
			toDelete := args.Get(0).(*models.WebAuthnCredential)
			filtered := make([]*models.WebAuthnCredential, 0, len(user.Credentials))
			for _, c := range user.Credentials {
				if c.Name != toDelete.Name {
					filtered = append(filtered, c)
				}
			}
			user.Credentials = filtered
		}).
		Return(nil)
}

func (suite *WebAuthnTestSuite) TestWebauthn_RegisterDeniedForNonLocalUsers() {
	suite.UserService.On("GetUserById", suite.UserNonLocal.ID).Return(suite.UserNonLocal, nil)
	suite.UserService.On("Update", mock.AnythingOfType("*models.User")).Return(suite.UserNonLocal, nil)

	cookies := suite.getLoginCookies(suite.UserNonLocal.ID, suite.UserNonLocal.ID+"_password")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings/webauthn/options", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	suite.Router.ServeHTTP(rec, req)

	res := rec.Result()
	body, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	suite.UserService.AssertExpectations(suite.T())
	suite.WebauthnService.AssertExpectations(suite.T())

	suite.Contains(string(body), "webauthn is only available for local users")
	suite.Equal(http.StatusBadRequest, res.StatusCode)
}

func (suite *WebAuthnTestSuite) TestWebauthn_RegisterAndLogin() {
	cookieUserA := suite.loginAsUser(suite.UserA)

	suite.WebauthnService.On("LoadCredentialIntoUser", suite.UserA).Return(nil)
	suite.mockCreateCredential(suite.UserA, "authenticator 1", suite.Credential1)
	suite.UserService.On("Update", mock.Anything).Return(suite.UserA, nil)
	suite.registerAuthenticator(suite.UserA, "authenticator 1", &suite.Authenticator1, suite.Credential1, cookieUserA)

	suite.UserService.On("GetUserByWebAuthnID", suite.UserA.WebauthnID).Return(suite.UserA, nil)
	suite.WebauthnService.On("UpdateCredential", mock.Anything).Return(nil)
	suite.tryLogin(&suite.Authenticator1, suite.Credential1, suite.UserA, true)
	suite.UserService.AssertExpectations(suite.T())
	suite.WebauthnService.AssertExpectations(suite.T())
}

func (suite *WebAuthnTestSuite) TestWebauthn_RegisterMultipleAndLogin() {
	cookieUserA := suite.loginAsUser(suite.UserA)

	// register first
	suite.mockCreateCredential(suite.UserA, "authenticator 1", suite.Credential1)
	suite.registerAuthenticator(suite.UserA, "authenticator 1", &suite.Authenticator1, suite.Credential1, cookieUserA)

	// register second
	suite.mockCreateCredential(suite.UserA, "authenticator 2", suite.Credential2)
	suite.registerAuthenticator(suite.UserA, "authenticator 2", &suite.Authenticator2, suite.Credential2, cookieUserA)

	// login with first
	suite.UserService.On("GetUserByWebAuthnID", suite.UserA.WebauthnID).Return(suite.UserA, nil).Once()
	suite.WebauthnService.On("UpdateCredential", mock.Anything).Return(nil).Once()
	suite.tryLogin(&suite.Authenticator1, suite.Credential1, suite.UserA, true)

	// login with second
	suite.UserService.On("GetUserByWebAuthnID", suite.UserA.WebauthnID).Return(suite.UserA, nil).Once()
	suite.WebauthnService.On("UpdateCredential", mock.Anything).Return(nil).Once()
	suite.tryLogin(&suite.Authenticator2, suite.Credential2, suite.UserA, true)

	suite.UserService.AssertExpectations(suite.T())
	suite.WebauthnService.AssertExpectations(suite.T())
}

func (suite *WebAuthnTestSuite) TestWebauthn_DeleteAndLoginFailure() {
	cookieUserA := suite.loginAsUser(suite.UserA)

	suite.UserService.On("GetUserById", suite.UserA.ID).Return(suite.UserA, nil)
	suite.mockCreateCredential(suite.UserA, "authenticator 1", suite.Credential1)
	suite.WebauthnService.On("LoadCredentialIntoUser", suite.UserA).Return(nil)
	suite.UserService.On("Update", mock.Anything).Return(suite.UserA, nil)
	suite.registerAuthenticator(suite.UserA, "authenticator 1", &suite.Authenticator1, suite.Credential1, cookieUserA)

	suite.mockDeleteCredential(suite.UserA, "authenticator 1")
	suite.deleteAuthenticator("authenticator 1", cookieUserA)

	suite.UserService.On("GetUserByWebAuthnID", suite.UserA.WebauthnID).Return(suite.UserA, nil)
	suite.tryLogin(&suite.Authenticator1, suite.Credential1, suite.UserA, false)

	suite.UserService.AssertExpectations(suite.T())
	suite.WebauthnService.AssertExpectations(suite.T())
}

func (suite *WebAuthnTestSuite) TestWebauthn_LoginFailureWithWrongAuthenticator() {
	suite.UserService.On("GetUserById", suite.UserA.ID).Return(suite.UserA, nil)
	suite.UserService.On("GetUserById", suite.UserB.ID).Return(suite.UserB, nil)
	suite.UserService.On("Update", mock.MatchedBy(func(user *models.User) bool { return user.ID == suite.UserA.ID })).Return(suite.UserA, nil)
	suite.UserService.On("Update", mock.MatchedBy(func(user *models.User) bool { return user.ID == suite.UserB.ID })).Return(suite.UserB, nil)
	cookieUserA := suite.getLoginCookies(suite.UserA.ID, suite.UserA.ID+"_password")
	cookieUserB := suite.getLoginCookies(suite.UserB.ID, suite.UserB.ID+"_password")

	suite.mockCreateCredential(suite.UserA, "authenticator 1", suite.Credential1)
	suite.mockCreateCredential(suite.UserB, "authenticator 2", suite.Credential2)
	suite.registerAuthenticator(suite.UserA, "authenticator 1", &suite.Authenticator1, suite.Credential1, cookieUserA)
	suite.registerAuthenticator(suite.UserB, "authenticator 2", &suite.Authenticator2, suite.Credential2, cookieUserB)

	suite.UserService.On("GetUserByWebAuthnID", suite.UserA.WebauthnID).Return(suite.UserA, nil)
	suite.UserService.On("GetUserByWebAuthnID", suite.UserB.WebauthnID).Return(suite.UserB, nil)
	suite.WebauthnService.On("UpdateCredential", mock.Anything).Return(nil).Twice()

	suite.tryLogin(&suite.Authenticator1, suite.Credential1, suite.UserA, true)
	suite.tryLogin(&suite.Authenticator2, suite.Credential2, suite.UserA, false)
	suite.tryLogin(&suite.Authenticator1, suite.Credential1, suite.UserB, false)
	suite.tryLogin(&suite.Authenticator2, suite.Credential2, suite.UserB, true)

	suite.UserService.AssertExpectations(suite.T())
	suite.WebauthnService.AssertExpectations(suite.T())
}

func (suite *WebAuthnTestSuite) TestWebauthn_RegisterDuplicateAuthenticatorNameDenied() {
	cookieUserA := suite.loginAsUser(suite.UserA)

	suite.mockCreateCredential(suite.UserA, "authenticator 1", suite.Credential1)
	suite.registerAuthenticator(suite.UserA, "authenticator 1", &suite.Authenticator1, suite.Credential1, cookieUserA)

	before := len(suite.UserA.Credentials)
	suite.registerAuthenticatorExpect(
		suite.UserA,
		"authenticator 1",
		&suite.Authenticator1,
		suite.Credential1,
		cookieUserA,
		http.StatusBadRequest,
		"authenticator name already in use",
	)
	suite.Equal(before, len(suite.UserA.Credentials))

	suite.UserService.AssertExpectations(suite.T())
	suite.WebauthnService.AssertExpectations(suite.T())
}

func (suite *WebAuthnTestSuite) TestWebauthn_LoginChallengeMismatchReplayProtection() {
	cookieUserA := suite.loginAsUser(suite.UserA)

	suite.mockCreateCredential(suite.UserA, "authenticator 1", suite.Credential1)
	suite.registerAuthenticator(suite.UserA, "authenticator 1", &suite.Authenticator1, suite.Credential1, cookieUserA)

	// create assertion for challenge #1
	optionsJSON1, cookies1 := suite.getLoginOption()
	parsedAssertionOptions1, err := virtualwebauthn.ParseAssertionOptions(optionsJSON1)
	suite.NoError(err)
	suite.Authenticator1.Options.UserHandle = []byte(suite.UserA.WebauthnID)
	assertionForOldChallenge := virtualwebauthn.CreateAssertionResponse(suite.RP, suite.Authenticator1, *suite.Credential1, *parsedAssertionOptions1)

	// obtain challenge #2 and deliberately submit assertion from challenge #1 with challenge #2 session
	_, cookies2 := suite.getLoginOption()

	suite.UserService.On("GetUserByWebAuthnID", suite.UserA.WebauthnID).Return(suite.UserA, nil).Maybe()

	rec := httptest.NewRecorder()
	reqBody := url.Values{
		"assertion_json": {assertionForOldChallenge},
	}
	req := httptest.NewRequest(http.MethodPost, "/webauthn/login", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range mergeCookies(cookies1, cookies2) {
		req.AddCookie(c)
	}

	suite.Router.ServeHTTP(rec, req)

	res := rec.Result()
	body, _ := io.ReadAll(res.Body)
	defer res.Body.Close()

	suite.Equal(http.StatusUnauthorized, res.StatusCode)
	suite.Contains(strings.ToLower(string(body)), "authentication failed")
	suite.WebauthnService.AssertNotCalled(suite.T(), "UpdateCredential", mock.Anything)

	suite.UserService.AssertExpectations(suite.T())
	suite.WebauthnService.AssertExpectations(suite.T())
}

func (suite *WebAuthnTestSuite) TestWebauthn_LoginSuccessThenTimeoutThenFail() {
	cookieUserA := suite.loginAsUser(suite.UserA)

	suite.mockCreateCredential(suite.UserA, "authenticator 1", suite.Credential1)
	suite.registerAuthenticator(suite.UserA, "authenticator 1", &suite.Authenticator1, suite.Credential1, cookieUserA)

	optionsJSON, webauthnCookies := suite.getLoginOption()
	parsedAssertionOptions, err := virtualwebauthn.ParseAssertionOptions(optionsJSON)
	suite.NoError(err)
	suite.Authenticator1.Options.UserHandle = []byte(suite.UserA.WebauthnID)
	assertion := virtualwebauthn.CreateAssertionResponse(suite.RP, suite.Authenticator1, *suite.Credential1, *parsedAssertionOptions)

	suite.UserService.On("GetUserByWebAuthnID", suite.UserA.WebauthnID).Return(suite.UserA, nil).Once()
	suite.WebauthnService.On("UpdateCredential", mock.Anything).Return(nil).Once()

	// first login succeeds
	firstRec := httptest.NewRecorder()
	firstReqBody := url.Values{"assertion_json": {assertion}}
	firstReq := httptest.NewRequest(http.MethodPost, "/webauthn/login", strings.NewReader(firstReqBody.Encode()))
	firstReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range webauthnCookies {
		firstReq.AddCookie(c)
	}
	suite.Router.ServeHTTP(firstRec, firstReq)
	firstRes := firstRec.Result()
	suite.Equal(http.StatusFound, firstRes.StatusCode)
	suite.Contains(firstRes.Header.Get("Location"), "/summary")

	// expire the webauthn session explicitly
	expireRec := httptest.NewRecorder()
	expireReq := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range mergeCookies(webauthnCookies, firstRes.Cookies()) {
		expireReq.AddCookie(c)
	}
	sess, _ := config.GetSessionStore().Get(expireReq, config.CookieKeySession)
	sess.Values[config.SessionValueWebAuthnExpiresAt] = time.Now().Add(-1 * time.Second).Unix()
	suite.NoError(sess.Save(expireReq, expireRec))

	// second login fails because session timed out
	secondRec := httptest.NewRecorder()
	secondReqBody := url.Values{"assertion_json": {assertion}}
	secondReq := httptest.NewRequest(http.MethodPost, "/webauthn/login", strings.NewReader(secondReqBody.Encode()))
	secondReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range mergeCookies(webauthnCookies, expireRec.Result().Cookies()) {
		secondReq.AddCookie(c)
	}
	suite.Router.ServeHTTP(secondRec, secondReq)

	secondRes := secondRec.Result()
	body, _ := io.ReadAll(secondRes.Body)
	defer secondRes.Body.Close()
	suite.Equal(http.StatusBadRequest, secondRes.StatusCode)
	suite.Contains(strings.ToLower(string(body)), "session expired")

	suite.UserService.AssertExpectations(suite.T())
	suite.WebauthnService.AssertExpectations(suite.T())
}

func (suite *WebAuthnTestSuite) getRegisterOption(cookies []*http.Cookie) (string, []*http.Cookie) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings/webauthn/options", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	suite.Router.ServeHTTP(rec, req)

	res := rec.Result()
	suite.Equal(http.StatusOK, res.StatusCode)

	body, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	return string(body), mergeCookies(cookies, res.Cookies())
}

func (suite *WebAuthnTestSuite) registerAuthenticator(user *models.User, name string, authenticator *virtualwebauthn.Authenticator, credential *virtualwebauthn.Credential, cookies []*http.Cookie) {
	suite.registerAuthenticatorExpect(user, name, authenticator, credential, cookies, http.StatusOK, "webauthn authenticator added successfully")
}

func (suite *WebAuthnTestSuite) registerAuthenticatorExpect(user *models.User, name string, authenticator *virtualwebauthn.Authenticator, credential *virtualwebauthn.Credential, cookies []*http.Cookie, expectedStatus int, expectedText string) {
	optionsJSON, cookies := suite.getRegisterOption(cookies)
	parsedAttestationOptions, err := virtualwebauthn.ParseAttestationOptions(optionsJSON)
	suite.NoError(err)

	*credential = virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)
	attestationResponse := virtualwebauthn.CreateAttestationResponse(suite.RP, *authenticator, *credential, *parsedAttestationOptions)
	authenticator.AddCredential(*credential)

	reqBody := url.Values{
		"action":             {"webauthn_add"},
		"credential_json":    {attestationResponse},
		"authenticator_name": {name},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	suite.Router.ServeHTTP(rec, req)
	body, _ := io.ReadAll(rec.Result().Body)
	defer rec.Result().Body.Close()
	suite.Equal(expectedStatus, rec.Result().StatusCode)
	suite.Contains(strings.ToLower(string(body)), strings.ToLower(expectedText))
}

func (suite *WebAuthnTestSuite) deleteAuthenticator(name string, cookies []*http.Cookie) {
	reqBody := url.Values{
		"action":          {"webauthn_delete"},
		"credential_name": {name},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	suite.Router.ServeHTTP(rec, req)
	res := rec.Result()
	suite.Equal(http.StatusOK, res.StatusCode)
}

func (suite *WebAuthnTestSuite) getLoginOption() (string, []*http.Cookie) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/webauthn/options", nil)
	suite.Router.ServeHTTP(rec, req)

	res := rec.Result()
	suite.Equal(http.StatusOK, res.StatusCode)

	body, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	return string(body), res.Cookies()
}

func (suite *WebAuthnTestSuite) tryLogin(authenticator *virtualwebauthn.Authenticator, credential *virtualwebauthn.Credential, expectedUser *models.User, shouldLogin bool) {
	optionsJSON, cookies := suite.getLoginOption()
	parsedAssertionOptions, err := virtualwebauthn.ParseAssertionOptions(optionsJSON)
	suite.NoError(err)

	authenticator.Options.UserHandle = []byte(expectedUser.WebauthnID) // in a real scenario, the client would store the user handle and send it back during login, but since we don't have a real client here, we need to set it manually for the assertion to be created correctly
	assertionResponse := virtualwebauthn.CreateAssertionResponse(suite.RP, *authenticator, *credential, *parsedAssertionOptions)

	reqBody := url.Values{
		"assertion_json": {assertionResponse},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webauthn/login", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	suite.Router.ServeHTTP(rec, req)

	res := rec.Result()
	if shouldLogin {
		suite.Equal(http.StatusFound, res.StatusCode)
		suite.Contains(res.Header.Get("Location"), "/summary")

		loginCookies := mergeCookies(cookies, res.Cookies())

		expectedAuthenticatorName := ""
		for _, c := range expectedUser.Credentials {
			if bytes.Equal(c.ID, credential.ID) {
				expectedAuthenticatorName = c.Name
				break
			}
		}
		suite.NotEmpty(expectedAuthenticatorName)

		settingsRec := httptest.NewRecorder()
		settingsReq := httptest.NewRequest(http.MethodGet, "/settings", nil)
		for _, c := range loginCookies {
			settingsReq.AddCookie(c)
		}
		suite.Router.ServeHTTP(settingsRec, settingsReq)

		settingsRes := settingsRec.Result()
		suite.Equal(http.StatusOK, settingsRes.StatusCode)
		settingsBody, _ := io.ReadAll(settingsRes.Body)
		defer settingsRes.Body.Close()
		suite.Contains(string(settingsBody), expectedAuthenticatorName)
	} else {
		suite.NotEqual(http.StatusFound, res.StatusCode)
	}
}

func mergeCookies(base []*http.Cookie, updates []*http.Cookie) []*http.Cookie {
	merged := map[string]*http.Cookie{}
	for _, c := range base {
		merged[c.Name] = c
	}
	for _, c := range updates {
		merged[c.Name] = c
	}
	out := make([]*http.Cookie, 0, len(merged))
	for _, c := range merged {
		out = append(out, c)
	}
	return out
}

func createUser() *models.User {
	userID := uuid.Must(uuid.NewV4()).String()
	passwdHash, _ := utils.HashPassword(userID+"_password", testPasswordSalt)
	return &models.User{
		ID:          userID,
		WebauthnID:  uuid.Must(uuid.NewV4()).String(),
		AuthType:    "local",
		Password:    passwdHash,
		Credentials: []*models.WebAuthnCredential{},
	}
}

func (suite *WebAuthnTestSuite) getLoginCookies(userID, password string) []*http.Cookie {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	// username=&password=
	reqBody := url.Values{
		"username": {userID},
		"password": {password},
	}
	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	suite.Router.ServeHTTP(rec, req)

	res := rec.Result()
	suite.Equal(http.StatusFound, res.StatusCode)
	return res.Cookies()
}

func createAndLoadConfig() {
	cfg := config.Empty()
	cfg.Security.PasswordSalt = testPasswordSalt
	hashKey := securecookie.GenerateRandomKey(64)
	blockKey := securecookie.GenerateRandomKey(32)
	sessionKey := securecookie.GenerateRandomKey(32)
	cfg.Security.SecureCookie = securecookie.New(hashKey, blockKey)
	cfg.Security.SessionKey = sessionKey
	cfg.Security.CookieMaxAgeSec = 120
	cfg.Security.PasswordResetMaxRate = "0/1m"
	cfg.Security.LoginMaxRate = "1000/1m"
	cfg.Security.SignupMaxRate = "0/1m"
	cfg.Server.PublicUrl = "https://example.com"
	config.Set(cfg)
	config.InitWebAuthn(cfg)
	config.ResetSessionStore()
}
