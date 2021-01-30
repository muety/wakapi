package routes

import (
	"encoding/base64"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"strconv"
	"time"
)

type SettingsHandler struct {
	config              *conf.Config
	userSrvc            services.IUserService
	summarySrvc         services.ISummaryService
	aliasSrvc           services.IAliasService
	aggregationSrvc     services.IAggregationService
	languageMappingSrvc services.ILanguageMappingService
	httpClient          *http.Client
}

var credentialsDecoder = schema.NewDecoder()

func NewSettingsHandler(userService services.IUserService, summaryService services.ISummaryService, aliasService services.IAliasService, aggregationService services.IAggregationService, languageMappingService services.ILanguageMappingService) *SettingsHandler {
	return &SettingsHandler{
		config:              conf.Get(),
		summarySrvc:         summaryService,
		aliasSrvc:           aliasService,
		aggregationSrvc:     aggregationService,
		languageMappingSrvc: languageMappingService,
		userSrvc:            userService,
		httpClient:          &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *SettingsHandler) RegisterRoutes(router *mux.Router) {
	router.Methods(http.MethodGet).HandlerFunc(h.GetIndex)
	router.Path("/credentials").Methods(http.MethodPost).HandlerFunc(h.PostCredentials)
	router.Path("/aliases").Methods(http.MethodPost).HandlerFunc(h.PostAlias)
	router.Path("/aliases/delete").Methods(http.MethodPost).HandlerFunc(h.DeleteAlias)
	router.Path("/language_mappings").Methods(http.MethodPost).HandlerFunc(h.PostLanguageMapping)
	router.Path("/language_mappings/delete").Methods(http.MethodPost).HandlerFunc(h.DeleteLanguageMapping)
	router.Path("/reset").Methods(http.MethodPost).HandlerFunc(h.PostResetApiKey)
	router.Path("/badges").Methods(http.MethodPost).HandlerFunc(h.PostToggleBadges)
	router.Path("/wakatime_integration").Methods(http.MethodPost).HandlerFunc(h.PostSetWakatimeApiKey)
	router.Path("/regenerate").Methods(http.MethodPost).HandlerFunc(h.PostRegenerateSummaries)
}

func (h *SettingsHandler) RegisterAPIRoutes(router *mux.Router) {}

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

	http.SetCookie(w, h.config.CreateCookie(models.AuthCookieKey, encoded, "/"))
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

	if mapping, err := h.languageMappingSrvc.GetById(uint(id)); err != nil || mapping == nil {
		w.WriteHeader(http.StatusNotFound)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("mapping not found"))
		return
	} else if mapping.UserID != user.ID {
		w.WriteHeader(http.StatusForbidden)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("not allowed to delete mapping"))
		return
	}

	if err := h.languageMappingSrvc.Delete(&models.LanguageMapping{ID: uint(id)}); err != nil {
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

func (h *SettingsHandler) DeleteAlias(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	aliasKey := r.PostFormValue("key")
	aliasType, err := strconv.Atoi(r.PostFormValue("type"))
	if err != nil {
		aliasType = 99 // nothing will be found later on
	}

	if aliases, err := h.aliasSrvc.GetByUserAndKeyAndType(user.ID, aliasKey, uint8(aliasType)); err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("aliases not found"))
		return
	} else if err := h.aliasSrvc.DeleteMulti(aliases); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("could not delete aliases"))
		return
	}

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithSuccess("aliases deleted successfully"))
}

func (h *SettingsHandler) PostAlias(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
	user := r.Context().Value(models.UserKey).(*models.User)
	aliasKey := r.PostFormValue("key")
	aliasValue := r.PostFormValue("value")
	aliasType, err := strconv.Atoi(r.PostFormValue("type"))
	if err != nil {
		aliasType = 99 // Alias.IsValid() will return false later on
	}

	alias := &models.Alias{
		UserID: user.ID,
		Key:    aliasKey,
		Value:  aliasValue,
		Type:   uint8(aliasType),
	}

	if _, err := h.aliasSrvc.Create(alias); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		// TODO: distinguish between bad request, conflict and server error
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("invalid input"))
		return
	}

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithSuccess("alias added successfully"))
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

func (h *SettingsHandler) PostSetWakatimeApiKey(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	apiKey := r.PostFormValue("api_key")

	// Healthcheck, if a new API key is set, i.e. the feature is activated
	if (user.WakatimeApiKey == "" && apiKey != "") && !h.validateWakatimeKey(apiKey) {
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("failed to connect to WakaTime, API key invalid?"))
		return
	}

	if _, err := h.userSrvc.SetWakatimeApiKey(user, apiKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("internal server error"))
		return
	}

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithSuccess("Wakatime API Key updated successfully"))
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

	logbuch.Info("clearing summaries for user '%s'", user.ID)
	if err := h.summarySrvc.DeleteByUser(user.ID); err != nil {
		logbuch.Error("failed to clear summaries: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("failed to delete old summaries"))
		return
	}

	if err := h.aggregationSrvc.Run(map[string]bool{user.ID: true}); err != nil {
		logbuch.Error("failed to regenerate summaries: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("failed to generate aggregations"))
		return
	}

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithSuccess("summaries are being regenerated – this may take a few seconds"))
}

func (h *SettingsHandler) validateWakatimeKey(apiKey string) bool {
	headers := http.Header{
		"Accept": []string{"application/json"},
		"Authorization": []string{
			fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(apiKey))),
		},
	}

	request, err := http.NewRequest(
		http.MethodGet,
		conf.WakatimeApiUrl+conf.WakatimeApiUserEndpoint,
		nil,
	)
	if err != nil {
		return false
	}

	request.Header = headers

	response, err := h.httpClient.Do(request)
	if err != nil || response.StatusCode < 200 || response.StatusCode >= 300 {
		return false
	}

	return true
}

func (h *SettingsHandler) buildViewModel(r *http.Request) *view.SettingsViewModel {
	user := r.Context().Value(models.UserKey).(*models.User)
	mappings, _ := h.languageMappingSrvc.GetByUser(user.ID)
	aliases, _ := h.aliasSrvc.GetByUser(user.ID)
	aliasMap := make(map[string][]*models.Alias)
	for _, a := range aliases {
		k := fmt.Sprintf("%s_%d", a.Key, a.Type)
		if _, ok := aliasMap[k]; !ok {
			aliasMap[k] = []*models.Alias{a}
		} else {
			aliasMap[k] = append(aliasMap[k], a)
		}
	}

	combinedAliases := make([]*view.SettingsVMCombinedAlias, 0)
	for _, l := range aliasMap {
		ca := &view.SettingsVMCombinedAlias{
			Key:    l[0].Key,
			Type:   l[0].Type,
			Values: make([]string, len(l)),
		}
		for i, a := range l {
			ca.Values[i] = a.Value
		}
		combinedAliases = append(combinedAliases, ca)
	}

	return &view.SettingsViewModel{
		User:             user,
		LanguageMappings: mappings,
		Aliases:          combinedAliases,
		Success:          r.URL.Query().Get("success"),
		Error:            r.URL.Query().Get("error"),
	}
}
