package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/emvi/logbuch"
	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/services"
)

type OAuthHandler struct {
	config   *conf.Config
	userSrvc services.IUserService
}

func NewOAuthHandler(userService services.IUserService) *OAuthHandler {
	return &OAuthHandler{
		userSrvc: userService,
		config:   conf.Get(),
	}
}

func (h *OAuthHandler) RegisterRoutes(router chi.Router) {
	if !h.config.OAuth.Enabled {
		return
	}

	logbuch.Info("exposing oauth routes at /api/oauth and /api/oath_callback")

	router.Get("/oauth", h.oauthRedirect)
	router.Get("/oauth_callback", h.oauthCallback)
}

func (h *OAuthHandler) oauthRedirect(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action") // login, bind
	if action == "" {
		conf.Log().Request(r).Error("missing action query param")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(conf.ErrBadRequest))
		return
	}

	var r_url string
	provider := h.config.OAuth.Provider
	clientId := h.config.OAuth.ClientId
	urlValues := url.Values{}
	redirectUri := r.Host + "/api/oauth_callback" + "?action=" + action

	urlValues.Add("response_type", "code")
	urlValues.Add("client_id", clientId)
	urlValues.Add("redirect_uri", redirectUri)

	switch provider {
	// TODO: add OIDC support
	case "github":
		r_url = "https://github.com/login/oauth/authorize?"
		urlValues.Add("scope", "read:user")
	case "microsoft":
		r_url = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize?"
		urlValues.Add("scope", "user.read")
		urlValues.Add("response_mode", "query")
	case "google":
		r_url = "https://accounts.google.com/o/oauth2/v2/auth?"
		urlValues.Add("scope", "https://www.googleapis.com/auth/userinfo.profile")
	default:
		conf.Log().Request(r).Error("invalid oauth provider")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(conf.ErrBadRequest))
		return
	}
	w.Header().Set("Location", r_url+urlValues.Encode())
	w.WriteHeader(http.StatusFound)
}

func (h *OAuthHandler) oauthCallback(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	if action == "bind" || action == "login" {
		provider := h.config.OAuth.Provider
		clientId := h.config.OAuth.ClientId
		clientSecret := h.config.OAuth.ClientSecret
		var url1, url2, additionalbody, scope, authstring, idstring string

		switch provider {
		case "github":
			url1 = "https://github.com/login/oauth/access_token"
			url2 = "https://api.github.com/user"
			additionalbody = ""
			authstring = "code"
			scope = "read:user"
			idstring = "id"
		case "microsoft":
			url1 = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
			url2 = "https://graph.microsoft.com/v1.0/me"
			additionalbody = "&grant_type=authorization_code"
			scope = "user.read"
			authstring = "code"
			idstring = "id"
		case "google":
			url1 = "https://oauth2.googleapis.com/token"
			url2 = "https://www.googleapis.com/oauth2/v1/userinfo"
			additionalbody = "&grant_type=authorization_code"
			scope = "https://www.googleapis.com/auth/userinfo.profile"
			authstring = "code"
			idstring = "id"
		default:
			conf.Log().Request(r).Error("invalid oauth provider")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(conf.ErrBadRequest))
			return
		}

		code := r.URL.Query().Get(authstring)
		if code == "" {
			conf.Log().Request(r).Error("missing oauth code")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(conf.ErrBadRequest))
			return
		}

		body := "client_id=" + clientId + "&client_secret=" + clientSecret + "&code=" + code + "&redirect_uri=" + r.Host + "/api/oauth_callback" + "&scope=" + scope + additionalbody
		req, _ := http.NewRequest(http.MethodPost, url1, strings.NewReader(body))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Accept", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			conf.Log().Request(r).Error("error while requesting access token")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(conf.ErrInternalServerError))
			return
		}
		defer res.Body.Close()

		// response body differs between providers
		var data map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			conf.Log().Request(r).Error("error while decoding access token response")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(conf.ErrInternalServerError))
			return
		}

		accessToken := data["access_token"].(string)
		req, _ = http.NewRequest(http.MethodGet, url2, nil)
		req.Header.Add("Authorization", "Bearer "+accessToken)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			conf.Log().Request(r).Error("error while requesting user data")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(conf.ErrInternalServerError))
			return
		}
		defer res.Body.Close()

		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			conf.Log().Request(r).Error("error while decoding user data response")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(conf.ErrInternalServerError))
			return
		}

		UserID := data[idstring].(string)
		if UserID == "0" || UserID == "" {
			conf.Log().Request(r).Error("error while getting user id")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(conf.ErrInternalServerError))
			return
		}

		if action == "bind" {
			// TODO: bind user
			conf.Log().Request(r).Info("binding user " + UserID)
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte("not implemented"))
			return
		} else if action == "login" {
			// TODO: login user
			conf.Log().Request(r).Info("logging in user " + UserID)
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte("not implemented"))
			return
		}

	} else {
		conf.Log().Request(r).Error("invalid action")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(conf.ErrBadRequest))
		return
	}
}
