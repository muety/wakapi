package routes

import (
	"fmt"
	"github.com/gorilla/schema"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
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

	templates["settings.tpl.html"].Execute(w, nil)
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
	http.Redirect(w, r, fmt.Sprintf("%s/settings", h.config.BasePath), http.StatusFound)
}

func (h *SettingsHandler) PostResetApiKey(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
}
