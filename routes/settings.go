package routes

import (
	"fmt"
	"github.com/gorilla/schema"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"net/url"
)

type SettingsHandler struct {
	config   *models.Config
	userSrvc *services.UserService
}

var credentialsDecoder = schema.NewDecoder()

func NewSettingsHandler(userService *services.UserService) *SettingsHandler {
	return &SettingsHandler{
		config:   models.GetConfig(),
		userSrvc: userService,
	}
}

func (h *SettingsHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	data := map[string]interface{}{
		"User": user,
	}

	// TODO: when alerts are present, other data will not be passed to the template
	if handleAlerts(w, r, "settings.tpl.html") {
		return
	}
	templates["settings.tpl.html"].Execute(w, data)
}

func (h *SettingsHandler) PostCredentials(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)

	var credentials models.CredentialsReset
	if err := r.ParseForm(); err != nil {
		respondAlert(w, "missing parameters", "", "settings.tpl.html", http.StatusBadRequest)
		return
	}
	if err := credentialsDecoder.Decode(&credentials, r.PostForm); err != nil {
		respondAlert(w, "missing parameters", "", "settings.tpl.html", http.StatusBadRequest)
		return
	}

	if !utils.CheckPasswordBcrypt(user, credentials.PasswordOld, h.config.PasswordSalt) {
		respondAlert(w, "invalid credentials", "", "settings.tpl.html", http.StatusUnauthorized)
		return
	}

	if !credentials.IsValid() {
		respondAlert(w, "invalid parameters", "", "settings.tpl.html", http.StatusBadRequest)
		return
	}

	user.Password = credentials.PasswordNew
	if err := utils.HashPassword(user, h.config.PasswordSalt); err != nil {
		respondAlert(w, "internal server error", "", "settings.tpl.html", http.StatusInternalServerError)
		return
	}

	if _, err := h.userSrvc.Update(user); err != nil {
		respondAlert(w, "internal server error", "", "settings.tpl.html", http.StatusInternalServerError)
		return
	}

	login := &models.Login{
		Username: user.ID,
		Password: user.Password,
	}
	encoded, err := h.config.SecureCookie.Encode(models.AuthCookieKey, login)
	if err != nil {
		respondAlert(w, "internal server error", "", "settings.tpl.html", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     models.AuthCookieKey,
		Value:    encoded,
		Path:     "/",
		Secure:   !h.config.InsecureCookies,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	msg := url.QueryEscape("password was updated successfully")
	http.Redirect(w, r, fmt.Sprintf("%s/settings?success=%s", h.config.BasePath, msg), http.StatusFound)
}

func (h *SettingsHandler) PostResetApiKey(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	if _, err := h.userSrvc.ResetApiKey(user); err != nil {
		respondAlert(w, "internal server error", "", "settings.tpl.html", http.StatusInternalServerError)
		return
	}

	msg := url.QueryEscape(fmt.Sprintf("your new api key is: %s", user.ApiKey))
	http.Redirect(w, r, fmt.Sprintf("%s/settings?success=%s", h.config.BasePath, msg), http.StatusFound)
}

func (h *SettingsHandler) PostToggleBadges(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)

	if _, err := h.userSrvc.ToggleBadges(user); err != nil {
		respondAlert(w, "internal server error", "", "settings.tpl.html", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/settings", h.config.BasePath), http.StatusFound)
}
