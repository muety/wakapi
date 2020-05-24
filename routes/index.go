package routes

import (
	"github.com/gorilla/schema"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
)

type IndexHandler struct {
	config   *models.Config
	userSrvc *services.UserService
}

var loginDecoder = schema.NewDecoder()

func NewIndexHandler(config *models.Config, userService *services.UserService) *IndexHandler {
	return &IndexHandler{
		config:   config,
		userSrvc: userService,
	}
}

func (h *IndexHandler) Index(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, "/summary", http.StatusFound)
		return
	}

	if err := r.URL.Query().Get("error"); err != "" {
		if err == "unauthorized" {
			respondError(w, err, http.StatusUnauthorized)
		} else {
			respondError(w, err, http.StatusInternalServerError)
		}
		return
	}

	templates["index.tpl.html"].Execute(w, nil)
}

func (h *IndexHandler) Login(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	var login models.Login
	if err := r.ParseForm(); err != nil {
		respondError(w, "missing parameters", http.StatusBadRequest)
		return
	}
	if err := loginDecoder.Decode(&login, r.PostForm); err != nil {
		respondError(w, "missing parameters", http.StatusBadRequest)
		return
	}

	user, err := h.userSrvc.GetUserById(login.Username)
	if err != nil {
		respondError(w, "resource not found", http.StatusNotFound)
		return
	}

	if !utils.CheckPassword(user, login.Password) {
		respondError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	encoded, err := h.config.SecureCookie.Encode(models.AuthCookieKey, login)
	if err != nil {
		respondError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     models.AuthCookieKey,
		Value:    encoded,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/summary", http.StatusFound)
}

func (h *IndexHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	utils.ClearCookie(w, models.AuthCookieKey)
	http.Redirect(w, r, "/", http.StatusFound)
}

func respondError(w http.ResponseWriter, error string, status int) {
	w.WriteHeader(status)
	templates["index.tpl.html"].Execute(w, struct {
		Error string
	}{Error: error})
}
