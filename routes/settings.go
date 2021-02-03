package routes

import (
	"encoding/base64"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
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
	r := router.PathPrefix("/settings").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc, []string{}).Handler,
	)
	r.Methods(http.MethodGet).HandlerFunc(h.GetIndex)
	r.Methods(http.MethodPost).HandlerFunc(h.PostIndex)
}

func (h *SettingsHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r))
}

func (h *SettingsHandler) PostIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("missing form values"))
		return
	}

	action := r.PostForm.Get("action")
	r.PostForm.Del("action")

	actionFunc := h.dispatchAction(action)
	if actionFunc == nil {
		logbuch.Warn("failed to dispatch action '%s'", action)
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError("unknown action requests"))
		return
	}

	status, successMsg, errorMsg := actionFunc(w, r)

	// action responded itself
	if status == -1 {
		return
	}

	if errorMsg != "" {
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithError(errorMsg))
		return
	}
	if successMsg != "" {
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r).WithSuccess(successMsg))
		return
	}
	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r))
}

func (h *SettingsHandler) dispatchAction(action string) action {
	switch action {
	case "change_password":
		return h.actionChangePassword
	case "reset_apikey":
		return h.actionResetApiKey
	case "delete_alias":
		return h.actionDeleteAlias
	case "add_alias":
		return h.actionAddAlias
	case "delete_mapping":
		return h.actionDeleteLanguageMapping
	case "add_mapping":
		return h.actionAddLanguageMapping
	case "toggle_badges":
		return h.actionToggleBadges
	case "toggle_wakatime":
		return h.actionSetWakatimeApiKey
	case "regenerate_summaries":
		return h.actionRegenerateSummaries
	case "delete_account":
		return h.actionDeleteUser
	}
	return nil
}

func (h *SettingsHandler) actionChangePassword(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)

	var credentials models.CredentialsReset
	if err := r.ParseForm(); err != nil {
		return http.StatusBadRequest, "", "missing parameters"
	}
	if err := credentialsDecoder.Decode(&credentials, r.PostForm); err != nil {
		return http.StatusBadRequest, "", "missing parameters"
	}

	if !utils.CompareBcrypt(user.Password, credentials.PasswordOld, h.config.Security.PasswordSalt) {
		return http.StatusUnauthorized, "", "invalid credentials"
	}

	if !credentials.IsValid() {
		return http.StatusBadRequest, "", "invalid parameters"
	}

	user.Password = credentials.PasswordNew
	if hash, err := utils.HashBcrypt(user.Password, h.config.Security.PasswordSalt); err != nil {
		return http.StatusInternalServerError, "", "internal server error"
	} else {
		user.Password = hash
	}

	if _, err := h.userSrvc.Update(user); err != nil {
		return http.StatusInternalServerError, "", "internal server error"
	}

	login := &models.Login{
		Username: user.ID,
		Password: user.Password,
	}
	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, login.Username)
	if err != nil {
		return http.StatusInternalServerError, "", "internal server error"
	}

	http.SetCookie(w, h.config.CreateCookie(models.AuthCookieKey, encoded, "/"))
	return http.StatusOK, "password was updated successfully", ""
}

func (h *SettingsHandler) actionResetApiKey(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	if _, err := h.userSrvc.ResetApiKey(user); err != nil {
		return http.StatusInternalServerError, "", "internal server error"
	}

	msg := fmt.Sprintf("your new api key is: %s", user.ApiKey)
	return http.StatusOK, msg, ""
}

func (h *SettingsHandler) actionDeleteAlias(w http.ResponseWriter, r *http.Request) (int, string, string) {
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
		return http.StatusNotFound, "", "aliases not found"
	} else if err := h.aliasSrvc.DeleteMulti(aliases); err != nil {
		return http.StatusInternalServerError, "", "could not delete aliases"
	}

	return http.StatusOK, "aliases deleted successfully", ""
}

func (h *SettingsHandler) actionAddAlias(w http.ResponseWriter, r *http.Request) (int, string, string) {
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
		// TODO: distinguish between bad request, conflict and server error
		return http.StatusBadRequest, "", "invalid input"
	}

	return http.StatusOK, "alias added successfully", ""
}

func (h *SettingsHandler) actionDeleteLanguageMapping(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	id, err := strconv.Atoi(r.PostFormValue("mapping_id"))
	if err != nil {
		return http.StatusInternalServerError, "", "could not delete mapping"
	}

	if mapping, err := h.languageMappingSrvc.GetById(uint(id)); err != nil || mapping == nil {
		return http.StatusNotFound, "", "mapping not found"
	} else if mapping.UserID != user.ID {
		return http.StatusForbidden, "", "not allowed to delete mapping"
	}

	if err := h.languageMappingSrvc.Delete(&models.LanguageMapping{ID: uint(id)}); err != nil {
		return http.StatusInternalServerError, "", "could not delete mapping"
	}

	return http.StatusOK, "mapping deleted successfully", ""
}

func (h *SettingsHandler) actionAddLanguageMapping(w http.ResponseWriter, r *http.Request) (int, string, string) {
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
		return http.StatusConflict, "", "mapping already exists"
	}

	return http.StatusOK, "mapping added successfully", ""
}

func (h *SettingsHandler) actionToggleBadges(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	if _, err := h.userSrvc.ToggleBadges(user); err != nil {
		return http.StatusInternalServerError, "", "internal server error"
	}

	return http.StatusOK, "", ""
}

func (h *SettingsHandler) actionSetWakatimeApiKey(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	apiKey := r.PostFormValue("api_key")

	// Healthcheck, if a new API key is set, i.e. the feature is activated
	if (user.WakatimeApiKey == "" && apiKey != "") && !h.validateWakatimeKey(apiKey) {
		return http.StatusBadRequest, "", "failed to connect to WakaTime, API key invalid?"
	}

	if _, err := h.userSrvc.SetWakatimeApiKey(user, apiKey); err != nil {
		return http.StatusInternalServerError, "", "internal server error"
	}

	return http.StatusOK, "Wakatime API Key updated successfully", ""
}

func (h *SettingsHandler) actionRegenerateSummaries(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)

	logbuch.Info("clearing summaries for user '%s'", user.ID)
	if err := h.summarySrvc.DeleteByUser(user.ID); err != nil {
		logbuch.Error("failed to clear summaries: %v", err)
		return http.StatusInternalServerError, "", "failed to delete old summaries"
	}

	if err := h.aggregationSrvc.Run(map[string]bool{user.ID: true}); err != nil {
		logbuch.Error("failed to regenerate summaries: %v", err)
		return http.StatusInternalServerError, "", "failed to generate aggregations"

	}

	return http.StatusOK, "summaries are being regenerated – this may take a few seconds", ""
}

func (h *SettingsHandler) actionDeleteUser(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	go func(user *models.User) {
		logbuch.Info("deleting user '%s' shortly", user.ID)
		time.Sleep(5 * time.Minute)
		if err := h.userSrvc.Delete(user); err != nil {
			logbuch.Error("failed to delete user '%s' – %v", user.ID, err)
		} else {
			logbuch.Info("successfully deleted user '%s'", user.ID)
		}
	}(user)

	http.SetCookie(w, h.config.GetClearCookie(models.AuthCookieKey, "/"))
	http.Redirect(w, r, fmt.Sprintf("%s/?success=%s", h.config.Server.BasePath, "Your account will be deleted in a few minutes. Sorry to you go."), http.StatusFound)
	return -1, "", ""
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
