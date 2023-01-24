package routes

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/services/imports"
	"github.com/muety/wakapi/utils"
)

const criticalError = "a critical error has occurred, sorry"

type SettingsHandler struct {
	config              *conf.Config
	userSrvc            services.IUserService
	summarySrvc         services.ISummaryService
	heartbeatSrvc       services.IHeartbeatService
	aliasSrvc           services.IAliasService
	aggregationSrvc     services.IAggregationService
	languageMappingSrvc services.ILanguageMappingService
	projectLabelSrvc    services.IProjectLabelService
	keyValueSrvc        services.IKeyValueService
	mailSrvc            services.IMailService
	httpClient          *http.Client
}

var credentialsDecoder = schema.NewDecoder()

func NewSettingsHandler(
	userService services.IUserService,
	heartbeatService services.IHeartbeatService,
	summaryService services.ISummaryService,
	aliasService services.IAliasService,
	aggregationService services.IAggregationService,
	languageMappingService services.ILanguageMappingService,
	projectLabelService services.IProjectLabelService,
	keyValueService services.IKeyValueService,
	mailService services.IMailService,
) *SettingsHandler {
	return &SettingsHandler{
		config:              conf.Get(),
		summarySrvc:         summaryService,
		aliasSrvc:           aliasService,
		aggregationSrvc:     aggregationService,
		languageMappingSrvc: languageMappingService,
		projectLabelSrvc:    projectLabelService,
		userSrvc:            userService,
		heartbeatSrvc:       heartbeatService,
		keyValueSrvc:        keyValueService,
		mailSrvc:            mailService,
		httpClient:          &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *SettingsHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/settings").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).
			WithRedirectTarget(defaultErrorRedirectTarget()).
			WithRedirectErrorMessage("unauthorized").
			Handler,
	)
	r.Methods(http.MethodGet).HandlerFunc(h.GetIndex)
	r.Methods(http.MethodPost).HandlerFunc(h.PostIndex)
}

func (h *SettingsHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w))
}

func (h *SettingsHandler) PostIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w).WithError("missing form values"))
		return
	}

	action := r.PostForm.Get("action")
	r.PostForm.Del("action")

	actionFunc := h.dispatchAction(action)
	if actionFunc == nil {
		logbuch.Warn("failed to dispatch action '%s'", action)
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w).WithError("unknown action requests"))
		return
	}

	status, successMsg, errorMsg := actionFunc(w, r)

	// action responded itself
	if status == -1 {
		return
	}

	if errorMsg != "" {
		w.WriteHeader(status)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w).WithError(errorMsg))
		return
	}
	if successMsg != "" {
		w.WriteHeader(status)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w).WithSuccess(successMsg))
		return
	}
	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w))
}

func (h *SettingsHandler) dispatchAction(action string) action {
	switch action {
	case "change_password":
		return h.actionChangePassword
	case "update_user":
		return h.actionUpdateUser
	case "reset_apikey":
		return h.actionResetApiKey
	case "delete_alias":
		return h.actionDeleteAlias
	case "add_alias":
		return h.actionAddAlias
	case "add_project_to_label":
		return h.addProjectToLabel
	case "add_label":
		return h.actionAddLabel
	case "delete_label":
		return h.actionDeleteLabel
	case "delete_mapping":
		return h.actionDeleteLanguageMapping
	case "add_mapping":
		return h.actionAddLanguageMapping
	case "update_sharing":
		return h.actionUpdateSharing
	case "update_leaderboard":
		return h.actionUpdateLeaderboard
	case "toggle_wakatime":
		return h.actionSetWakatimeApiKey
	case "import_wakatime":
		return h.actionImportWakatime
	case "regenerate_summaries":
		return h.actionRegenerateSummaries
	case "clear_data":
		return h.actionClearData
	case "delete_account":
		return h.actionDeleteUser
	}
	return nil
}

func (h *SettingsHandler) actionUpdateUser(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)

	var payload models.UserDataUpdate
	if err := r.ParseForm(); err != nil {
		return http.StatusBadRequest, "", "missing parameters"
	}
	if err := credentialsDecoder.Decode(&payload, r.PostForm); err != nil {
		return http.StatusBadRequest, "", "missing parameters"
	}

	if !payload.IsValid() {
		return http.StatusBadRequest, "", "invalid parameters - perhaps invalid e-mail address?"
	}

	if payload.Email == "" && user.HasActiveSubscription() {
		return http.StatusBadRequest, "", "cannot unset email while subscription is active"
	}

	user.Email = payload.Email
	user.Location = payload.Location
	user.ReportsWeekly = payload.ReportsWeekly
	user.PublicLeaderboard = payload.PublicLeaderboard

	if _, err := h.userSrvc.Update(user); err != nil {
		return http.StatusInternalServerError, "", conf.ErrInternalServerError
	}

	return http.StatusOK, "user updated successfully", ""
}

func (h *SettingsHandler) actionChangePassword(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)

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
		return http.StatusInternalServerError, "", conf.ErrInternalServerError
	} else {
		user.Password = hash
	}

	if _, err := h.userSrvc.Update(user); err != nil {
		return http.StatusInternalServerError, "", conf.ErrInternalServerError
	}

	login := &models.Login{
		Username: user.ID,
		Password: user.Password,
	}
	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, login.Username)
	if err != nil {
		return http.StatusInternalServerError, "", conf.ErrInternalServerError
	}

	http.SetCookie(w, h.config.CreateCookie(models.AuthCookieKey, encoded))
	return http.StatusOK, "password was updated successfully", ""
}

func (h *SettingsHandler) actionResetApiKey(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	if _, err := h.userSrvc.ResetApiKey(user); err != nil {
		return http.StatusInternalServerError, "", conf.ErrInternalServerError
	}

	msg := fmt.Sprintf("your new api key is: %s", user.ApiKey)
	return http.StatusOK, msg, ""
}

func (h *SettingsHandler) actionUpdateLeaderboard(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	var err error
	user := middlewares.GetPrincipal(r)
	defer h.userSrvc.FlushCache()

	user.PublicLeaderboard, err = strconv.ParseBool(r.PostFormValue("enable_leaderboard"))

	if err != nil {
		return http.StatusBadRequest, "", "invalid input"
	}
	if _, err := h.userSrvc.Update(user); err != nil {
		return http.StatusInternalServerError, "", "internal sever error"
	}
	return http.StatusOK, "settings updated", ""
}

func (h *SettingsHandler) actionUpdateSharing(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	var err error
	user := middlewares.GetPrincipal(r)

	defer h.userSrvc.FlushUserCache(user.ID)

	user.ShareProjects, err = strconv.ParseBool(r.PostFormValue("share_projects"))
	user.ShareLanguages, err = strconv.ParseBool(r.PostFormValue("share_languages"))
	user.ShareEditors, err = strconv.ParseBool(r.PostFormValue("share_editors"))
	user.ShareOSs, err = strconv.ParseBool(r.PostFormValue("share_oss"))
	user.ShareMachines, err = strconv.ParseBool(r.PostFormValue("share_machines"))
	user.ShareLabels, err = strconv.ParseBool(r.PostFormValue("share_labels"))
	user.ShareDataMaxDays, err = strconv.Atoi(r.PostFormValue("max_days"))

	if err != nil {
		return http.StatusBadRequest, "", "invalid input"
	}

	if _, err := h.userSrvc.Update(user); err != nil {
		return http.StatusInternalServerError, "", "internal sever error"
	}

	return http.StatusOK, "settings updated", ""
}

func (h *SettingsHandler) actionDeleteAlias(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
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
	user := middlewares.GetPrincipal(r)
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

func (h *SettingsHandler) actionAddLabel(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}
	user := middlewares.GetPrincipal(r)

	label := &models.ProjectLabel{
		UserID:     user.ID,
		ProjectKey: r.PostFormValue("key"),
		Label:      r.PostFormValue("value"),
	}

	if !label.IsValid() {
		return http.StatusBadRequest, "", "invalid input"
	}

	if _, err := h.projectLabelSrvc.Create(label); err != nil {
		// TODO: distinguish between bad request, conflict and server error
		return http.StatusBadRequest, "", "invalid input"
	}

	return http.StatusOK, "label added successfully", ""
}

func (h *SettingsHandler) addProjectToLabel(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}
	user := middlewares.GetPrincipal(r)
	label := &models.ProjectLabel{
		UserID:     user.ID,
		ProjectKey: r.PostFormValue("project"),
		Label:      r.PostFormValue("label"),
	}

	if !label.IsValid() {
		return http.StatusBadRequest, "", "invalid input"
	}

	if _, err := h.projectLabelSrvc.Create(label); err != nil {
		// TODO: distinguish between bad request, conflict and server error
		return http.StatusBadRequest, "", "invalid input"
	}

	return http.StatusOK, "added project to label successfully", ""
}

func (h *SettingsHandler) actionDeleteLabel(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	labelKey := r.PostFormValue("key")     // label key
	labelValue := r.PostFormValue("value") // project key

	labels, err := h.projectLabelSrvc.GetByUser(user.ID)
	if err != nil {
		return http.StatusInternalServerError, "", "could not delete label"
	}

	for _, l := range labels {
		if l.Label == labelKey && l.ProjectKey == labelValue {
			if err := h.projectLabelSrvc.Delete(l); err != nil {
				return http.StatusInternalServerError, "", "could not delete label"
			}
			return http.StatusOK, "label deleted successfully", ""
		}
	}
	return http.StatusNotFound, "", "label not found"
}

func (h *SettingsHandler) actionDeleteLanguageMapping(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	id, err := strconv.Atoi(r.PostFormValue("mapping_id"))
	if err != nil {
		return http.StatusInternalServerError, "", "could not delete mapping"
	}

	mapping, err := h.languageMappingSrvc.GetById(uint(id))
	if err != nil || mapping == nil {
		return http.StatusNotFound, "", "mapping not found"
	} else if mapping.UserID != user.ID {
		return http.StatusForbidden, "", "not allowed to delete mapping"
	}

	if err := h.languageMappingSrvc.Delete(mapping); err != nil {
		return http.StatusInternalServerError, "", "could not delete mapping"
	}

	return http.StatusOK, "mapping deleted successfully", ""
}

func (h *SettingsHandler) actionAddLanguageMapping(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}
	user := middlewares.GetPrincipal(r)
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

func (h *SettingsHandler) actionSetWakatimeApiKey(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	apiKey := r.PostFormValue("api_key")
	apiUrl := r.PostFormValue("api_url")
	if apiUrl == conf.WakatimeApiUrl || apiKey == "" {
		apiUrl = ""
	}

	// Healthcheck, if a new API key is set, i.e. the feature is activated
	if (user.WakatimeApiKey == "" && apiKey != "") && !h.validateWakatimeKey(apiKey, apiUrl) {
		return http.StatusBadRequest, "", "failed to connect to WakaTime, API key or endpoint URL invalid?"
	}

	if _, err := h.userSrvc.SetWakatimeApiCredentials(user, apiKey, apiUrl); err != nil {
		return http.StatusInternalServerError, "", conf.ErrInternalServerError
	}

	return http.StatusOK, "Wakatime API Key updated successfully", ""
}

func (h *SettingsHandler) actionImportWakatime(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	if user.WakatimeApiKey == "" {
		return http.StatusForbidden, "", "not connected to wakatime"
	}

	kvKey := fmt.Sprintf("%s_%s", conf.KeyLastImportImport, user.ID)

	if !h.config.IsDev() {
		lastImportKv := h.keyValueSrvc.MustGetString(kvKey)
		lastImport, _ := time.Parse(time.RFC822, lastImportKv.Value)
		if time.Now().Sub(lastImport) < time.Duration(h.config.App.ImportBackoffMin)*time.Minute {
			return http.StatusTooManyRequests,
				"",
				fmt.Sprintf("Too many data imports. You are only allowed to request an import every %d minutes.", h.config.App.ImportBackoffMin)
		}
	}

	go func(user *models.User) {
		start := time.Now()
		importer := imports.NewWakatimeHeartbeatImporter(user.WakatimeApiKey)

		countBefore, err := h.heartbeatSrvc.CountByUser(user)
		if err != nil {
			println(err)
		}

		var stream <-chan *models.Heartbeat
		if latest, err := h.heartbeatSrvc.GetLatestByOriginAndUser(imports.OriginWakatime, user); latest == nil || err != nil {
			stream = importer.ImportAll(user)
		} else {
			// if an import has happened before, only import heartbeats newer than the latest of the last import
			stream = importer.Import(user, latest.Time.T(), time.Now())
		}

		count := 0
		batch := make([]*models.Heartbeat, 0)

		insert := func(batch []*models.Heartbeat) {
			if err := h.heartbeatSrvc.InsertBatch(batch); err != nil {
				logbuch.Warn("failed to insert imported heartbeat, already existing? - %v", err)
			}
		}

		for hb := range stream {
			count++
			batch = append(batch, hb)

			if len(batch) == h.config.App.ImportBatchSize {
				insert(batch)
				batch = make([]*models.Heartbeat, 0)
			}
		}

		if len(batch) > 0 {
			insert(batch)
		}

		countAfter, _ := h.heartbeatSrvc.CountByUser(user)
		logbuch.Info("downloaded %d heartbeats for user '%s' (%d actually imported)", count, user.ID, countAfter-countBefore)

		h.regenerateSummaries(user)

		if !user.HasData {
			user.HasData = true
			if _, err := h.userSrvc.Update(user); err != nil {
				conf.Log().Request(r).Error("failed to set 'has_data' flag for user %s - %v", user.ID, err)
			}
		}

		if user.Email != "" {
			if err := h.mailSrvc.SendImportNotification(user, time.Now().Sub(start), int(countAfter-countBefore)); err != nil {
				conf.Log().Request(r).Error("failed to send import notification mail to %s - %v", user.ID, err)
			} else {
				logbuch.Info("sent import notification mail to %s", user.ID)
			}
		}
	}(user)

	h.keyValueSrvc.PutString(&models.KeyStringValue{
		Key:   kvKey,
		Value: time.Now().Format(time.RFC822),
	})

	return http.StatusAccepted, "Import started. This will take several minutes. Please check back later.", ""
}

func (h *SettingsHandler) actionRegenerateSummaries(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	go func(user *models.User) {
		if err := h.regenerateSummaries(user); err != nil {
			conf.Log().Request(r).Error("failed to regenerate summaries for user '%s' - %v", user.ID, err)
		}
	}(middlewares.GetPrincipal(r))

	return http.StatusAccepted, "summaries are being regenerated - this may take a up to a couple of minutes, please come back later", ""
}

func (h *SettingsHandler) actionClearData(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	logbuch.Info("user '%s' requested to delete all data", user.ID)

	go func(user *models.User) {
		logbuch.Info("deleting summaries for user '%s'", user.ID)
		if err := h.summarySrvc.DeleteByUser(user.ID); err != nil {
			logbuch.Error("failed to clear summaries: %v", err)
		}

		logbuch.Info("deleting heartbeats for user '%s'", user.ID)
		if err := h.heartbeatSrvc.DeleteByUser(user); err != nil {
			logbuch.Error("failed to clear heartbeats: %v", err)
		}
	}(user)

	return http.StatusAccepted, "deletion in progress, this may take a couple of seconds", ""
}

func (h *SettingsHandler) actionDeleteUser(w http.ResponseWriter, r *http.Request) (int, string, string) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	go func(user *models.User) {
		logbuch.Info("deleting user '%s' shortly", user.ID)
		time.Sleep(5 * time.Minute)
		if err := h.userSrvc.Delete(user); err != nil {
			conf.Log().Request(r).Error("failed to delete user '%s' - %v", user.ID, err)
		} else {
			logbuch.Info("successfully deleted user '%s'", user.ID)
		}
	}(user)

	routeutils.SetSuccess(r, w, "Your account will be deleted in a few minutes. Sorry to you go.")
	http.SetCookie(w, h.config.GetClearCookie(models.AuthCookieKey))
	http.Redirect(w, r, h.config.Server.BasePath, http.StatusFound)
	return -1, "", ""
}

func (h *SettingsHandler) validateWakatimeKey(apiKey string, baseUrl string) bool {
	if baseUrl == "" {
		baseUrl = conf.WakatimeApiUrl
	}

	headers := http.Header{
		"Accept": []string{"application/json"},
		"Authorization": []string{
			fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(apiKey))),
		},
	}

	request, err := http.NewRequest(
		http.MethodGet,
		baseUrl+conf.WakatimeApiUserUrl,
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

func (h *SettingsHandler) regenerateSummaries(user *models.User) error {
	logbuch.Info("clearing summaries for user '%s'", user.ID)
	if err := h.summarySrvc.DeleteByUser(user.ID); err != nil {
		conf.Log().Error("failed to clear summaries: %v", err)
		return err
	}

	if err := h.aggregationSrvc.AggregateSummaries(datastructure.NewSet(user.ID)); err != nil {
		conf.Log().Error("failed to regenerate summaries: %v", err)
		return err
	}

	return nil
}

func (h *SettingsHandler) buildViewModel(r *http.Request, w http.ResponseWriter) *view.SettingsViewModel {
	user := middlewares.GetPrincipal(r)

	// mappings
	mappings, _ := h.languageMappingSrvc.GetByUser(user.ID)

	// aliases
	aliases, err := h.aliasSrvc.GetByUser(user.ID)
	if err != nil {
		conf.Log().Request(r).Error("error while building alias map - %v", err)
		return &view.SettingsViewModel{Messages: view.Messages{Error: criticalError}}
	}
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

	// labels
	labelMap, err := h.projectLabelSrvc.GetByUserGroupedInverted(user.ID)
	if err != nil {
		conf.Log().Request(r).Error("error while building settings project label map - %v", err)
		return &view.SettingsViewModel{Messages: view.Messages{Error: criticalError}}
	}

	combinedLabels := make([]*view.SettingsVMCombinedLabel, 0)
	for _, l := range labelMap {
		cl := &view.SettingsVMCombinedLabel{
			Key:    l[0].Label,
			Values: make([]string, len(l)),
		}
		for i, l1 := range l {
			cl.Values[i] = l1.ProjectKey
		}
		combinedLabels = append(combinedLabels, cl)
	}
	sort.Slice(combinedLabels, func(i, j int) bool {
		return strings.Compare(combinedLabels[i].Key, combinedLabels[j].Key) < 0
	})

	// projects
	projects, err := routeutils.GetEffectiveProjectsList(user, h.heartbeatSrvc, h.aliasSrvc)
	if err != nil {
		conf.Log().Request(r).Error("error while fetching projects - %v", err)
		return &view.SettingsViewModel{Messages: view.Messages{Error: criticalError}}
	}

	// subscriptions
	var subscriptionPrice string
	if h.config.Subscriptions.Enabled {
		subscriptionPrice = h.config.Subscriptions.StandardPrice
	}

	// user first data
	var firstData time.Time
	firstDataKv := h.keyValueSrvc.MustGetString(fmt.Sprintf("%s_%s", conf.KeyFirstHeartbeat, user.ID))
	if firstDataKv.Value != "" {
		firstData, _ = time.Parse(time.RFC822Z, firstDataKv.Value)
	}

	vm := &view.SettingsViewModel{
		User:                user,
		LanguageMappings:    mappings,
		Aliases:             combinedAliases,
		Labels:              combinedLabels,
		Projects:            projects,
		ApiKey:              user.ApiKey,
		UserFirstData:       firstData,
		SubscriptionPrice:   subscriptionPrice,
		SupportContact:      h.config.App.SupportContact,
		DataRetentionMonths: h.config.App.DataRetentionMonths,
	}
	return routeutils.WithSessionMessages(vm, r, w)
}
