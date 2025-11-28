package routes

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/utils"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type LoginHandlerTestSuite struct {
	suite.Suite
	TestUser              *models.User
	OidcMock              *mockoidc.MockOIDC
	UserService           *mocks.UserServiceMock
	KeyValueService       *mocks.KeyValueServiceMock
	Cfg                   *config.Config
	Sut                   *LoginHandler
	OidcUserNew           *mockoidc.MockUser
	OidcUserExisting      *mockoidc.MockUser
	oidcMockDefaultConfig mockoidc.Config
}

const (
	testProvider             = "mock"
	testOauthCode            = "some-code"
	testOauthState           = "some-state"
	testUserExistingId       = "user1"
	testUserExistingEmail    = "foo@example.org"
	testUserExistingSub      = "111"
	testUserExistingPassword = "supersecret"
	testUserNewId            = "user2"
	testUserNewEmail         = "bar@example.org"
	testUserNewSub           = "222"
	testUserNewPassword      = "ssssshhhhhh"
	testPasswordSalt         = "salty"
)

func (suite *LoginHandlerTestSuite) SetupSuite() {
	if m, err := mockoidc.Run(); err == nil {
		suite.OidcMock = m
		suite.oidcMockDefaultConfig = *suite.OidcMock.Config()
	}

	testUserPassword, _ := utils.HashPassword(testUserExistingPassword, testPasswordSalt)

	suite.OidcUserNew = &mockoidc.MockUser{
		Subject:           testUserNewSub,
		Email:             testUserNewEmail,
		PreferredUsername: testUserNewId,
	}

	suite.OidcUserExisting = &mockoidc.MockUser{
		Subject:           testUserExistingSub,
		Email:             testUserExistingEmail,
		PreferredUsername: testUserExistingId,
	}

	suite.TestUser = &models.User{
		ID:       testUserExistingId,
		Email:    testUserExistingEmail,
		AuthType: testProvider,
		Sub:      testUserExistingSub,
		Password: testUserPassword,
	}
}

func (suite *LoginHandlerTestSuite) TearDownSuite() {
	suite.OidcMock.Shutdown()
}

func (suite *LoginHandlerTestSuite) BeforeTest(suiteName, testName string) {
	suite.UserService = new(mocks.UserServiceMock)
	suite.KeyValueService = new(mocks.KeyValueServiceMock)

	cfg := config.Empty()
	cfg.Security.SecureCookie = securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)
	cfg.Security.PasswordSalt = testPasswordSalt
	config.Set(cfg)
	suite.Cfg = cfg

	suite.resetOidcMockTtl()
	suite.setupOidcProvider(testProvider)

	suite.Sut = NewLoginHandler(suite.UserService, nil, suite.KeyValueService)
	Init() // load templates
}

func TestLoginHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(LoginHandlerTestSuite))
}

// Test cases

func (suite *LoginHandlerTestSuite) TestPostLogin_Success() {
	form := url.Values{}
	form.Add("username", testUserExistingId)
	form.Add("password", testUserExistingPassword)

	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.UserService.On("GetUserById", testUserExistingId).Return(suite.TestUser, nil)
	suite.UserService.On("Update", mock.Anything).Return(suite.TestUser, nil)

	suite.Sut.PostLogin(w, r)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Equal(suite.T(), "/summary", w.Header().Get("Location"))
	assert.Contains(suite.T(), w.Header().Get("Set-Cookie"), "wakapi_auth=")
}

func (suite *LoginHandlerTestSuite) TestPostLogin_ValidAuthCookie() {
	// TODO: implement this
}

func (suite *LoginHandlerTestSuite) TestPostLogin_EmptyLoginForm() {
	form := url.Values{}

	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.UserService.On("Count").Return(1, nil)

	suite.Sut.PostLogin(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), string(body), "Missing parameters")
	assert.Empty(suite.T(), w.Header().Get("Set-Cookie"))
}

func (suite *LoginHandlerTestSuite) TestPostLogin_NonExistingUser() {
	form := url.Values{}
	form.Add("username", "nonexisting")
	form.Add("password", testUserExistingPassword)

	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.UserService.On("GetUserById", "nonexisting").Return(nil, errors.New(""))
	suite.UserService.On("GetUserByEmail", "nonexisting").Return(nil, errors.New(""))
	suite.UserService.On("Count").Return(1, nil)

	suite.Sut.PostLogin(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	assert.Contains(suite.T(), string(body), "Resource not found")
	assert.Empty(suite.T(), w.Header().Get("Set-Cookie"))
}

func (suite *LoginHandlerTestSuite) TestPostLogin_WrongPassword() {
	form := url.Values{}
	form.Add("username", testUserExistingId)
	form.Add("password", "wrongpassword")

	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.UserService.On("GetUserById", testUserExistingId).Return(suite.TestUser, nil)
	suite.UserService.On("Count").Return(1, nil)

	suite.Sut.PostLogin(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), string(body), "Invalid credentials")
	assert.Empty(suite.T(), w.Header().Get("Set-Cookie"))
}

func (suite *LoginHandlerTestSuite) TestPostSignup_Success() {
	form := url.Values{}
	form.Add("username", testUserNewId)
	form.Add("email", testUserNewEmail)
	form.Add("password", testUserNewPassword)
	form.Add("password_repeat", testUserNewPassword)

	r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.UserService.On("Count", mock.Anything).Return(1, nil)
	suite.UserService.On("CreateOrGet", mock.Anything, mock.Anything).Return(&models.User{}, true, nil)
	suite.Cfg.Security.AllowSignup = true
	suite.Cfg.Security.OidcAllowSignup = false

	suite.Sut.PostSignup(w, r)

	argSignup := suite.UserService.Calls[1].Arguments[0].(*models.Signup)
	argIsAdmin := suite.UserService.Calls[1].Arguments[1].(bool)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Equal(suite.T(), testUserNewId, argSignup.Username)
	assert.Equal(suite.T(), testUserNewEmail, argSignup.Email)
	assert.Equal(suite.T(), testUserNewPassword, argSignup.Password)
	assert.False(suite.T(), argIsAdmin)
	assert.Equal(suite.T(), "/", w.Header().Get("Location"))
}

func (suite *LoginHandlerTestSuite) TestPostSignup_InvalidForm() {
	form := url.Values{}
	form.Add("username", "")
	form.Add("password", testUserNewPassword)
	form.Add("password_repeat", testUserNewPassword)

	r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.UserService.On("Count", mock.Anything).Return(1, nil)
	suite.Cfg.Security.AllowSignup = true
	suite.Cfg.Security.OidcAllowSignup = false

	suite.Sut.PostSignup(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), string(body), "User name is invalid")
}

func (suite *LoginHandlerTestSuite) TestPostSignup_ExistingUser() {
	form := url.Values{}
	form.Add("username", testUserExistingId)
	form.Add("password", testUserExistingPassword)
	form.Add("password_repeat", testUserExistingPassword)

	r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.UserService.On("Count", mock.Anything).Return(1, nil)
	suite.UserService.On("CreateOrGet", mock.Anything, mock.Anything).Return(suite.TestUser, false, nil)
	suite.Cfg.Security.AllowSignup = true
	suite.Cfg.Security.OidcAllowSignup = false

	suite.Sut.PostSignup(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusConflict, w.Code)
	assert.Contains(suite.T(), string(body), "User already existing")
}

func (suite *LoginHandlerTestSuite) TestPostSignup_SignupDisabled() {
	suite.Cfg.Security.AllowSignup = false
	suite.Cfg.Security.OidcAllowSignup = true

	form := url.Values{}
	form.Add("username", testUserNewId)
	form.Add("password", testUserNewPassword)
	form.Add("password_repeat", testUserNewPassword)

	r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.UserService.On("Count", mock.Anything).Return(1, nil)

	suite.Sut.PostSignup(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	assert.Contains(suite.T(), string(body), "Registration is disabled on this server")
}

func (suite *LoginHandlerTestSuite) TestGetOidcLogin_Redirect() {
	r := httptest.NewRequest(http.MethodGet, "/oidc/{provider}/login", nil)
	r = WithUrlParam(r, "provider", testProvider)
	w := httptest.NewRecorder()

	suite.Sut.GetOidcLogin(w, r)

	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.True(suite.T(), strings.HasPrefix(w.Header().Get("Location"), suite.OidcMock.AuthorizationEndpoint()))
	assert.Contains(suite.T(), w.Header().Get("Location"), fmt.Sprintf("state=%s", routeutils.GetOidcState(r)))
}

func (suite *LoginHandlerTestSuite) TestGetOidcLogin_NoMatchingProvider() {
	r := httptest.NewRequest(http.MethodGet, "/oidc/{provider}/login", nil)
	r = WithUrlParam(r, "provider", "mock2")
	w := httptest.NewRecorder()

	suite.Sut.GetOidcCallback(w, r)

	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Equal(suite.T(), "/login", w.Header().Get("Location"))
	assert.Equal(suite.T(), "oidc provider \"mock2\" not registered", suite.getSessionError(r))
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_Success() {
	url := suite.authorizeUser(suite.OidcUserExisting)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r = WithUrlParam(r, "provider", testProvider)
	w := httptest.NewRecorder()

	routeutils.SetOidcState(testOauthState, r, w)
	suite.UserService.On("GetUserByOidc", testProvider, suite.OidcUserExisting.Subject).Return(suite.TestUser, nil)
	suite.UserService.On("Update", mock.Anything).Return(suite.TestUser, nil)

	suite.Sut.GetOidcCallback(w, r)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Empty(suite.T(), suite.getSessionError(r))
	assert.Equal(suite.T(), "/summary", w.Header().Get("Location"))
	assert.Contains(suite.T(), w.Header().Get("Set-Cookie"), "wakapi_auth=")
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_Success_CreateUser() {
	suite.Cfg.Security.AllowSignup = false
	suite.Cfg.Security.OidcAllowSignup = true

	url := suite.authorizeUser(suite.OidcUserNew)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r = WithUrlParam(r, "provider", testProvider)
	w := httptest.NewRecorder()

	routeutils.SetOidcState(testOauthState, r, w)
	suite.UserService.On("GetUserByOidc", testProvider, suite.OidcUserNew.Subject).Return(nil, errors.New(""))
	suite.UserService.On("GetUserById", suite.OidcUserNew.PreferredUsername).Return(nil, errors.New(""))
	suite.UserService.On("CreateOrGet", mock.Anything, mock.Anything).Return(suite.TestUser, true, nil)
	suite.UserService.On("Update", mock.Anything).Return(suite.TestUser, nil)

	suite.Sut.GetOidcCallback(w, r)

	argSignup := suite.UserService.Calls[2].Arguments[0].(*models.Signup)
	argIsAdmin := suite.UserService.Calls[2].Arguments[1].(bool)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), suite.OidcUserNew.PreferredUsername, argSignup.Username)
	assert.Equal(suite.T(), suite.OidcUserNew.Email, argSignup.Email)
	assert.Equal(suite.T(), suite.OidcUserNew.Subject, argSignup.OidcSubject)
	assert.Equal(suite.T(), testProvider, argSignup.OidcProvider)
	assert.NotEmpty(suite.T(), argSignup.Password)
	assert.False(suite.T(), argIsAdmin)
	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Empty(suite.T(), suite.getSessionError(r))
	assert.Equal(suite.T(), "/summary", w.Header().Get("Location"))
	assert.Contains(suite.T(), w.Header().Get("Set-Cookie"), "wakapi_auth=")
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_SignupDisabled() {
	suite.Cfg.Security.AllowSignup = true
	suite.Cfg.Security.OidcAllowSignup = false

	url := suite.authorizeUser(suite.OidcUserNew)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r = WithUrlParam(r, "provider", testProvider)
	w := httptest.NewRecorder()

	routeutils.SetOidcState(testOauthState, r, w)
	suite.UserService.On("GetUserByOidc", testProvider, suite.OidcUserNew.Subject).Return(nil, errors.New(""))

	suite.Sut.GetOidcCallback(w, r)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Equal(suite.T(), "registration is disabled on this server", suite.getSessionError(r))
	assert.Equal(suite.T(), "/login", w.Header().Get("Location"))
	assert.Empty(suite.T(), w.Header().Get("Set-Cookie"))
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_InvalidState() {
	url := suite.authorizeUser(suite.OidcUserNew)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r = WithUrlParam(r, "provider", testProvider)
	w := httptest.NewRecorder()

	routeutils.SetOidcState("some-other-state", r, w)

	suite.Sut.GetOidcCallback(w, r)

	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Equal(suite.T(), "suspicious operation, got invalid state in oidc callback", suite.getSessionError(r))
	assert.Equal(suite.T(), "/login", w.Header().Get("Location"))
	assert.Empty(suite.T(), w.Header().Get("Set-Cookie"))
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_AuthExchangeFailure() {
	url := suite.authorizeUser(suite.OidcUserNew)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r = WithUrlParam(r, "provider", testProvider)
	w := httptest.NewRecorder()

	routeutils.SetOidcState(testOauthState, r, w)

	// token endpoint will be called twice, see https://github.com/golang/oauth2/blob/792c8776358f0c8689d84eef0d0c966937d560fb/internal/token.go#L231-L243
	suite.OidcMock.QueueError(&mockoidc.ServerError{Code: http.StatusInternalServerError})
	suite.OidcMock.QueueError(&mockoidc.ServerError{Code: http.StatusInternalServerError})

	suite.Sut.GetOidcCallback(w, r)

	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Equal(suite.T(), "failed to exchange authorization code for access token", suite.getSessionError(r))
	assert.Equal(suite.T(), "/login", w.Header().Get("Location"))
	assert.Empty(suite.T(), w.Header().Get("Set-Cookie"))
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_IdTokenExpired() {
	url := suite.authorizeUser(suite.OidcUserNew)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r = WithUrlParam(r, "provider", testProvider)
	w := httptest.NewRecorder()

	routeutils.SetOidcState(testOauthState, r, w)

	suite.OidcMock.AccessTTL = 0 // related: https://github.com/oauth2-proxy/mockoidc/issues/38

	suite.Sut.GetOidcCallback(w, r)

	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Equal(suite.T(), "failed to verify and decode id_token", suite.getSessionError(r))
	assert.Equal(suite.T(), "/login", w.Header().Get("Location"))
	assert.Empty(suite.T(), w.Header().Get("Set-Cookie"))
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_NoMatchingProvider() {
	r := httptest.NewRequest(http.MethodGet, "/oidc/{provider}/callback", nil)
	r = WithUrlParam(r, "provider", "mock2")
	w := httptest.NewRecorder()

	suite.Sut.GetOidcLogin(w, r)

	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Equal(suite.T(), "/login", w.Header().Get("Location"))
	assert.Equal(suite.T(), "oidc provider \"mock2\" not registered", suite.getSessionError(r))
}

// Private utility methods
func (suite *LoginHandlerTestSuite) setupOidcProvider(name string) {
	config.WithOidcProvider(suite.Cfg, name, suite.OidcMock.ClientID, suite.OidcMock.ClientSecret, suite.OidcMock.Addr()+"/oidc")
}

func (suite *LoginHandlerTestSuite) getSessionError(r *http.Request) string {
	session, _ := config.GetSessionStore().Get(r, config.CookieKeySession)
	if errors := session.Flashes("error"); len(errors) > 0 {
		return errors[0].(string)
	}
	return ""
}

func (suite *LoginHandlerTestSuite) authorizeUser(user *mockoidc.MockUser) string { // returns the location header's redirect url
	r := httptest.NewRequest(http.MethodGet, suite.OidcMock.AuthorizationEndpoint(), nil)
	q := r.URL.Query()
	q.Set("code", testOauthCode)
	q.Set("client_id", suite.OidcMock.ClientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("state", testOauthState)
	q.Set("redirect_uri", fmt.Sprintf("/oidc/%s/callback", testProvider))
	r.URL.RawQuery = q.Encode()
	w := httptest.NewRecorder()

	suite.OidcMock.QueueUser(user)
	suite.OidcMock.QueueCode(testOauthCode)

	suite.OidcMock.Authorize(w, r)
	return w.Header().Get("Location")
}

func (suite *LoginHandlerTestSuite) resetOidcMockTtl() {
	suite.OidcMock.AccessTTL = 600 * time.Second
	suite.OidcMock.RefreshTTL = 60 * time.Minute
}

// TODO: test all remaining endpoints
