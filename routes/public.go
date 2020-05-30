package routes

import (
	"fmt"
	"github.com/gorilla/schema"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"net/url"
	"time"
)

type IndexHandler struct {
	config       *models.Config
	userSrvc     *services.UserService
	keyValueSrvc *services.KeyValueService
}

var loginDecoder = schema.NewDecoder()
var signupDecoder = schema.NewDecoder()

func NewIndexHandler(userService *services.UserService, keyValueService *services.KeyValueService) *IndexHandler {
	return &IndexHandler{
		config:       models.GetConfig(),
		userSrvc:     userService,
		keyValueSrvc: keyValueService,
	}
}

func (h *IndexHandler) Index(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.BasePath), http.StatusFound)
		return
	}

	if handleAlerts(w, r, "") {
		return
	}

	// TODO: make this more generic and reusable
	if success := r.URL.Query().Get("success"); success != "" {
		templates["index.tpl.html"].Execute(w, struct {
			Success string
			Error   string
		}{Success: success})
		return
	}
	templates["index.tpl.html"].Execute(w, nil)
}

func (h *IndexHandler) Imprint(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	text := "failed to load content"
	if data, err := h.keyValueSrvc.GetString(models.ImprintKey); err == nil {
		text = data.Value
	}

	templates["imprint.tpl.html"].Execute(w, &struct {
		HtmlText string
	}{HtmlText: text})
}

func (h *IndexHandler) Login(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.BasePath), http.StatusFound)
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
	if !middlewares.CheckAndMigratePassword(user, &login, h.config.PasswordSalt, h.userSrvc) {
		respondAlert(w, "invalid credentials", "", "", http.StatusUnauthorized)
		return
	}

	encoded, err := h.config.SecureCookie.Encode(models.AuthCookieKey, login)
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
		Secure:   !h.config.InsecureCookies,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.BasePath), http.StatusFound)
}

func (h *IndexHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	utils.ClearCookie(w, models.AuthCookieKey, !h.config.InsecureCookies)
	http.Redirect(w, r, fmt.Sprintf("%s/", h.config.BasePath), http.StatusFound)
}

func (h *IndexHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.BasePath), http.StatusFound)
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.handlePostSignup(w, r)
		return
	default:
		h.handleGetSignup(w, r)
		return
	}
}

func (h *IndexHandler) handleGetSignup(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.BasePath), http.StatusFound)
		return
	}

	if handleAlerts(w, r, "signup.tpl.html") {
		return
	}

	templates["signup.tpl.html"].Execute(w, nil)
}

func (h *IndexHandler) handlePostSignup(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.BasePath), http.StatusFound)
		return
	}

	var signup models.Signup
	if err := r.ParseForm(); err != nil {
		respondAlert(w, "missing parameters", "", "signup.tpl.html", http.StatusBadRequest)
		return
	}
	if err := signupDecoder.Decode(&signup, r.PostForm); err != nil {
		respondAlert(w, "missing parameters", "", "signup.tpl.html", http.StatusBadRequest)
		return
	}

	if !signup.IsValid() {
		respondAlert(w, "invalid parameters", "", "signup.tpl.html", http.StatusBadRequest)
		return
	}

	_, created, err := h.userSrvc.CreateOrGet(&signup)
	if err != nil {
		respondAlert(w, "failed to create new user", "", "signup.tpl.html", http.StatusInternalServerError)
		return
	}
	if !created {
		respondAlert(w, "user already existing", "", "signup.tpl.html", http.StatusConflict)
		return
	}

	msg := url.QueryEscape("account created successfully")
	http.Redirect(w, r, fmt.Sprintf("%s/?success=%s", h.config.BasePath, msg), http.StatusFound)
}
