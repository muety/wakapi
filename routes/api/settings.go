package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

type SettingsHandler struct {
	config     *config.Config
	userSrvc   services.IUserService
	httpClient *http.Client
}

type actionResult struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Values  map[string]interface{} `json:"values,omitempty"`
}

func NewSettingsHandler(userService services.IUserService) *SettingsHandler {
	return &SettingsHandler{
		config:     config.Get(),
		userSrvc:   userService,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *SettingsHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(
			middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
		)
		r.Post("/settings", h.PostIndex)
	})
}

type WakatimeSettingsPayload struct {
	ApiKey string `json:"api_key"`
	ApiUrl string `json:"api_url"`
}

type SettingsPayload struct {
	Action string `json:"action"`
	WakatimeSettingsPayload
}

func (h *SettingsHandler) PostIndex(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)

	var reqBody = &SettingsPayload{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if reqBody.Action != "toggle_wakatime" {
		h.respondWithError(w, http.StatusBadRequest, "Unknown action")
		return
	}

	result := h.actionSetWakatimeApiKey(reqBody, user)
	if result.Code != -1 {
		h.respondWithJSON(w, result.Code, result)
	}
}

func (h *SettingsHandler) actionSetWakatimeApiKey(wakatimeSettings *SettingsPayload, user *models.User) actionResult {
	if wakatimeSettings.ApiKey == "" {
		return actionResult{http.StatusBadRequest, "", "invalid input: no or invalid api key", nil}
	}

	if !h.validateWakatimeKey(wakatimeSettings.ApiKey, wakatimeSettings.ApiUrl) {
		return actionResult{http.StatusBadRequest, "", "invalid input: failed to validate api key against wakatime server", nil}
	}

	user.WakatimeApiKey = wakatimeSettings.ApiKey

	if _, err := h.userSrvc.Update(user); err != nil {
		return actionResult{http.StatusInternalServerError, "", "Internal server error", nil}
	}

	return actionResult{http.StatusOK, "Wakatime API key set", "", nil}
}

func (h *SettingsHandler) respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func (h *SettingsHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *SettingsHandler) validateWakatimeKey(apiKey string, baseUrl string) bool {
	if baseUrl == "" {
		baseUrl = config.WakatimeApiUrl
	}

	headers := http.Header{
		"Accept": []string{"application/json"},
		"Authorization": []string{
			fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(apiKey))),
		},
	}

	request, err := http.NewRequest(
		http.MethodGet,
		baseUrl+config.WakatimeApiUserUrl,
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
