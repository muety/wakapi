package middlewares

import (
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gofrs/uuid/v5"
	"github.com/muety/wakapi/helpers"
	"net"
	"net/http"
	"strings"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

const (
	// queryApiKey is the query parameter name for api key.
	queryApiKey = "api_key"
)

var (
	errEmptyKey = fmt.Errorf("the api_key is empty")
)

type AuthenticateMiddleware struct {
	config               *conf.Config
	userSrvc             services.IUserService
	optionalForPaths     []string
	optionalForMethods   []string
	redirectTarget       string // optional
	redirectErrorMessage string // optional
}

func NewAuthenticateMiddleware(userService services.IUserService) *AuthenticateMiddleware {
	return &AuthenticateMiddleware{
		config:             conf.Get(),
		userSrvc:           userService,
		optionalForPaths:   []string{},
		optionalForMethods: []string{},
	}
}

func (m *AuthenticateMiddleware) WithOptionalFor(paths ...string) *AuthenticateMiddleware {
	m.optionalForPaths = paths
	return m
}

func (m *AuthenticateMiddleware) WithOptionalForMethods(methods ...string) *AuthenticateMiddleware {
	m.optionalForMethods = methods
	return m
}

func (m *AuthenticateMiddleware) WithRedirectTarget(path string) *AuthenticateMiddleware {
	m.redirectTarget = path
	return m
}

func (m *AuthenticateMiddleware) WithRedirectErrorMessage(message string) *AuthenticateMiddleware {
	m.redirectErrorMessage = message
	return m
}

func (m *AuthenticateMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.ServeHTTP(w, r, h.ServeHTTP)
	})
}

func (m *AuthenticateMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var user *models.User

	user, err := m.tryGetUserByCookie(r)
	if err != nil {
		user, err = m.tryGetUserByApiKeyHeader(r)
	}
	if err != nil {
		user, err = m.tryGetUserByApiKeyQuery(r)
	}
	if err != nil && m.config.Security.TrustedHeaderAuth {
		user, err = m.tryGetUserByTrustedHeader(r, m.config.Security.TrustedHeaderAuthAllowSignup)
	}

	if err != nil || user == nil {
		if m.isOptional(r) {
			next(w, r)
			return
		}

		if m.redirectTarget == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(conf.ErrUnauthorized))
		} else {
			if m.redirectErrorMessage != "" {
				session, _ := conf.GetSessionStore().Get(r, conf.SessionKeyDefault)
				session.AddFlash(m.redirectErrorMessage, "error")
				session.Save(r, w)
			}
			http.SetCookie(w, m.config.GetClearCookie(models.AuthCookieKey))
			http.Redirect(w, r, m.redirectTarget, http.StatusFound)
		}
		return
	}

	SetPrincipal(r, user)
	next(w, r)
}

func (m *AuthenticateMiddleware) isOptional(r *http.Request) bool {
	for _, p := range m.optionalForPaths {
		if strings.HasPrefix(r.URL.Path, p) || r.URL.Path == p {
			return true
		}
	}
	for _, m := range m.optionalForMethods {
		if r.Method == strings.ToUpper(m) {
			return true
		}
	}
	return false
}

func (m *AuthenticateMiddleware) tryGetUserByApiKeyHeader(r *http.Request) (*models.User, error) {
	key, err := utils.ExtractBearerAuth(r)
	if err != nil {
		return nil, err
	}

	var user *models.User
	userKey := strings.TrimSpace(key)
	user, err = m.userSrvc.GetUserByKey(userKey)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (m *AuthenticateMiddleware) tryGetUserByApiKeyQuery(r *http.Request) (*models.User, error) {
	key := r.URL.Query().Get(queryApiKey)
	var user *models.User
	userKey := strings.TrimSpace(key)
	if userKey == "" {
		return nil, errEmptyKey
	}
	user, err := m.userSrvc.GetUserByKey(userKey)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (m *AuthenticateMiddleware) tryGetUserByTrustedHeader(r *http.Request, create bool) (*models.User, error) {
	if !m.config.Security.TrustedHeaderAuth {
		return nil, errors.New("trusted header auth disabled")
	}
	create = create && m.config.Security.TrustedHeaderAuthAllowSignup // double-check

	remoteUser := r.Header.Get(m.config.Security.TrustedHeaderAuthKey)
	if remoteUser == "" {
		return nil, errors.New("trusted header field empty")
	}
	if addr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr); err != nil || !slice.ContainBy[net.IPNet](m.config.Security.TrustReverseProxyIPs(), func(ipNet net.IPNet) bool {
		return ipNet.Contains(addr.IP) // if err != nil, addr is nil
	}) {
		return nil, errors.New("reverse proxy not trusted")
	}

	user, err := m.userSrvc.GetUserById(remoteUser)
	if err == nil {
		return user, nil
	}

	if err.Error() != "record not found" || !create {
		return nil, err
	}

	// register new user solely based on upstream provided username (see https://github.com/muety/wakapi/issues/808)
	signup := &models.Signup{
		Username: remoteUser,
		Password: uuid.Must(uuid.NewV4()).String(), // throwaway random string as password
	}

	conf.Log().Request(r).Warn("registering new remotely authenticated user based on trusted header auth", "user_id", remoteUser)
	if _, _, err := m.userSrvc.CreateOrGet(signup, false); err != nil {
		return nil, err
	}
	return m.userSrvc.GetUserById(remoteUser)
}

func (m *AuthenticateMiddleware) tryGetUserByCookie(r *http.Request) (*models.User, error) {
	username, err := helpers.ExtractCookieAuth(r, m.config)
	if err != nil {
		return nil, err
	}

	user, err := m.userSrvc.GetUserById(*username)
	if err != nil {
		return nil, err
	}

	// no need to check password here, as securecookie decoding will fail anyway,
	// if cookie is not properly signed

	return user, nil
}
