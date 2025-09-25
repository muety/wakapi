package routes

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
)

const (
	// HTTP constants
	contentTypeHeader   = "Content-Type"
	applicationJSONType = "application/json"

	// Error message constants
	invalidJSONMessage     = "invalid JSON"
	internalServerErrorMsg = "internal server error"
)

type WebAuthnHandler struct {
	config   *conf.Config
	userSrvc services.IUserService
}

func NewWebAuthnHandler(userService services.IUserService) *WebAuthnHandler {
	return &WebAuthnHandler{
		config:   conf.Get(),
		userSrvc: userService,
	}
}

func (h *WebAuthnHandler) RegisterRoutes(router chi.Router) {
	if !h.config.Security.WebAuthnEnabled {
		return
	}

	webAuthnRouter := chi.NewRouter()
	webAuthnRouter.Use(httprate.LimitByRealIP(h.config.Security.GetLoginMaxRate()))

	// Registration endpoints (require authentication)
	webAuthnRouter.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
		r.Get("/credentials", h.GetCredentials)
		r.Post("/register/begin", h.PostRegisterBegin)
		r.Post("/register/finish", h.PostRegisterFinish)
		r.Delete("/credentials/{credentialId}", h.DeleteCredential)
	})

	// Authentication endpoints (public)
	webAuthnRouter.Post("/login/begin", h.PostLoginBegin)
	webAuthnRouter.Post("/login/finish", h.PostLoginFinish)

	router.Mount("/auth/webauthn", webAuthnRouter)
}

// PostRegisterBegin starts WebAuthn credential registration for authenticated users
func (h *WebAuthnHandler) PostRegisterBegin(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CredentialName string `json:"credentialName,omitempty"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, invalidJSONMessage, http.StatusBadRequest)
			return
		}
	}

	options, sessionData, err := h.userSrvc.WebAuthnBeginRegistration(user, req.CredentialName)
	if err != nil {
		conf.Log().Request(r).Error("WebAuthn registration begin failed", "error", err)
		http.Error(w, "failed to begin registration", http.StatusInternalServerError)
		return
	}

	// Store session data in the response for the frontend to include in the finish request
	response := map[string]interface{}{
		"options":     options,
		"sessionData": sessionData,
	}

	w.Header().Set(contentTypeHeader, applicationJSONType)
	json.NewEncoder(w).Encode(response)
}

// PostRegisterFinish completes WebAuthn credential registration
func (h *WebAuthnHandler) PostRegisterFinish(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	if user == nil {
		conf.Log().Request(r).Error("WebAuthn registration finish - no user")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conf.Log().Request(r).Info("WebAuthn registration finish started", "userID", user.ID)

	// Read the raw body first for debugging
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	var req struct {
		SessionData interface{} `json:"sessionData"`
		Response    interface{} `json:"response"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, invalidJSONMessage, http.StatusBadRequest)
		return
	}

	conf.Log().Request(r).Info("WebAuthn registration finish - JSON decoded successfully")

	// Convert response to JSON bytes for the service layer
	responseBytes, err := json.Marshal(req.Response)
	if err != nil {
		conf.Log().Request(r).Error("WebAuthn registration finish - response marshal failed", "error", err)
		http.Error(w, "invalid credential response", http.StatusBadRequest)
		return
	}

	conf.Log().Request(r).Info("WebAuthn registration finish - calling service layer")

	err = h.userSrvc.WebAuthnFinishRegistration(user, req.SessionData, responseBytes)
	if err != nil {
		conf.Log().Request(r).Error("WebAuthn registration finish failed", "error", err)
		http.Error(w, "failed to finish registration", http.StatusBadRequest)
		return
	}

	conf.Log().Request(r).Info("WebAuthn registration finish - success")

	w.Header().Set(contentTypeHeader, applicationJSONType)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "WebAuthn credential registered successfully",
	})
}

// PostLoginBegin starts WebAuthn authentication
func (h *WebAuthnHandler) PostLoginBegin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, invalidJSONMessage, http.StatusBadRequest)
		return
	}

	// Support both username-based and usernameless flows
	var options interface{}
	var sessionData interface{}
	var err error

	if req.Username != "" {
		// Username-based authentication
		options, sessionData, err = h.userSrvc.WebAuthnBeginLogin(req.Username)
	} else {
		// Usernameless authentication - use empty credential list to allow any registered credential
		options, sessionData, err = h.userSrvc.WebAuthnBeginLoginUsernameless()
	}

	if err != nil {
		conf.Log().Request(r).Error("WebAuthn login begin failed", "error", err)
		http.Error(w, "failed to begin login", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"options":     options,
		"sessionData": sessionData,
	}

	w.Header().Set(contentTypeHeader, applicationJSONType)
	json.NewEncoder(w).Encode(response)
}

// PostLoginFinish completes WebAuthn authentication and logs in the user
func (h *WebAuthnHandler) PostLoginFinish(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username    string      `json:"username,omitempty"`
		SessionData interface{} `json:"sessionData"`
		Response    interface{} `json:"response"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, invalidJSONMessage, http.StatusBadRequest)
		return
	}

	// Convert response to JSON bytes for the service layer
	responseBytes, err := json.Marshal(req.Response)
	if err != nil {
		http.Error(w, "invalid credential response", http.StatusBadRequest)
		return
	}

	var user *models.User
	if req.Username != "" {
		// Username-based authentication
		user, err = h.userSrvc.WebAuthnFinishLogin(req.Username, req.SessionData, responseBytes)
	} else {
		// Usernameless authentication - identify user from credential response
		user, err = h.userSrvc.WebAuthnFinishLoginUsernameless(req.SessionData, responseBytes)
	}

	if err != nil {
		conf.Log().Request(r).Error("WebAuthn login finish failed", "error", err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	// Create session cookie same way as traditional login
	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, user.ID)
	if err != nil {
		conf.Log().Request(r).Error("failed to encode secure cookie", "error", err)
		http.Error(w, internalServerErrorMsg, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, h.config.CreateCookie(models.AuthCookieKey, encoded))

	w.Header().Set(contentTypeHeader, applicationJSONType)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"redirectTo": h.config.Server.BasePath + "/summary",
	})
}

// GetCredentials returns the user's registered WebAuthn credentials
func (h *WebAuthnHandler) GetCredentials(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user with WebAuthn credentials loaded
	userWithCreds, err := h.userSrvc.GetUserById(user.ID)
	if err != nil {
		conf.Log().Request(r).Error("failed to get user", "error", err)
		http.Error(w, internalServerErrorMsg, http.StatusInternalServerError)
		return
	}

	// Access the user's custom WebAuthn credentials
	if userWithCreds.WebAuthn.CredentialsJSON == "" {
		w.Header().Set(contentTypeHeader, applicationJSONType)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"credentials": []interface{}{},
		})
		return
	}

	// Parse the credentials from JSON
	var credentials []models.WebAuthnCredential
	if err := json.Unmarshal([]byte(userWithCreds.WebAuthn.CredentialsJSON), &credentials); err != nil {
		conf.Log().Request(r).Error("failed to parse WebAuthn credentials", "error", err)
		http.Error(w, internalServerErrorMsg, http.StatusInternalServerError)
		return
	}

	// Format credentials for response
	var credentialsResponse []map[string]interface{}
	for _, cred := range credentials {
		// Handle zero time values gracefully
		var createdAtStr string
		if cred.CreatedAt.T().IsZero() {
			createdAtStr = "Unknown"
		} else {
			createdAtStr = cred.CreatedAt.T().Format("2006-01-02 15:04:05")
		}

		credentialsResponse = append(credentialsResponse, map[string]interface{}{
			"id":        cred.ID,
			"name":      cred.Name,
			"createdAt": createdAtStr,
		})
	}

	w.Header().Set(contentTypeHeader, applicationJSONType)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"credentials": credentialsResponse,
	})
}

// DeleteCredential handles deleting a WebAuthn credential
func (h *WebAuthnHandler) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	user, err := routeutils.CheckEffectiveUser(w, r, h.userSrvc, "current")
	if err != nil || user == nil {
		return
	}

	credentialId := chi.URLParam(r, "credentialId")
	if credentialId == "" {
		http.Error(w, "credential ID required", http.StatusBadRequest)
		return
	}

	// Get user with WebAuthn data
	userWithCreds, err := h.userSrvc.GetUserById(user.ID)
	if err != nil {
		conf.Log().Request(r).Error("failed to fetch user", "error", err)
		http.Error(w, internalServerErrorMsg, http.StatusInternalServerError)
		return
	}

	// Parse existing credentials
	var credentials []models.WebAuthnCredential
	if userWithCreds.WebAuthn.CredentialsJSON != "" {
		if err := json.Unmarshal([]byte(userWithCreds.WebAuthn.CredentialsJSON), &credentials); err != nil {
			conf.Log().Request(r).Error("failed to parse WebAuthn credentials", "error", err)
			http.Error(w, internalServerErrorMsg, http.StatusInternalServerError)
			return
		}
	}

	// Find and remove the credential
	var newCredentials []models.WebAuthnCredential
	found := false
	for _, cred := range credentials {
		if cred.ID != credentialId {
			newCredentials = append(newCredentials, cred)
		} else {
			found = true
		}
	}

	if !found {
		http.Error(w, "credential not found", http.StatusNotFound)
		return
	}

	// Update credentials in database
	newCredentialsJSON, err := json.Marshal(newCredentials)
	if err != nil {
		conf.Log().Request(r).Error("failed to marshal credentials", "error", err)
		http.Error(w, internalServerErrorMsg, http.StatusInternalServerError)
		return
	}

	// Use the user service to update the field
	if err := h.userSrvc.UpdateUserCredentials(userWithCreds, string(newCredentialsJSON)); err != nil {
		conf.Log().Request(r).Error("failed to update user credentials", "error", err)
		http.Error(w, internalServerErrorMsg, http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentTypeHeader, applicationJSONType)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":               true,
		"message":               "Credential deleted successfully",
		"remaining_credentials": len(newCredentials),
	})
}
