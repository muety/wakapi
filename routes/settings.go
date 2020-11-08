package routes

import (
	"fmt"
	"github.com/gorilla/schema"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"log"
	"net/http"
	"strconv"
)

type SettingsHandler struct {
	config              *conf.Config
	userSrvc            services.IUserService
	summarySrvc         services.ISummaryService
	aggregationSrvc     services.IAggregationService
	languageMappingSrvc services.ILanguageMappingService
}

var credentialsDecoder = schema.NewDecoder()

func NewSettingsHandler(userService services.IUserService, summaryService services.ISummaryService, aggregationService services.IAggregationService, languageMappingService services.ILanguageMappingService) *SettingsHandler {
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

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r))
}

func (h *SettingsHandler) PostCredentials(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)

	var credentials models.CredentialsReset
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("missing parameters"))
		return
	}
	if err := credentialsDecoder.Decode(&credentials, r.PostForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("missing parameters"))
		return
	}

	if !utils.CompareBcrypt(user.Password, credentials.PasswordOld, h.config.Security.PasswordSalt) {
		w.WriteHeader(http.StatusUnauthorized)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("invalid credentials"))
		return
	}

	if !credentials.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("invalid parameters"))
		return
	}

	user.Password = credentials.PasswordNew
	if hash, err := utils.HashBcrypt(user.Password, h.config.Security.PasswordSalt); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("internal server error"))
		return
	} else {
		user.Password = hash
	}

	if _, err := h.userSrvc.Update(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("internal server error"))
		return
	}

	login := &models.Login{
		Username: user.ID,
		Password: user.Password,
	}
	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("internal server error"))
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

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithSuccess("password was updated successfully"))
}

func (h *SettingsHandler) DeleteLanguageMapping(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	id, err := strconv.Atoi(r.PostFormValue("mapping_id"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("could not delete mapping"))
		return
	}

	mapping := &models.LanguageMapping{
		ID:     uint(id),
		UserID: user.ID,
	}

	err = h.languageMappingSrvc.Delete(mapping)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("could not delete mapping"))
		return
	}

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithSuccess("mapping deleted successfully"))
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
		w.WriteHeader(http.StatusConflict)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("mapping already exists"))
		return
	}

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithSuccess("mapping added successfully"))
}

func (h *SettingsHandler) PostResetApiKey(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	if _, err := h.userSrvc.ResetApiKey(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("internal server error"))
		return
	}

	msg := fmt.Sprintf("your new api key is: %s", user.ApiKey)
	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithSuccess(msg))
}

func (h *SettingsHandler) PostToggleBadges(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	if _, err := h.userSrvc.ToggleBadges(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("internal server error"))
		return
	}

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r))
}

func (h *SettingsHandler) PostRegenerateSummaries(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)

	log.Printf("clearing summaries for user '%s'\n", user.ID)
	if err := h.summarySrvc.DeleteByUser(user.ID); err != nil {
		log.Printf("failed to clear summaries: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("failed to delete old summaries"))
		return
	}

	if err := h.aggregationSrvc.Run(map[string]bool{user.ID: true}); err != nil {
		log.Printf("failed to regenerate summaries: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("failed to generate aggregations"))
		return
	}

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithSuccess("summaries are being regenerated – this may take a few seconds"))
}

func (h *SettingsHandler) buildViewModel(r *http.Request) *view.SettingsViewModel {
	user := r.Context().Value(models.UserKey).(*models.User)
	mappings, _ := h.languageMappingSrvc.GetByUser(user.ID)
	return &view.SettingsViewModel{
		User:             user,
		LanguageMappings: mappings,
		Success:          r.URL.Query().Get("success"),
		Error:            r.URL.Query().Get("error"),
	}
}
