package routes

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/condition"
	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/schema"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
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
	durationSrvc        services.IDurationService
	aliasSrvc           services.IAliasService
	aggregationSrvc     services.IAggregationService
	languageMappingSrvc services.ILanguageMappingService
	projectLabelSrvc    services.IProjectLabelService
	keyValueSrvc        services.IKeyValueService
	mailSrvc            services.IMailService
	apiKeySrvc          services.IApiKeyService
	httpClient          *http.Client
	aggregationLocks    map[string]bool
}

type action func(w http.ResponseWriter, r *http.Request) actionResult

type actionResult struct {
	code    int
	success string
	error   string
	values  *map[string]interface{}
}

const valueInviteCode = "invite_code"

var credentialsDecoder = schema.NewDecoder()

func NewSettingsHandler(
	userService services.IUserService,
	heartbeatService services.IHeartbeatService,
	durationService services.IDurationService,
	summaryService services.ISummaryService,
	aliasService services.IAliasService,
	aggregationService services.IAggregationService,
	languageMappingService services.ILanguageMappingService,
	projectLabelService services.IProjectLabelService,
	keyValueService services.IKeyValueService,
	mailService services.IMailService,
	apiKeyService services.IApiKeyService,
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
		durationSrvc:        durationService,
		keyValueSrvc:        keyValueService,
		mailSrvc:            mailService,
		apiKeySrvc:          apiKeyService,
		httpClient:          &http.Client{Timeout: 10 * time.Second},
		aggregationLocks:    make(map[string]bool),
	}
}

func (h *SettingsHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).
			WithRedirectTarget(defaultErrorRedirectTarget()).
			WithRedirectErrorMessage("unauthorized").Handler,
	)
	r.Get("/", h.GetIndex)
	r.Post("/", h.PostIndex)

	router.Mount("/settings", r)
}

func (h *SettingsHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
	err := templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w, nil))
	if err != nil {
		panic(err)
	}
}

func (h *SettingsHandler) PostIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err = templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w, nil).WithError("missing form values"))
		if err != nil {
			panic(err)
		}
		return
	}

	action := r.PostForm.Get("action")
	r.PostForm.Del("action")

	actionFunc := h.dispatchAction(action)
	if actionFunc == nil {
		slog.Warn("failed to dispatch action", "action", action)
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w, nil).WithError("unknown action requests"))
		return
	}

	result := actionFunc(w, r)

	// action responded itself
	if result.code == -1 {
		return
	}

	if result.error != "" {
		w.WriteHeader(result.code)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w, result.values).WithError(result.error))
		return
	}
	if result.success != "" {
		w.WriteHeader(result.code)
		templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w, result.values).WithSuccess(result.success))
		return
	}
	templates[conf.SettingsTemplate].Execute(w, h.buildViewModel(r, w, result.values))
}

func (h *SettingsHandler) dispatchAction(action string) action {
	switch action {
	case "change_password":
		return h.actionChangePassword
	case "change_userid":
		return h.actionChangeUserId
	case "update_user":
		return h.actionUpdateUser
	case "reset_apikey":
		return h.actionResetApiKey
	case "delete_alias":
		return h.actionDeleteAlias
	case "add_alias":
		return h.actionAddAlias
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
	case "generate_invite":
		return h.actionGenerateInvite
	case "update_unknown_projects":
		return h.actionUpdateExcludeUnknownProjects
	case "update_heartbeats_timeout":
		return h.actionUpdateHeartbeatsTimeout
	case "update_readme_stats_base_url":
		return h.actionUpdateReadmeStatsBaseUrl
	case "add_api_key":
		return h.actionAddApiKey
	case "delete_api_key":
		return h.actionDeleteApiKey
	}
	return nil
}

func (h *SettingsHandler) actionUpdateUser(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)

	var payload models.UserDataUpdate
	if err := r.ParseForm(); err != nil {
		return actionResult{http.StatusBadRequest, "", "missing parameters", nil}
	}
	if err := credentialsDecoder.Decode(&payload, r.PostForm); err != nil {
		return actionResult{http.StatusBadRequest, "", "missing parameters", nil}
	}

	if !payload.IsValid() {
		return actionResult{http.StatusBadRequest, "", "invalid parameters - perhaps invalid e-mail address?", nil}
	}

	if payload.Email == "" && user.HasActiveSubscription() {
		return actionResult{http.StatusBadRequest, "", "cannot unset email while subscription is active", nil}
	}

	user.Email = payload.Email
	user.Location = payload.Location
	user.StartOfWeek = payload.StartOfWeek
	user.ReportsWeekly = payload.ReportsWeekly
	user.PublicLeaderboard = payload.PublicLeaderboard

	if _, err := h.userSrvc.Update(user); err != nil {
		if strings.Contains(err.Error(), "email address already in use") {
			return actionResult{http.StatusBadRequest, "", "got invalid user data (email already taken?)", nil}
		}
		return actionResult{http.StatusInternalServerError, "", conf.ErrInternalServerError, nil}
	}

	return actionResult{http.StatusOK, "user updated successfully", "", nil}
}

func (h *SettingsHandler) actionChangePassword(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)

	if user.AuthType != "local" {
		return actionResult{http.StatusBadRequest, "", "cannot reset password for non-local user", nil}
	}

	var credentials models.CredentialsReset
	if err := r.ParseForm(); err != nil {
		return actionResult{http.StatusBadRequest, "", "missing parameters", nil}
	}
	if err := credentialsDecoder.Decode(&credentials, r.PostForm); err != nil {
		return actionResult{http.StatusBadRequest, "", "missing parameters", nil}
	}

	if !utils.ComparePassword(user.Password, credentials.PasswordOld, h.config.Security.PasswordSalt) {
		return actionResult{http.StatusUnauthorized, "", "invalid credentials", nil}
	}

	if !credentials.IsValid() {
		return actionResult{http.StatusBadRequest, "", "invalid parameters", nil}
	}

	user.Password = credentials.PasswordNew
	if hash, err := utils.HashPassword(user.Password, h.config.Security.PasswordSalt); err != nil {
		return actionResult{http.StatusInternalServerError, "", conf.ErrInternalServerError, nil}
	} else {
		user.Password = hash
	}

	if _, err := h.userSrvc.Update(user); err != nil {
		return actionResult{http.StatusInternalServerError, "", conf.ErrInternalServerError, nil}
	}

	login := &models.Login{
		Username: user.ID,
		Password: user.Password,
	}
	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, login.Username)
	if err != nil {
		return actionResult{http.StatusInternalServerError, "", conf.ErrInternalServerError, nil}
	}

	http.SetCookie(w, h.config.CreateCookie(models.AuthCookieKey, encoded))
	return actionResult{http.StatusOK, "password was updated successfully", "", nil}
}

func (h *SettingsHandler) actionChangeUserId(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)

	newUserId := strings.TrimSpace(r.PostFormValue("new_userid"))
	if !models.ValidateUsername(newUserId) || newUserId == user.ID {
		return actionResult{http.StatusBadRequest, "", "invalid username", nil}
	}
	if existing, _ := h.userSrvc.GetUserById(newUserId); existing != nil {
		return actionResult{http.StatusConflict, "", "already taken", nil}
	}

	if _, err := h.userSrvc.ChangeUserId(user, newUserId); err != nil {
		return actionResult{http.StatusInternalServerError, "", conf.ErrInternalServerError, nil}
	}

	routeutils.SetSuccess(r, w, fmt.Sprintf("Successfully changed your username to %s, please log back in.", newUserId))
	http.SetCookie(w, h.config.GetClearCookie(models.AuthCookieKey))
	http.Redirect(w, r, h.config.Server.BasePath, http.StatusFound)
	return actionResult{-1, "", "", nil}
}

func (h *SettingsHandler) actionResetApiKey(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	if _, err := h.userSrvc.ResetApiKey(user); err != nil {
		return actionResult{http.StatusInternalServerError, "", conf.ErrInternalServerError, nil}
	}

	msg := fmt.Sprintf("your new api key is: %s", user.ApiKey)
	return actionResult{http.StatusOK, msg, "", nil}
}

func (h *SettingsHandler) actionUpdateLeaderboard(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	var err error
	user := middlewares.GetPrincipal(r)
	defer h.userSrvc.FlushCache()

	user.PublicLeaderboard, err = strconv.ParseBool(r.PostFormValue("enable_leaderboard"))

	if err != nil {
		return actionResult{http.StatusBadRequest, "", "invalid input", nil}
	}
	if _, err := h.userSrvc.Update(user); err != nil {
		return actionResult{http.StatusInternalServerError, "", "internal sever error", nil}
	}
	return actionResult{http.StatusOK, "settings updated", "", nil}
}

func (h *SettingsHandler) actionUpdateExcludeUnknownProjects(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	var err error
	user := middlewares.GetPrincipal(r)
	defer h.userSrvc.FlushCache()

	if h.isAggregationLocked(user.ID) {
		return actionResult{http.StatusConflict, "", "summary regeneration already in progress, please wait", nil}
	}

	user.ExcludeUnknownProjects, err = strconv.ParseBool(r.PostFormValue("exclude_unknown_projects"))

	if err != nil {
		return actionResult{http.StatusBadRequest, "", "invalid input", nil}
	}
	if _, err := h.userSrvc.Update(user); err != nil {
		return actionResult{http.StatusInternalServerError, "", "internal sever error", nil}
	}

	go func(user *models.User, r *http.Request) {
		h.toggleAggregationLock(user.ID, true)
		defer h.toggleAggregationLock(user.ID, false)
		if err := h.regenerateSummaries(user); err != nil {
			conf.Log().Request(r).Error("failed to regenerate summaries for user", "userID", user.ID, "error", err)
		}
	}(user, r)

	return actionResult{http.StatusOK, "regenerating summaries, this might take a while", "", nil}
}

func (h *SettingsHandler) actionUpdateHeartbeatsTimeout(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	var err error
	user := middlewares.GetPrincipal(r)
	defer h.userSrvc.FlushCache()

	val, err := strconv.ParseInt(r.PostFormValue("heartbeats_timeout"), 0, 0)
	dur := time.Duration(val) * time.Minute
	if err != nil || dur < models.MinHeartbeatsTimeout || dur > models.MaxHeartbeatsTimeout {
		return actionResult{http.StatusBadRequest, "", "invalid input", nil}
	}
	user.HeartbeatsTimeoutSec = int(dur.Seconds())

	if _, err := h.userSrvc.Update(user); err != nil {
		return actionResult{http.StatusInternalServerError, "", "internal sever error", nil}
	}

	return actionResult{http.StatusOK, "Done. To apply this change to already existing data, please regenerate your summaries.", "", nil}
}

func (h *SettingsHandler) actionUpdateReadmeStatsBaseUrl(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	defer h.userSrvc.FlushUserCache(user.ID)

	user.ReadmeStatsBaseUrl = r.PostFormValue("readme_stats_base_url")

	if _, err := h.userSrvc.Update(user); err != nil {
		return actionResult{http.StatusInternalServerError, "", "internal sever error", nil}
	}

	return actionResult{http.StatusOK, "settings updated", "", nil}
}

func (h *SettingsHandler) actionUpdateSharing(w http.ResponseWriter, r *http.Request) actionResult {
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
	user.ShareActivityChart, err = strconv.ParseBool(r.PostFormValue("share_activity_chart"))
	user.ShareDataMaxDays, err = strconv.Atoi(r.PostFormValue("max_days"))

	if err != nil {
		return actionResult{http.StatusBadRequest, "", "invalid input", nil}
	}

	if _, err := h.userSrvc.Update(user); err != nil {
		return actionResult{http.StatusInternalServerError, "", "internal sever error", nil}
	}

	return actionResult{http.StatusOK, "settings updated", "", nil}
}

func (h *SettingsHandler) actionDeleteAlias(w http.ResponseWriter, r *http.Request) actionResult {
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
		return actionResult{http.StatusNotFound, "", "aliases not found", nil}
	} else if err := h.aliasSrvc.DeleteMulti(aliases); err != nil {
		return actionResult{http.StatusInternalServerError, "", "could not delete aliases", nil}
	}

	return actionResult{http.StatusOK, "aliases deleted successfully", "", nil}
}

func (h *SettingsHandler) actionAddAlias(w http.ResponseWriter, r *http.Request) actionResult {
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
		return actionResult{http.StatusBadRequest, "", "invalid input", nil}
	}

	return actionResult{http.StatusOK, "alias added successfully", "", nil}
}

func (h *SettingsHandler) actionAddLabel(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}
	user := middlewares.GetPrincipal(r)

	var labels []*models.ProjectLabel

	for _, key := range r.Form["key"] {
		label := &models.ProjectLabel{
			UserID:     user.ID,
			ProjectKey: key,
			Label:      r.PostFormValue("value"),
		}
		labels = append(labels, label)
	}

	for _, label := range labels {
		msg := "invalid input for project: " + label.ProjectKey
		if !label.IsValid() {
			return actionResult{http.StatusBadRequest, "", msg, nil}
		}
		if _, err := h.projectLabelSrvc.Create(label); err != nil {
			// TODO: distinguish between bad request, conflict and server error
			return actionResult{http.StatusBadRequest, "", msg, nil}
		}
	}
	return actionResult{http.StatusOK, "label added to project successfully", "", nil}
}

func (h *SettingsHandler) actionDeleteLabel(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	labelKey := r.PostFormValue("key")     // label key
	labelValue := r.PostFormValue("value") // project key

	labels, err := h.projectLabelSrvc.GetByUser(user.ID)
	if err != nil {
		return actionResult{http.StatusInternalServerError, "", "could not delete label", nil}
	}

	for _, l := range labels {
		if l.Label == labelKey && l.ProjectKey == labelValue {
			if err := h.projectLabelSrvc.Delete(l); err != nil {
				return actionResult{http.StatusInternalServerError, "", "could not delete label", nil}
			}
			return actionResult{http.StatusOK, "label deleted successfully", "", nil}
		}
	}
	return actionResult{http.StatusNotFound, "", "label not found", nil}
}

func (h *SettingsHandler) actionDeleteLanguageMapping(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	id, err := strconv.Atoi(r.PostFormValue("mapping_id"))
	if err != nil {
		return actionResult{http.StatusInternalServerError, "", "could not delete mapping", nil}
	}

	mapping, err := h.languageMappingSrvc.GetById(uint(id))
	if err != nil || mapping == nil {
		return actionResult{http.StatusNotFound, "", "mapping not found", nil}
	} else if mapping.UserID != user.ID {
		return actionResult{http.StatusForbidden, "", "not allowed to delete mapping", nil}
	}

	if err := h.languageMappingSrvc.Delete(mapping); err != nil {
		return actionResult{http.StatusInternalServerError, "", "could not delete mapping", nil}
	}

	return actionResult{http.StatusOK, "mapping deleted successfully", "", nil}
}

func (h *SettingsHandler) actionAddLanguageMapping(w http.ResponseWriter, r *http.Request) actionResult {
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
		return actionResult{http.StatusConflict, "", "mapping already exists", nil}
	}

	return actionResult{http.StatusOK, "mapping added successfully", "", nil}
}

func (h *SettingsHandler) actionSetWakatimeApiKey(w http.ResponseWriter, r *http.Request) actionResult {
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
		return actionResult{http.StatusBadRequest, "", "failed to connect to WakaTime, API key or endpoint URL invalid?", nil}
	}

	if _, err := h.userSrvc.SetWakatimeApiCredentials(user, apiKey, apiUrl); err != nil {
		return actionResult{http.StatusInternalServerError, "", conf.ErrInternalServerError, nil}
	}

	return actionResult{http.StatusOK, "Wakatime API Key updated successfully", "", nil}
}

func (h *SettingsHandler) actionImportWakatime(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	if !h.config.App.ImportEnabled {
		return actionResult{http.StatusForbidden, "", "imports are disabled on this server", nil}
	}

	user := middlewares.GetPrincipal(r)
	if user.WakatimeApiKey == "" {
		return actionResult{http.StatusForbidden, "", "not connected to wakatime", nil}
	}

	useLegacyImporter, _ := strconv.ParseBool(r.PostFormValue("use_legacy_importer"))
	kvKeyLastImport := fmt.Sprintf("%s_%s", conf.KeyLastImport, user.ID)
	kvKeyLastImportSuccess := fmt.Sprintf("%s_%s", conf.KeyLastImportSuccess, user.ID)

	if !h.config.IsDev() {
		lastImport, _ := time.Parse(time.RFC822, h.keyValueSrvc.MustGetString(kvKeyLastImport).Value)
		if time.Now().Sub(lastImport) < time.Duration(h.config.App.ImportBackoffMin)*time.Minute {
			return actionResult{
				http.StatusTooManyRequests,
				"",
				fmt.Sprintf("Too many data imports - you are only allowed to request an import every %d minutes.", h.config.App.ImportBackoffMin),
				nil,
			}
		}

		lastImportSuccess, _ := time.Parse(time.RFC822, h.keyValueSrvc.MustGetString(kvKeyLastImportSuccess).Value)
		if time.Now().Sub(lastImportSuccess) < time.Duration(h.config.App.ImportMaxRate)*time.Hour {
			return actionResult{
				http.StatusTooManyRequests,
				"",
				fmt.Sprintf("Too many data imports - last import ran less than %d hours ago, please wait.", h.config.App.ImportMaxRate),
				nil,
			}
		}
	}

	go func(user *models.User, r *http.Request) {
		start := time.Now()
		importer := imports.NewWakatimeImporter(user.WakatimeApiKey, useLegacyImporter)

		countBefore, _ := h.heartbeatSrvc.CountByUser(user)

		var (
			stream      <-chan *models.Heartbeat
			importError error
		)
		if latest, err := h.heartbeatSrvc.GetLatestByOriginAndUser(imports.OriginWakatime, user); latest == nil || err != nil {
			stream, importError = importer.ImportAll(user)
		} else {
			// if an import has happened before, only import heartbeats newer than the latest of the last import
			stream, importError = importer.Import(user, latest.Time.T(), time.Now())
		}
		if importError != nil {
			conf.Log().Error("wakatime import for user failed", "userID", user.ID, "error", importError)
			return
		}

		// import successful
		h.keyValueSrvc.PutString(&models.KeyStringValue{
			Key:   kvKeyLastImportSuccess,
			Value: time.Now().Format(time.RFC822),
		})

		count := 0
		batch := make([]*models.Heartbeat, 0, h.config.App.ImportBatchSize)

		insert := func(batch []*models.Heartbeat) {
			if err := h.heartbeatSrvc.InsertBatch(batch); err != nil {
				slog.Warn("failed to insert imported heartbeat, already existing?", "error", err)
			}
		}

		for hb := range stream {
			count++
			batch = append(batch, hb)

			if len(batch) == h.config.App.ImportBatchSize {
				insert(batch)
				batch = make([]*models.Heartbeat, 0, h.config.App.ImportBatchSize)
			}
		}
		if len(batch) > 0 {
			insert(batch)
		}

		countAfter, _ := h.heartbeatSrvc.CountByUser(user)
		slog.Info("downloaded heartbeats for user", "count", count, "userID", user.ID, "importedCount", countAfter-countBefore)

		h.regenerateSummaries(user)

		if !user.HasData {
			user.HasData = true
			if _, err := h.userSrvc.Update(user); err != nil {
				conf.Log().Request(r).Error("failed to set 'has_data' flag for user", "userID", user.ID, "error", err)
			}
		}

		if user.Email != "" {
			if err := h.mailSrvc.SendImportNotification(user, time.Now().Sub(start), int(countAfter-countBefore)); err != nil {
				conf.Log().Request(r).Error("failed to send import notification mail", "userID", user.ID, "error", err)
			} else {
				slog.Info("sent import notification mail", "userID", user.ID)
			}
		}
	}(user, r)

	h.keyValueSrvc.PutString(&models.KeyStringValue{
		Key:   kvKeyLastImport,
		Value: time.Now().Format(time.RFC822),
	})

	return actionResult{http.StatusAccepted, "Import started. This will take several minutes. Please check back later.", "", nil}
}

func (h *SettingsHandler) actionRegenerateSummaries(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)

	if h.isAggregationLocked(user.ID) {
		return actionResult{http.StatusConflict, "", "summary regeneration already in progress, please wait", nil}
	}

	go func(user *models.User, r *http.Request) {
		h.toggleAggregationLock(user.ID, true)
		defer h.toggleAggregationLock(user.ID, false)
		if err := h.regenerateSummaries(user); err != nil {
			conf.Log().Request(r).Error("failed to regenerate summaries for user", "userID", user.ID, "error", err)
		}
	}(user, r)

	return actionResult{http.StatusAccepted, "summaries are being regenerated - this may take a up to a couple of minutes, please come back later", "", nil}
}

func (h *SettingsHandler) actionClearData(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	slog.Info("user requested to delete all data", "userID", user.ID)

	go func(user *models.User, r *http.Request) {
		slog.Info("deleting summaries for user", "userID", user.ID)
		if err := h.summarySrvc.DeleteByUser(user.ID); err != nil {
			conf.Log().Request(r).Error("failed to clear summaries", "error", err)
		}

		slog.Info("deleting durations for user", "userID", user.ID)
		if err := h.durationSrvc.DeleteByUser(user); err != nil {
			conf.Log().Request(r).Error("failed to clear durations", "error", err)
		}

		slog.Info("deleting heartbeats for user", "userID", user.ID)
		if err := h.heartbeatSrvc.DeleteByUser(user); err != nil {
			conf.Log().Request(r).Error("failed to clear heartbeats", "error", err)
		}
	}(user, r)

	return actionResult{http.StatusAccepted, "deletion in progress, this may take a couple of seconds", "", nil}
}

func (h *SettingsHandler) actionDeleteUser(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	go func(user *models.User, r *http.Request) {
		slog.Info("deleting user shortly", "userID", user.ID)
		//time.Sleep(5 * time.Minute)
		if err := h.userSrvc.Delete(user); err != nil {
			conf.Log().Request(r).Error("failed to delete user", "userID", user.ID, "error", err)
		} else {
			slog.Info("successfully deleted user", "userID", user.ID)
		}
	}(user, r)

	routeutils.SetSuccess(r, w, "Your account will be deleted in a few minutes. Sorry to see you go.")
	http.SetCookie(w, h.config.GetClearCookie(models.AuthCookieKey))
	http.Redirect(w, r, h.config.Server.BasePath, http.StatusFound)
	return actionResult{-1, "", "", nil}
}

func (h *SettingsHandler) actionGenerateInvite(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	inviteCode := uuid.Must(uuid.NewV4()).String()[0:8]

	if err := h.keyValueSrvc.PutString(&models.KeyStringValue{
		Key:   fmt.Sprintf("%s_%s", conf.KeyInviteCode, inviteCode),
		Value: fmt.Sprintf("%s,%s", user.ID, time.Now().Format(time.RFC3339)),
	}); err != nil {
		return actionResult{http.StatusInternalServerError, "", "failed to generate invite code", nil}
	}

	return actionResult{
		http.StatusOK,
		"Successfully generated new invite code (see below)",
		"",
		&map[string]interface{}{
			valueInviteCode: inviteCode,
		},
	}
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

	if _, err = utils.RaiseForStatus(h.httpClient.Do(request)); err != nil {
		return false
	}

	return true
}

func (h *SettingsHandler) regenerateSummaries(user *models.User) error {
	slog.Info("clearing summaries and durations for user", "userID", user.ID)

	if err := h.summarySrvc.DeleteByUser(user.ID); err != nil {
		conf.Log().Error("failed to clear summaries", "error", err)
		return err
	}

	if err := h.aggregationSrvc.AggregateSummaries(datastructure.New(user.ID)); err != nil { // involves regenerating durations as well
		conf.Log().Error("failed to regenerate summaries", "error", err)
		return err
	}

	return nil
}

func (h *SettingsHandler) actionAddApiKey(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	apiKey := uuid.Must(uuid.NewV4()).String()

	if _, err := h.apiKeySrvc.Create(&models.ApiKey{
		User:     middlewares.GetPrincipal(r),
		Label:    r.PostFormValue("api_name"),
		ApiKey:   apiKey,
		ReadOnly: r.PostFormValue("api_readonly") == "true",
	}); err != nil {
		return actionResult{http.StatusInternalServerError, "", conf.ErrInternalServerError, nil}
	}

	msg := fmt.Sprintf("you added new api key: %s", apiKey)
	return actionResult{http.StatusOK, msg, "", nil}
}

func (h *SettingsHandler) actionDeleteApiKey(w http.ResponseWriter, r *http.Request) actionResult {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	apiKeyValue := r.PostFormValue("api_key_value")

	if apiKeyValue == user.ApiKey {
		return actionResult{http.StatusBadRequest, "", "Main api key can only be regenerated, not deleted", nil}
	}

	apiKeys, err := h.apiKeySrvc.GetByUser(user.ID)
	if err != nil {
		return actionResult{http.StatusInternalServerError, "", "could not delete API key", nil}
	}

	for _, k := range apiKeys {
		if k.ApiKey == apiKeyValue {
			if err := h.apiKeySrvc.Delete(k); err != nil {
				return actionResult{http.StatusInternalServerError, "", "could not delete API key", nil}
			}
			return actionResult{http.StatusOK, "API key deleted successfully", "", nil}
		}
	}
	return actionResult{http.StatusNotFound, "", "API key not found", nil}
}

func (h *SettingsHandler) buildViewModel(r *http.Request, w http.ResponseWriter, args *map[string]interface{}) *view.SettingsViewModel {
	user := middlewares.GetPrincipal(r)

	// mappings
	mappings, _ := h.languageMappingSrvc.GetByUser(user.ID)

	// aliases
	aliases, err := h.aliasSrvc.GetByUser(user.ID)
	if err != nil {
		conf.Log().Request(r).Error("error while building alias map", "error", err)
		return &view.SettingsViewModel{
			SharedLoggedInViewModel: view.SharedLoggedInViewModel{
				SharedViewModel: view.NewSharedViewModel(h.config, &view.Messages{Error: criticalError}),
				User:            user,
			},
		}
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
		conf.Log().Request(r).Error("error while building settings project label map", "error", err)
		return &view.SettingsViewModel{
			SharedLoggedInViewModel: view.SharedLoggedInViewModel{
				SharedViewModel: view.NewSharedViewModel(h.config, &view.Messages{Error: criticalError}),
				User:            user,
			},
		}
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
		conf.Log().Request(r).Error("error while fetching projects", "error", err)
		return &view.SettingsViewModel{
			SharedLoggedInViewModel: view.SharedLoggedInViewModel{
				SharedViewModel: view.NewSharedViewModel(h.config, &view.Messages{Error: criticalError}),
				User:            user,
			},
		}
	}

	// subscriptions
	var subscriptionPrice string
	if h.config.Subscriptions.Enabled {
		subscriptionPrice = h.config.Subscriptions.StandardPrice
	}

	// user first data
	firstData, err := h.heartbeatSrvc.GetFirstByUser(user)
	if err != nil {
		conf.Log().Request(r).Error("error while user's heartbeats range", "user", user.ID, "error", err)
		return &view.SettingsViewModel{
			SharedLoggedInViewModel: view.SharedLoggedInViewModel{
				SharedViewModel: view.NewSharedViewModel(h.config, &view.Messages{Error: criticalError}),
				User:            user,
			},
		}
	}

	// invite link
	inviteCode := getVal[string](args, valueInviteCode, "")
	inviteLink := condition.Ternary[bool, string](inviteCode == "", "", fmt.Sprintf("%s/signup?invite=%s", h.config.Server.GetPublicUrl(), inviteCode))

	// API keys
	combinedApiKeys := []*view.SettingsApiKeys{
		{
			Name:     "Main API Key",
			Value:    user.ApiKey,
			ReadOnly: false,
		},
	}

	apiKeys, err := h.apiKeySrvc.GetByUser(user.ID)
	if err != nil {
		conf.Log().Request(r).Error("error while fetching user's api keys", "user", user.ID, "error", err)
		return &view.SettingsViewModel{
			SharedLoggedInViewModel: view.SharedLoggedInViewModel{
				SharedViewModel: view.NewSharedViewModel(h.config, &view.Messages{Error: criticalError}),
				User:            user,
			},
		}
	}
	for _, apiKey := range apiKeys {
		combinedApiKeys = append(combinedApiKeys, &view.SettingsApiKeys{
			Name:     apiKey.Label,
			Value:    apiKey.ApiKey,
			ReadOnly: apiKey.ReadOnly,
		})
	}

	vm := &view.SettingsViewModel{
		SharedLoggedInViewModel: view.SharedLoggedInViewModel{
			SharedViewModel: view.NewSharedViewModel(h.config, nil),
			User:            user,
		},
		LanguageMappings:    mappings,
		Aliases:             combinedAliases,
		Labels:              combinedLabels,
		Projects:            projects,
		UserFirstData:       firstData,
		SubscriptionPrice:   subscriptionPrice,
		SupportContact:      h.config.App.SupportContact,
		DataRetentionMonths: h.config.App.DataRetentionMonths,
		InviteLink:          inviteLink,
		ApiKeys:             combinedApiKeys,
	}

	// readme card params
	readmeCardTitle := "Wakapi.dev Stats"
	if err, maxRange := helpers.ResolveMaximumRange(user.ShareDataMaxDays); err == nil {
		readmeCardTitle += fmt.Sprintf(" (%v)", maxRange.GetHumanReadable())
	}
	vm.ReadmeCardCustomTitle = readmeCardTitle

	return routeutils.WithSessionMessages(vm, r, w)
}

func (h *SettingsHandler) toggleAggregationLock(userId string, locked bool) {
	h.aggregationLocks[userId] = locked
}

func (h *SettingsHandler) isAggregationLocked(userId string) bool {
	locked, _ := h.aggregationLocks[userId]
	return locked
}

func getVal[T any](values *map[string]interface{}, key string, fallback T) T {
	if values == nil {
		return fallback
	}
	valuesMap := *values
	val, ok := valuesMap[key]
	if !ok {
		return fallback
	}
	return val.(T)
}
