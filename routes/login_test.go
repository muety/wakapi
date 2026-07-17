package routes

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/securecookie"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/utils"
	testutils "github.com/muety/wakapi/utils/test"
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
	WebAuthnService       *mocks.WebAuthnServiceMock
	Cfg                   *config.Config
	Sut                   *LoginHandler
	OidcUserNew           *mockoidc.MockUser
	OidcUserExisting      *mockoidc.MockUser
	oidcMockDefaultConfig mockoidc.Config
}

const (
	testProvider             = "mock"
	testProvider2            = "otherProviderMock"
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
	suite.WebAuthnService = new(mocks.WebAuthnServiceMock)
	suite.UserService.On("Count").Return(1, nil).Maybe()

	cfg := config.Empty()
	cfg.Security.CookieKeyBytes = securecookie.GenerateRandomKey(128)
	cfg.Security.PasswordSalt = testPasswordSalt
	cfg.Security.LoginMaxRate = "100/1m"
	cfg.Security.SignupMaxRate = "100/1m"
	cfg.Security.PasswordResetMaxRate = "100/1m"
	config.Set(cfg)
	config.InitializeCookies()
	suite.Cfg = cfg

	suite.resetOidcMockTtl()
	suite.setupOidcProvider(testProvider)

	suite.Sut = NewLoginHandler(suite.UserService, nil, suite.KeyValueService, suite.WebAuthnService)
	Init() // load templates
}

func TestLoginHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(LoginHandlerTestSuite))
}

// Test cases
func (suite *LoginHandlerTestSuite) TestGetLogin_OnlyLocalAuth() {
	suite.Cfg.Security.DisableLocalAuth = false
	suite.Cfg.Security.OidcProviders = nil

	r := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	suite.Sut.GetIndex(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	// Test if the local login form exists
	assert.Contains(suite.T(), string(body), "Local Sign-On")
	assert.Contains(suite.T(), string(body), "id=\"username\"")
	assert.Contains(suite.T(), string(body), "id=\"password\"")
}

func (suite *LoginHandlerTestSuite) TestGetLogin_LocalAuthAndOIDC() {
	suite.Cfg.Security.DisableLocalAuth = false

	r := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	suite.Sut.GetIndex(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	// Test if the local login form and the oidc button exists
	assert.Contains(suite.T(), string(body), "Local Sign-On")
	assert.Contains(suite.T(), string(body), "id=\"username\"")
	assert.Contains(suite.T(), string(body), "id=\"password\"")
	assert.Contains(suite.T(), string(body), "Single Sign-On")
	assert.Contains(suite.T(), string(body), "Login with "+strutil.Capitalize(testProvider))
}

func (suite *LoginHandlerTestSuite) TestGetLogin_DirectRedirectToOidc() {
	suite.Cfg.Security.DisableLocalAuth = true

	r := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	suite.Sut.GetIndex(w, r)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Contains(suite.T(), w.Header().Get("Location"), "/oidc/"+testProvider+"/login")
}

func (suite *LoginHandlerTestSuite) TestGetLogin_TwoOidc() {
	suite.Cfg.Security.DisableLocalAuth = true
	suite.setupOidcProvider(testProvider2)

	r := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	suite.Sut.GetIndex(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	// Test if the two oidc buttons exist, no redirect expected
	assert.NotContains(suite.T(), string(body), "Local Sign-On")
	assert.Contains(suite.T(), string(body), "Single Sign-On")
	assert.Contains(suite.T(), string(body), "Login with "+strutil.Capitalize(testProvider))
	assert.Contains(suite.T(), string(body), "Login with "+strutil.Capitalize(testProvider2))
}

func (suite *LoginHandlerTestSuite) TestGetLogin_NoAuthenticationMethod() {
	suite.Cfg.Security.DisableLocalAuth = true
	suite.Cfg.Security.OidcProviders = nil

	r := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	suite.Sut.GetIndex(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Contains(suite.T(), string(body), "No authentication method is enabled or configured on this server")
}

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

	suite.Sut.PostLogin(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), string(body), "Missing parameters")
	suite.assertCookieAbsent(w, config.CookieKeyAuth, config.CookieKeyOidcIdToken, config.CookieKeyOidcRefreshToken, config.CookieKeyOidcProvider)
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

	suite.Sut.PostLogin(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	assert.Contains(suite.T(), string(body), "Resource not found")
	suite.assertCookieAbsent(w, config.CookieKeyAuth, config.CookieKeyOidcIdToken, config.CookieKeyOidcRefreshToken, config.CookieKeyOidcProvider)
}

func (suite *LoginHandlerTestSuite) TestPostLogin_WrongPassword() {
	form := url.Values{}
	form.Add("username", testUserExistingId)
	form.Add("password", "wrongpassword")

	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.UserService.On("GetUserById", testUserExistingId).Return(suite.TestUser, nil)

	suite.Sut.PostLogin(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), string(body), "Invalid credentials")
	suite.assertCookieAbsent(w, config.CookieKeyAuth, config.CookieKeyOidcIdToken, config.CookieKeyOidcRefreshToken, config.CookieKeyOidcProvider)
}

func (suite *LoginHandlerTestSuite) TestPostLogin_LocalAuthenticationDisabled_NonExistingUser() {
	suite.Cfg.Security.DisableLocalAuth = true

	form := url.Values{}
	form.Add("username", "nonexisting")
	form.Add("password", testUserExistingPassword)

	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.Sut.PostLogin(w, r)
	body, _ := io.ReadAll(w.Body)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	assert.Contains(suite.T(), string(body), "Local authentication is disabled on this server")
	suite.assertCookieAbsent(w, config.CookieKeyAuth, config.CookieKeyOidcIdToken, config.CookieKeyOidcRefreshToken, config.CookieKeyOidcProvider)
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

	suite.UserService.On("Count").Return(1, nil)
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

func (suite *LoginHandlerTestSuite) TestPostSignup_Success_FirstUserIsAdmin() {
	form := url.Values{}
	form.Add("username", testUserNewId)
	form.Add("email", testUserNewEmail)
	form.Add("password", testUserNewPassword)
	form.Add("password_repeat", testUserNewPassword)

	r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.UserService.On("Count").Unset()
	suite.UserService.On("Count").Return(0, nil)
	suite.UserService.On("CreateOrGet", mock.Anything, mock.Anything).Return(&models.User{}, true, nil)
	suite.Cfg.Security.AllowSignup = true
	suite.Cfg.Security.OidcAllowSignup = false

	suite.Sut.PostSignup(w, r)

	suite.UserService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.True(suite.T(), suite.UserService.Calls[1].Arguments[1].(bool))
}

func (suite *LoginHandlerTestSuite) TestPostSignup_InvalidForm() {
	form := url.Values{}
	form.Add("username", "")
	form.Add("password", testUserNewPassword)
	form.Add("password_repeat", testUserNewPassword)

	r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.Cfg.Security.AllowSignup = true
	suite.Cfg.Security.OidcAllowSignup = false

	suite.Sut.PostSignup(w, r)
	body, _ := io.ReadAll(w.Body)

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

	suite.UserService.On("Count").Return(1, nil)
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

	suite.Sut.PostSignup(w, r)
	body, _ := io.ReadAll(w.Body)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	assert.Contains(suite.T(), string(body), "Registration is disabled on this server")
}

func (suite *LoginHandlerTestSuite) TestPostSignup_LocalAuthenticationDisabled() {
	suite.Cfg.Security.DisableLocalAuth = true
	suite.Cfg.Security.AllowSignup = true

	form := url.Values{}
	form.Add("username", testUserNewId)
	form.Add("password", testUserNewPassword)
	form.Add("password_repeat", testUserNewPassword)

	r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	suite.Sut.PostSignup(w, r)
	body, _ := io.ReadAll(w.Body)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	assert.Contains(suite.T(), string(body), "Local authentication is disabled on this server.")
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
	url := suite.authorizeUser(suite.OidcUserExisting, testProvider)
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

	testutils.AssertContainsHeaderMatching(suite.T(), w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_id_token=")
	}, "OIDC id_token cookie not set in response")
	testutils.AssertContainsHeaderMatching(suite.T(), w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_provider="+testProvider)
	}, "OIDC provider cookie not set in response")
	testutils.AssertContainsHeaderMatching(suite.T(), w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_refresh_token=")
	}, "OIDC refresh token not set in response")
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_Success_CreateUser() {
	suite.Cfg.Security.AllowSignup = false
	suite.Cfg.Security.OidcAllowSignup = true

	url := suite.authorizeUser(suite.OidcUserNew, testProvider)
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

	testutils.AssertContainsHeaderMatching(suite.T(), w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_id_token=")
	}, "OIDC id_token cookie not set in response")
	testutils.AssertContainsHeaderMatching(suite.T(), w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_provider="+testProvider)
	}, "OIDC provider cookie not set in response")
	testutils.AssertContainsHeaderMatching(suite.T(), w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_refresh_token=")
	}, "OIDC refresh token not set in response")
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_Success_CreateUser_CustomUsernameClaim() {
	suite.Cfg.Security.AllowSignup = false
	suite.Cfg.Security.OidcAllowSignup = true

	const customClaimName = "custom_username"
	const customUsername = "custom_user_from_claim"

	config.WithOidcProvider(suite.Cfg, "custom-claim-provider", suite.OidcMock.ClientID, suite.OidcMock.ClientSecret, suite.OidcMock.Addr()+"/oidc", customClaimName)

	customUser := &WakapiMockOIDCUser{
		MockUser: mockoidc.MockUser{
			Subject:           "custom-sub-123",
			Email:             "custom@example.org",
			PreferredUsername: "preferred_user",
		},
		CustomClaimName:  customClaimName,
		CustomClaimValue: customUsername,
	}

	url := suite.authorizeUser(customUser, "custom-claim-provider")
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r = WithUrlParam(r, "provider", "custom-claim-provider")
	w := httptest.NewRecorder()

	routeutils.SetOidcState(testOauthState, r, w)
	suite.UserService.On("GetUserByOidc", "custom-claim-provider", customUser.Subject).Return(nil, errors.New(""))
	suite.UserService.On("GetUserById", customUsername).Return(nil, errors.New(""))
	suite.UserService.On("CreateOrGet", mock.Anything, mock.Anything).Return(suite.TestUser, true, nil)
	suite.UserService.On("Update", mock.Anything).Return(suite.TestUser, nil)

	suite.Sut.GetOidcCallback(w, r)

	argSignup := suite.UserService.Calls[2].Arguments[0].(*models.Signup)
	argIsAdmin := suite.UserService.Calls[2].Arguments[1].(bool)

	suite.UserService.AssertExpectations(suite.T())
	// verify that the username comes from the custom claim, not preferred_username
	assert.Equal(suite.T(), customUsername, argSignup.Username)
	assert.Equal(suite.T(), customUser.Email, argSignup.Email)
	assert.Equal(suite.T(), customUser.Subject, argSignup.OidcSubject)
	assert.Equal(suite.T(), "custom-claim-provider", argSignup.OidcProvider)
	assert.NotEmpty(suite.T(), argSignup.Password)
	assert.False(suite.T(), argIsAdmin)
	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Empty(suite.T(), suite.getSessionError(r))
	assert.Equal(suite.T(), "/summary", w.Header().Get("Location"))

	testutils.AssertContainsHeaderMatching(suite.T(), w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_id_token=")
	}, "OIDC id_token cookie not set in response")
	testutils.AssertContainsHeaderMatching(suite.T(), w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_provider=custom-claim-provider")
	}, "OIDC provider cookie not set in response")
	testutils.AssertContainsHeaderMatching(suite.T(), w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_refresh_token=")
	}, "OIDC refresh token not set in response")
}

func (suite *LoginHandlerTestSuite) TestGetOidcLogin_CustomScopesRequested() {
	customScopes := []string{"groups", "roles", "offline_access"}
	config.WithOidcProviderAndScopes(suite.Cfg, "custom-scopes-provider", suite.OidcMock.ClientID, suite.OidcMock.ClientSecret, suite.OidcMock.Addr()+"/oidc", "", customScopes)

	r := httptest.NewRequest(http.MethodGet, "/oidc/{provider}/login", nil)
	r = WithUrlParam(r, "provider", "custom-scopes-provider")
	w := httptest.NewRecorder()

	suite.Sut.GetOidcLogin(w, r)

	assert.Equal(suite.T(), http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	assert.True(suite.T(), strings.HasPrefix(location, suite.OidcMock.AuthorizationEndpoint()))

	// verify that custom (and default) scopes are included in the auth url
	for _, scope := range customScopes {
		assert.Contains(suite.T(), location, "scope=")
		assert.Contains(suite.T(), location, scope)
	}
	assert.Contains(suite.T(), location, "openid")
	assert.Contains(suite.T(), location, "profile")
	assert.Contains(suite.T(), location, "email")
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_SignupDisabled() {
	suite.Cfg.Security.AllowSignup = true
	suite.Cfg.Security.OidcAllowSignup = false

	url := suite.authorizeUser(suite.OidcUserNew, testProvider)
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
	suite.assertCookieAbsent(w, config.CookieKeyAuth, config.CookieKeyOidcIdToken, config.CookieKeyOidcRefreshToken, config.CookieKeyOidcProvider)
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_InvalidState() {
	url := suite.authorizeUser(suite.OidcUserNew, testProvider)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r = WithUrlParam(r, "provider", testProvider)
	w := httptest.NewRecorder()

	routeutils.SetOidcState("some-other-state", r, w)

	suite.Sut.GetOidcCallback(w, r)

	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Equal(suite.T(), "suspicious operation, got invalid state in oidc callback", suite.getSessionError(r))
	assert.Equal(suite.T(), "/login", w.Header().Get("Location"))
	suite.assertCookieAbsent(w, config.CookieKeyAuth, config.CookieKeyOidcIdToken, config.CookieKeyOidcRefreshToken, config.CookieKeyOidcProvider)
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_AuthExchangeFailure() {
	url := suite.authorizeUser(suite.OidcUserNew, testProvider)
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
	suite.assertCookieAbsent(w, config.CookieKeyAuth, config.CookieKeyOidcIdToken, config.CookieKeyOidcRefreshToken, config.CookieKeyOidcProvider)
}

func (suite *LoginHandlerTestSuite) TestGetOidcLoginCallback_IdTokenExpired() {
	url := suite.authorizeUser(suite.OidcUserNew, testProvider)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r = WithUrlParam(r, "provider", testProvider)
	w := httptest.NewRecorder()

	routeutils.SetOidcState(testOauthState, r, w)

	suite.OidcMock.AccessTTL = 0 // related: https://github.com/oauth2-proxy/mockoidc/issues/38

	suite.Sut.GetOidcCallback(w, r)

	assert.Equal(suite.T(), http.StatusFound, w.Code)
	assert.Equal(suite.T(), "failed to verify and decode id_token", suite.getSessionError(r))
	assert.Equal(suite.T(), "/login", w.Header().Get("Location"))
	suite.assertCookieAbsent(w, config.CookieKeyAuth, config.CookieKeyOidcIdToken, config.CookieKeyOidcRefreshToken, config.CookieKeyOidcProvider)
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
	config.WithOidcProvider(suite.Cfg, name, suite.OidcMock.ClientID, suite.OidcMock.ClientSecret, suite.OidcMock.Addr()+"/oidc", "")
}

func (suite *LoginHandlerTestSuite) getSessionError(r *http.Request) string {
	session, _ := config.GetSessionStore().Get(r, config.CookieKeySession)
	if errors := session.Flashes("error"); len(errors) > 0 {
		return errors[0].(string)
	}
	return ""
}

func (suite *LoginHandlerTestSuite) authorizeUser(user mockoidc.User, provider string) string {
	r := httptest.NewRequest(http.MethodGet, suite.OidcMock.AuthorizationEndpoint(), nil)
	q := r.URL.Query()
	q.Set("code", testOauthCode)
	q.Set("client_id", suite.OidcMock.ClientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("state", testOauthState)
	q.Set("redirect_uri", fmt.Sprintf("/oidc/%s/callback", provider))
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

func (suite *LoginHandlerTestSuite) assertCookieAbsent(w *httptest.ResponseRecorder, keys ...string) {
	cookies := w.Result().Cookies()
	if len(keys) == 0 {
		assert.Empty(suite.T(), cookies)
		return
	}
	for _, c := range cookies {
		for _, k := range keys {
			if c.Name == k {
				suite.FailNowf("cookie set", "Expected cookie %q to be absent, but got: %s", k, c.Raw)
			}
		}
	}
}

func (suite *LoginHandlerTestSuite) TestPostLogin_RateLimiting() {
	suite.Cfg.Security.LoginMaxRate = "2/1m"
	suite.Cfg.Security.ParseTrustReverseProxyIPs()

	router := chi.NewRouter()
	router.Use(middleware.ClientIPFromRemoteAddr)
	suite.Sut.RegisterRoutes(router)

	form := url.Values{}
	form.Add("username", testUserExistingId)
	form.Add("password", testUserExistingPassword)

	suite.UserService.On("GetUserById", testUserExistingId).Return(suite.TestUser, nil)
	suite.UserService.On("Update", mock.Anything).Return(suite.TestUser, nil)

	// First request - 302 Found
	req1 := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req1.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(suite.T(), http.StatusFound, w1.Code)

	// Second request - 302 Found
	req2 := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req2.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(suite.T(), http.StatusFound, w2.Code)

	// Third request - 429 Too Many Requests
	req3 := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req3.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(suite.T(), http.StatusTooManyRequests, w3.Code)
}

func (suite *LoginHandlerTestSuite) TestPostLogin_RateLimiting_TrustReverseProxy() {
	suite.Cfg.Security.TrustReverseProxyIps = "192.168.0.0/24"
	suite.Cfg.Security.LoginMaxRate = "1/1m"
	suite.Cfg.Security.SignupMaxRate = "100/1m"
	suite.Cfg.Security.PasswordResetMaxRate = "100/1m"
	suite.Cfg.Security.ParseTrustReverseProxyIPs()

	router := chi.NewRouter() // analogously to main.go
	trustedProxies := suite.Cfg.Security.TrustReverseProxyIPs()
	if len(trustedProxies) > 0 {
		cidrs := slice.Map[net.IPNet, string](trustedProxies, func(_ int, ipNet net.IPNet) string {
			return ipNet.String()
		})
		router.Use(middleware.ClientIPFromXFF(cidrs...))
	} else {
		router.Use(middleware.ClientIPFromRemoteAddr)
	}
	suite.Sut.RegisterRoutes(router)

	form := url.Values{}
	form.Add("username", testUserExistingId)
	form.Add("password", testUserExistingPassword)

	suite.UserService.On("GetUserById", testUserExistingId).Return(suite.TestUser, nil)
	suite.UserService.On("Update", mock.Anything).Return(suite.TestUser, nil)

	// Scenario: spoofing attempt through trusted proxy
	// Request from trusted proxy (192.168.0.10), representing client (1.1.1.1) who tries to spoof 2.2.2.2.
	reqA1 := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	reqA1.RemoteAddr = "192.168.0.10:12345"
	reqA1.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	reqA1.Header.Add("X-Forwarded-For", "2.2.2.2, 1.1.1.1") // Left is spoofed, right is appended by trusted proxy
	wA1 := httptest.NewRecorder()
	router.ServeHTTP(wA1, reqA1)
	assert.Equal(suite.T(), http.StatusFound, wA1.Code)

	// Subsequent request from trusted proxy representing same client (1.1.1.1) trying to spoof different IP (3.3.3.3).
	reqA2 := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	reqA2.RemoteAddr = "192.168.0.10:12345"
	reqA2.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	reqA2.Header.Add("X-Forwarded-For", "3.3.3.3, 1.1.1.1")
	wA2 := httptest.NewRecorder()
	router.ServeHTTP(wA2, reqA2)
	assert.Equal(suite.T(), http.StatusTooManyRequests, wA2.Code)

	// Scenario: Legitimate proxy requests from trusted proxy
	// Trusted proxy forwards request from client (5.5.5.5). Should map to 5.5.5.5. (Status 302)
	reqB1 := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	reqB1.RemoteAddr = "192.168.0.10:12345"
	reqB1.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	reqB1.Header.Add("X-Forwarded-For", "5.5.5.5")
	wB1 := httptest.NewRecorder()
	router.ServeHTTP(wB1, reqB1)
	assert.Equal(suite.T(), http.StatusFound, wB1.Code)

	// Trusted proxy forwards request from client (6.6.6.6). Should map to 6.6.6.6. (Status 302)
	reqB2 := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	reqB2.RemoteAddr = "192.168.0.10:12345"
	reqB2.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	reqB2.Header.Add("X-Forwarded-For", "6.6.6.6")
	wB2 := httptest.NewRecorder()
	router.ServeHTTP(wB2, reqB2)
	assert.Equal(suite.T(), http.StatusFound, wB2.Code)
}

// TODO: test all remaining endpoints
