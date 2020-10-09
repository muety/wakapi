package routes

import (
	"fmt"
	"github.com/gorilla/schema"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"net/url"
	"time"
)

type IndexHandler struct {
	config       *conf.Config
	userSrvc     *services.UserService
	keyValueSrvc *services.KeyValueService
}

var loginDecoder = schema.NewDecoder()
var signupDecoder = schema.NewDecoder()

func NewIndexHandler(userService *services.UserService, keyValueService *services.KeyValueService) *IndexHandler {
	return &IndexHandler{
		config:       conf.Get(),
		userSrvc:     userService,
		keyValueSrvc: keyValueService,
	}
}

func (h *IndexHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	if handleAlerts(w, r, "") {
		return
	}

	templates[conf.IndexTemplate].Execute(w, nil)
}

func (h *IndexHandler) GetImprint(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	text := "failed to load content"
	if data, err := h.keyValueSrvc.GetString(models.ImprintKey); err == nil {
		text = data.Value
	}

	templates[conf.ImprintTemplate].Execute(w, &struct {
		HtmlText string
	}{HtmlText: text})
}

func (h *IndexHandler) PostLogin(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	var login models.Login
	if err := r.ParseForm(); err != nil {
		respondAlert(w, "missing parameters", "", "", http.StatusBadRequest)
		return
	}
	if err := loginDecoder.Decode(&login, r.PostForm); err != nil {
		respondAlert(w, "missing parameters", "", "", http.StatusBadRequest)
		return
	}

	user, err := h.userSrvc.GetUserById(login.Username)
	if err != nil {
		respondAlert(w, "resource not found", "", "", http.StatusNotFound)
		return
	}

	// TODO: depending on middleware package here is a hack
	if !middlewares.CheckAndMigratePassword(user, &login, h.config.Security.PasswordSalt, h.userSrvc) {
		respondAlert(w, "invalid credentials", "", "", http.StatusUnauthorized)
		return
	}

	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, login)
	if err != nil {
		respondAlert(w, "internal server error", "", "", http.StatusInternalServerError)
		return
	}

	user.LastLoggedInAt = models.CustomTime(time.Now())
	h.userSrvc.Update(user)

	cookie := &http.Cookie{
		Name:     models.AuthCookieKey,
		Value:    encoded,
		Path:     "/",
		Secure:   !h.config.Security.InsecureCookies,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
}

func (h *IndexHandler) PostLogout(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	utils.ClearCookie(w, models.AuthCookieKey, !h.config.Security.InsecureCookies)
	http.Redirect(w, r, fmt.Sprintf("%s/", h.config.Server.BasePath), http.StatusFound)
}

func (h *IndexHandler) GetSignup(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	if handleAlerts(w, r, conf.SignupTemplate) {
		return
	}

	templates[conf.SignupTemplate].Execute(w, nil)
}

func (h *IndexHandler) PostSignup(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	var signup models.Signup
	if err := r.ParseForm(); err != nil {
		respondAlert(w, "missing parameters", "", conf.SignupTemplate, http.StatusBadRequest)
		return
	}
	if err := signupDecoder.Decode(&signup, r.PostForm); err != nil {
		respondAlert(w, "missing parameters", "", conf.SignupTemplate, http.StatusBadRequest)
		return
	}

	if !signup.IsValid() {
		respondAlert(w, "invalid parameters", "", conf.SignupTemplate, http.StatusBadRequest)
		return
	}

	_, created, err := h.userSrvc.CreateOrGet(&signup)
	if err != nil {
		respondAlert(w, "failed to create new user", "", conf.SignupTemplate, http.StatusInternalServerError)
		return
	}
	if !created {
		respondAlert(w, "user already existing", "", conf.SignupTemplate, http.StatusConflict)
		return
	}

	msg := url.QueryEscape("account created successfully")
	http.Redirect(w, r, fmt.Sprintf("%s/?success=%s", h.config.Server.BasePath, msg), http.StatusFound)
}
