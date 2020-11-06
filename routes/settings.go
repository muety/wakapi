package routes

import (
	"fmt"
	"github.com/gorilla/schema"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type SettingsHandler struct {
	config              *conf.Config
	userSrvc            *services.UserService
	summarySrvc         *services.SummaryService
	aggregationSrvc     *services.AggregationService
	languageMappingSrvc *services.LanguageMappingService
}

var credentialsDecoder = schema.NewDecoder()

func NewSettingsHandler(userService *services.UserService, summaryService *services.SummaryService, aggregationService *services.AggregationService, languageMappingService *services.LanguageMappingService) *SettingsHandler {
	return &SettingsHandler{
		config:              conf.Get(),
		summarySrvc:         summaryService,
		aggregationSrvc:     aggregationService,
		languageMappingSrvc: languageMappingService,
		userSrvc:            userService,
	}
}

func (h *SettingsHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	mappings, _ := h.languageMappingSrvc.GetByUser(user.ID)
	data := map[string]interface{}{
		"User":             user,
		"LanguageMappings": mappings,
		"Success":          r.FormValue("success"),
		"Error":            r.FormValue("error"),
	}

	templates[conf.SettingsTemplate].Execute(w, data)
}

func (h *SettingsHandler) PostCredentials(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)

	var credentials models.CredentialsReset
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("missing parameters")), http.StatusFound)
		return
	}
	if err := credentialsDecoder.Decode(&credentials, r.PostForm); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("missing parameters")), http.StatusFound)
		return
	}

	if !utils.CompareBcrypt(user.Password, credentials.PasswordOld, h.config.Security.PasswordSalt) {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("invalid credentials")), http.StatusFound)
		return
	}

	if !credentials.IsValid() {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("invalid parameters")), http.StatusFound)
		return
	}

	user.Password = credentials.PasswordNew
	if hash, err := utils.HashBcrypt(user.Password, h.config.Security.PasswordSalt); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("internal server error")), http.StatusFound)
		return
	} else {
		user.Password = hash
	}

	if _, err := h.userSrvc.Update(user); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("internal server error")), http.StatusFound)
		return
	}

	login := &models.Login{
		Username: user.ID,
		Password: user.Password,
	}
	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, login)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("internal server error")), http.StatusFound)
		return
	}

	cookie := &http.Cookie{
		Name:     models.AuthCookieKey,
		Value:    encoded,
		Path:     "/",
		Secure:   !h.config.Security.InsecureCookies,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, fmt.Sprintf("%s/settings?success=%s", h.config.Server.BasePath, url.QueryEscape("password was updated successfully")), http.StatusFound)
}

func (h *SettingsHandler) DeleteLanguageMapping(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	id, err := strconv.Atoi(r.PostFormValue("mapping_id"))
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("could not delete mapping")), http.StatusFound)
		return
	}

	mapping := &models.LanguageMapping{
		ID:     uint(id),
		UserID: user.ID,
	}

	err = h.languageMappingSrvc.Delete(mapping)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("could not delete mapping")), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/settings?success=%s", h.config.Server.BasePath, url.QueryEscape("mapping deleted successfully")), http.StatusFound)
}

func (h *SettingsHandler) PostLanguageMapping(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
	user := r.Context().Value(models.UserKey).(*models.User)
	extension := r.PostFormValue("extension")
	language := r.PostFormValue("language")

	if extension[0] == '.' {
		extension = extension[1:]
	}

	mapping := &models.LanguageMapping{
		UserID:    user.ID,
		Extension: extension,
		Language:  language,
	}

	if _, err := h.languageMappingSrvc.Create(mapping); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("mapping already exists")), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/settings?success=%s", h.config.Server.BasePath, url.QueryEscape("mapping added successfully")), http.StatusFound)
}

func (h *SettingsHandler) PostResetApiKey(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	if _, err := h.userSrvc.ResetApiKey(user); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("internal server error")), http.StatusFound)
		return
	}

	msg := url.QueryEscape(fmt.Sprintf("your new api key is: %s", user.ApiKey))
	http.Redirect(w, r, fmt.Sprintf("%s/settings?success=%s", h.config.Server.BasePath, msg), http.StatusFound)
}

func (h *SettingsHandler) PostToggleBadges(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	if _, err := h.userSrvc.ToggleBadges(user); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("internal server error")), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/settings", h.config.Server.BasePath), http.StatusFound)
}

func (h *SettingsHandler) PostRegenerateSummaries(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)

	log.Printf("clearing summaries for user '%s'\n", user.ID)
	if err := h.summarySrvc.DeleteByUser(user.ID); err != nil {
		log.Printf("failed to clear summaries: %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("failed to delete old summaries")), http.StatusFound)
		return
	}

	if err := h.aggregationSrvc.Run(map[string]bool{user.ID: true}); err != nil {
		log.Printf("failed to regenerate summaries: %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("%s/settings?error=%s", h.config.Server.BasePath, url.QueryEscape("failed to generate aggregations")), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/settings?success=%s", h.config.Server.BasePath, url.QueryEscape("summaries are being regenerated – this may take a few second")), http.StatusFound)
}
