package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
)

type actionResult struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Values  map[string]interface{} `json:"values,omitempty"`
}

type WakatimeSettingsPayload struct {
	ApiKey string `json:"api_key"`
	ApiUrl string `json:"api_url"`
}

type SettingsPayload struct {
	Action string `json:"action"`
	WakatimeSettingsPayload
}

func (a *APIv1) UpdateWakatimeSettings(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)

	var reqBody = &SettingsPayload{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		a.respondWithError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if reqBody.Action != "toggle_wakatime" {
		a.respondWithError(w, r, http.StatusBadRequest, "Unknown action")
		return
	}

	result := a.actionSetWakatimeApiKey(reqBody, user)
	if result.Code != -1 {
		helpers.RespondJSON(w, r, result.Code, result)
	}
}

func (a *APIv1) SaveProfile(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)

	var reqBody = &models.Profile{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		a.respondWithError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	fmt.Println("Saving the fucker")

	result := a.db.Model(user).Updates(reqBody)
	if err := result.Error; err != nil {
		helpers.RespondJSON(w, r, 400, map[string]interface{}{
			"code":  400,
			"error": err.Error(),
		})
	}
	helpers.RespondJSON(w, r, 200, user)
}

func (a *APIv1) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	defer a.services.Users().FlushCache()
	helpers.RespondJSON(w, r, 200, user)
}

func (a *APIv1) actionSetWakatimeApiKey(wakatimeSettings *SettingsPayload, user *models.User) actionResult {
	if wakatimeSettings.ApiKey == "" {
		return actionResult{http.StatusBadRequest, "", "invalid input: no or invalid api key", nil}
	}

	if !a.validateWakatimeKey(wakatimeSettings.ApiKey, wakatimeSettings.ApiUrl) {
		return actionResult{http.StatusBadRequest, "", "invalid input: failed to validate api key against wakatime server", nil}
	}

	if _, err := a.services.Users().SetWakatimeApiCredentials(user, wakatimeSettings.ApiKey, wakatimeSettings.ApiUrl); err != nil {
		return actionResult{http.StatusInternalServerError, "", config.ErrInternalServerError, nil}
	}

	return actionResult{http.StatusOK, "Wakatime API key set", "", nil}
}

func (a *APIv1) respondWithError(w http.ResponseWriter, r *http.Request, code int, message string) {
	helpers.RespondJSON(w, r, code, map[string]string{"error": message})
}

func (a *APIv1) validateWakatimeKey(apiKey string, baseUrl string) bool {
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

	if _, err = utils.RaiseForStatus(a.httpClient.Do(request)); err != nil {
		return false
	}

	return true
}
