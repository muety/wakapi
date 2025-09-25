package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid/v5"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
)

type UserService struct {
	config              *config.Config
	cache               *cache.Cache
	eventBus            *hub.Hub
	keyValueService     IKeyValueService
	mailService         IMailService
	repository          repositories.IUserRepository
	currentOnlineUsers  *cache.Cache
	countersInitialized atomic.Bool
	webAuthn            *webauthn.WebAuthn
}

func NewUserService(keyValueService IKeyValueService, mailService IMailService, userRepo repositories.IUserRepository) *UserService {
	cfg := config.Get()
	srv := &UserService{
		config:             cfg,
		eventBus:           config.EventBus(),
		cache:              cache.New(1*time.Hour, 2*time.Hour),
		keyValueService:    keyValueService,
		mailService:        mailService,
		repository:         userRepo,
		currentOnlineUsers: cache.New(models.DefaultHeartbeatsTimeout, 1*time.Minute),
	}

	// Initialize WebAuthn if enabled
	if cfg.Security.WebAuthnEnabled {
		wconfig := &webauthn.Config{
			RPDisplayName: "Wakapi",
			RPID:          cfg.Security.WebAuthnRPID,
			RPOrigins:     []string{cfg.Security.WebAuthnRPOrigin},
		}

		webAuthn, err := webauthn.New(wconfig)
		if err != nil {
			slog.Error("failed to initialize WebAuthn", "error", err)
		} else {
			srv.webAuthn = webAuthn
		}
	}

	sub1 := srv.eventBus.Subscribe(0, config.EventWakatimeFailure)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			user := m.Fields[config.FieldUser].(*models.User)
			n := m.Fields[config.FieldPayload].(int)

			slog.Warn("resetting wakatime api key for user due to too many failures", "userID", user.ID, "failureCount", n)

			if _, err := srv.SetWakatimeApiCredentials(user, "", ""); err != nil {
				config.Log().Error("failed to set wakatime api key for user", "userID", user.ID)
			}

			if user.Email != "" {
				if err := mailService.SendWakatimeFailureNotification(user, n); err != nil {
					config.Log().Error("failed to send wakatime failure notification mail to user", "userID", user.ID)
				} else {
					slog.Info("sent wakatime connection failure mail", "userID", user.ID)
				}
			}
		}
	}(&sub1)

	sub2 := srv.eventBus.Subscribe(0, config.EventHeartbeatCreate)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			heartbeat := m.Fields[config.FieldPayload].(*models.Heartbeat)
			if time.Now().Sub(heartbeat.Time.T()) > models.DefaultHeartbeatsTimeout {
				continue
			}
			srv.currentOnlineUsers.SetDefault(heartbeat.UserID, true)
		}
	}(&sub2)

	return srv
}

func (srv *UserService) GetUserById(userId string) (*models.User, error) {
	if userId == "" {
		return nil, errors.New("user id must not be empty")
	}

	if u, ok := srv.cache.Get(userId); ok {
		return u.(*models.User), nil
	}

	u, err := srv.repository.FindOne(models.User{ID: userId})
	if err != nil {
		return nil, err
	}

	srv.cache.SetDefault(u.ID, u)
	return u, nil
}

func (srv *UserService) GetUserByKey(key string) (*models.User, error) {
	if key == "" {
		return nil, errors.New("key must not be empty")
	}

	if u, ok := srv.cache.Get(key); ok {
		return u.(*models.User), nil
	}

	u, err := srv.repository.FindOne(models.User{ApiKey: key})
	if err != nil {
		return nil, err
	}

	srv.cache.SetDefault(u.ID, u)
	return u, nil
}

func (srv *UserService) GetUserByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email must not be empty")
	}
	return srv.repository.FindOne(models.User{Email: email})
}

func (srv *UserService) GetUserByResetToken(resetToken string) (*models.User, error) {
	if resetToken == "" {
		return nil, errors.New("reset token must not be empty")
	}
	return srv.repository.FindOne(models.User{ResetToken: resetToken})
}

func (srv *UserService) GetUserByStripeCustomerId(customerId string) (*models.User, error) {
	if customerId == "" {
		return nil, errors.New("customer id must not be empty")
	}
	return srv.repository.FindOne(models.User{StripeCustomerId: customerId})
}

func (srv *UserService) GetAll() ([]*models.User, error) {
	return srv.repository.GetAll()
}

func (srv *UserService) GetAllMapped() (map[string]*models.User, error) {
	users, err := srv.repository.GetAll()
	if err != nil {
		return nil, err
	}
	return srv.MapUsersById(users), nil
}

func (srv *UserService) GetMany(ids []string) ([]*models.User, error) {
	return srv.repository.GetMany(ids)
}

func (srv *UserService) GetManyMapped(ids []string) (map[string]*models.User, error) {
	users, err := srv.repository.GetMany(ids)
	if err != nil {
		return nil, err
	}
	return srv.MapUsersById(users), nil
}

func (srv *UserService) GetAllByReports(reportsEnabled bool) ([]*models.User, error) {
	return srv.repository.GetAllByReports(reportsEnabled)
}

func (srv *UserService) GetAllByLeaderboard(leaderboardEnabled bool) ([]*models.User, error) {
	return srv.repository.GetAllByLeaderboard(leaderboardEnabled)
}

func (srv *UserService) GetActive(exact bool) ([]*models.User, error) {
	minDate := time.Now().AddDate(0, 0, -1*srv.config.App.InactiveDays)
	if !exact {
		minDate = datetime.BeginOfHour(minDate)
	}

	cacheKey := fmt.Sprintf("%s--active", minDate.String())
	if u, ok := srv.cache.Get(cacheKey); ok {
		return u.([]*models.User), nil
	}

	results, err := srv.repository.GetByLastActiveAfter(minDate)
	if err != nil {
		return nil, err
	}

	srv.cache.SetDefault(cacheKey, results)
	return results, nil
}

func (srv *UserService) Count() (int64, error) {
	return srv.repository.Count()
}

func (srv *UserService) CountCurrentlyOnline() (int, error) {
	if !srv.countersInitialized.Load() {
		minDate := time.Now().Add(-1 * models.DefaultHeartbeatsTimeout)
		result, err := srv.repository.GetByLastActiveAfter(minDate)
		if err != nil {
			return 0, err
		}
		for _, r := range result {
			srv.currentOnlineUsers.SetDefault(r.ID, true)
		}
		srv.countersInitialized.Store(true)
	}

	return srv.currentOnlineUsers.ItemCount(), nil
}

func (srv *UserService) CreateOrGet(signup *models.Signup, isAdmin bool) (*models.User, bool, error) {
	u := &models.User{
		ID:        signup.Username,
		ApiKey:    uuid.Must(uuid.NewV4()).String(),
		Email:     signup.Email,
		Location:  signup.Location,
		Password:  signup.Password,
		IsAdmin:   isAdmin,
		InvitedBy: signup.InvitedBy,
	}

	if hash, err := utils.HashPassword(u.Password, srv.config.Security.PasswordSalt); err != nil {
		return nil, false, err
	} else {
		u.Password = hash
	}

	return srv.repository.InsertOrGet(u)
}

func (srv *UserService) Update(user *models.User) (*models.User, error) {
	srv.FlushUserCache(user.ID)
	srv.notifyUpdate(user)
	return srv.repository.Update(user)
}

func (srv *UserService) ChangeUserId(user *models.User, newUserId string) (*models.User, error) {
	if !srv.checkUpdateCascade() {
		return nil, errors.New("sqlite database too old to perform user id change consistently")
	}

	// https://github.com/muety/wakapi/issues/739
	oldUserId := user.ID
	defer srv.FlushUserCache(oldUserId)

	// TODO: make this transactional somehow
	userNew, err := srv.repository.UpdateField(user, "id", newUserId)
	if err != nil {
		return nil, err
	}

	err = srv.keyValueService.ReplaceKeySuffix(fmt.Sprintf("_%s", oldUserId), fmt.Sprintf("_%s", newUserId))
	if err != nil {
		// try roll back "manually"
		config.Log().Error("failed to update key string values during user id change, trying to roll back manually", "userID", oldUserId, "newUserID", newUserId)
		if _, err := srv.repository.UpdateField(userNew, "id", oldUserId); err != nil {
			config.Log().Error("manual user id rollback failed", "userID", oldUserId, "newUserID", newUserId)
		}
		return nil, err
	}

	config.Log().Info("user changed their user id", "userID", oldUserId, "newUserID", newUserId)

	return userNew, err
}

func (srv *UserService) ResetApiKey(user *models.User) (*models.User, error) {
	srv.FlushUserCache(user.ID)
	user.ApiKey = uuid.Must(uuid.NewV4()).String()
	return srv.Update(user)
}

func (srv *UserService) SetWakatimeApiCredentials(user *models.User, apiKey string, apiUrl string) (*models.User, error) {
	srv.FlushUserCache(user.ID)

	if apiKey != user.WakatimeApiKey {
		if u, err := srv.repository.UpdateField(user, "wakatime_api_key", apiKey); err != nil {
			return u, err
		}
	}

	if apiUrl != user.WakatimeApiUrl {
		return srv.repository.UpdateField(user, "wakatime_api_url", apiUrl)
	}

	return user, nil
}

func (srv *UserService) GenerateResetToken(user *models.User) (*models.User, error) {
	return srv.repository.UpdateField(user, "reset_token", uuid.Must(uuid.NewV4()))
}

func (srv *UserService) Delete(user *models.User) error {
	srv.FlushUserCache(user.ID)

	user.ReportsWeekly = false
	srv.notifyUpdate(user)
	srv.notifyDelete(user)

	return srv.repository.Delete(user)
}

func (srv *UserService) MapUsersById(users []*models.User) map[string]*models.User {
	return convertor.ToMap[*models.User, string, *models.User](users, func(u *models.User) (string, *models.User) {
		return u.ID, u
	})
}

func (srv *UserService) FlushCache() {
	srv.cache.Flush()
}

func (srv *UserService) FlushUserCache(userId string) {
	srv.cache.Delete(userId)
}

func (srv *UserService) notifyUpdate(user *models.User) {
	srv.eventBus.Publish(hub.Message{
		Name:   config.EventUserUpdate,
		Fields: map[string]interface{}{config.FieldPayload: user},
	})
}

func (srv *UserService) notifyDelete(user *models.User) {
	srv.eventBus.Publish(hub.Message{
		Name:   config.EventUserDelete,
		Fields: map[string]interface{}{config.FieldPayload: user},
	})
}

func (srv *UserService) checkUpdateCascade() bool {
	if dialector := srv.repository.GetDialector(); dialector == "sqlite" || dialector == "sqlite3" {
		ddl, _ := srv.repository.GetTableDDLSqlite("heartbeats")
		return strings.Contains(ddl, "ON UPDATE CASCADE")
	}
	return true
}

// WebAuthn methods

func (srv *UserService) WebAuthnBeginRegistration(user *models.User, credentialName string) (interface{}, interface{}, error) {
	if srv.webAuthn == nil {
		return nil, nil, errors.New("WebAuthn is not enabled")
	}

	options, sessionData, err := srv.webAuthn.BeginRegistration(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	// Store session data in cache instead of database to avoid serialization issues
	sessionBytes, err := json.Marshal(sessionData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Use cache with 5 minute expiry for session data
	cacheKey := fmt.Sprintf("webauthn_session_%s", user.ID)
	srv.cache.Set(cacheKey, string(sessionBytes), 5*time.Minute)

	// Store credential name separately in cache for use during finish registration
	credentialNameKey := fmt.Sprintf("webauthn_credential_name_%s", user.ID)
	if credentialName != "" {
		srv.cache.Set(credentialNameKey, credentialName, 5*time.Minute)
	} else {
		// Use default name if none provided
		defaultName := fmt.Sprintf("Security Key %d", len(user.GetWebAuthnCredentials())+1)
		srv.cache.Set(credentialNameKey, defaultName, 5*time.Minute)
	}

	return options, string(sessionBytes), nil
}

func (srv *UserService) WebAuthnFinishRegistration(user *models.User, sessionData interface{}, credentialCreationResponse interface{}) error {
	if srv.webAuthn == nil {
		return errors.New("WebAuthn is not enabled")
	}

	fmt.Printf("DEBUG: WebAuthnFinishRegistration - sessionData type: %T, value: %v\n", sessionData, sessionData)
	fmt.Printf("DEBUG: WebAuthnFinishRegistration - credentialCreationResponse type: %T\n", credentialCreationResponse)

	// Check if sessionData is nil
	if sessionData == nil {
		fmt.Printf("DEBUG: WebAuthnFinishRegistration - sessionData is nil\n")
		return errors.New("session data is missing")
	}

	// Parse session data
	sessionDataStruct := &webauthn.SessionData{}
	sessionDataStr, ok := sessionData.(string)
	if !ok {
		fmt.Printf("DEBUG: WebAuthnFinishRegistration - sessionData is not a string: %T\n", sessionData)
		return errors.New("session data must be a string")
	}
	
	if err := json.Unmarshal([]byte(sessionDataStr), sessionDataStruct); err != nil {
		fmt.Printf("DEBUG: WebAuthnFinishRegistration - session unmarshal failed: %v\n", err)
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	fmt.Printf("DEBUG: WebAuthnFinishRegistration - session data parsed successfully\n")

	// Verify session exists in cache
	cacheKey := fmt.Sprintf("webauthn_session_%s", user.ID)
	if _, found := srv.cache.Get(cacheKey); !found {
		fmt.Printf("DEBUG: WebAuthnFinishRegistration - session not found in cache or expired\n")
		return errors.New("session expired or invalid")
	}

	fmt.Printf("DEBUG: WebAuthnFinishRegistration - session valid\n")

	// Parse credential creation response
	ccr, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(credentialCreationResponse.([]byte)))
	if err != nil {
		fmt.Printf("DEBUG: WebAuthnFinishRegistration - parse credential creation response failed: %v\n", err)
		return fmt.Errorf("failed to parse credential creation response: %w", err)
	}

	fmt.Printf("DEBUG: WebAuthnFinishRegistration - credential creation response parsed\n")

	credential, err := srv.webAuthn.CreateCredential(user, *sessionDataStruct, ccr)
	if err != nil {
		fmt.Printf("DEBUG: WebAuthnFinishRegistration - create credential failed: %v\n", err)
		return fmt.Errorf("failed to create credential: %w", err)
	}

	fmt.Printf("DEBUG: WebAuthnFinishRegistration - credential created successfully\n")

	// Retrieve credential name from cache
	credentialNameKey := fmt.Sprintf("webauthn_credential_name_%s", user.ID)
	credentialName := "Security Key"
	if cachedName, found := srv.cache.Get(credentialNameKey); found {
		if nameStr, ok := cachedName.(string); ok {
			credentialName = nameStr
		}
	}

	// Convert to our model and add to user
	webAuthnCred := &models.WebAuthnCredential{
		ID:              fmt.Sprintf("cred_%d", time.Now().Unix()),
		Name:            credentialName,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport:       credential.Transport,
		Flags: struct {
			UserPresent    bool `json:"user_present"`
			UserVerified   bool `json:"user_verified"`
			BackupEligible bool `json:"backup_eligible"`
			BackupState    bool `json:"backup_state"`
		}{
			UserPresent:    credential.Flags.UserPresent,
			UserVerified:   credential.Flags.UserVerified,
			BackupEligible: credential.Flags.BackupEligible,
			BackupState:    credential.Flags.BackupState,
		},
		Authenticator: struct {
			AAGUID       []byte `json:"aaguid" gorm:"type:blob"`
			SignCount    uint32 `json:"sign_count"`
			CloneWarning bool   `json:"clone_warning"`
		}{
			AAGUID:       credential.Authenticator.AAGUID,
			SignCount:    credential.Authenticator.SignCount,
			CloneWarning: credential.Authenticator.CloneWarning,
		},
		CreatedAt: models.CustomTime(time.Now()),
		UpdatedAt: models.CustomTime(time.Now()),
	}

	user.AddWebAuthnCredential(webAuthnCred)

	// Clear session from cache
	srv.cache.Delete(cacheKey)
	
	// Clear credential name from cache
	credentialNameKey = fmt.Sprintf("webauthn_credential_name_%s", user.ID)
	srv.cache.Delete(credentialNameKey)

	// Update credentials directly using UpdateField to avoid WebAuthn interface serialization issues
	fmt.Printf("DEBUG: WebAuthnFinishRegistration - about to save credentials using UpdateField\n")
	_, err = srv.repository.UpdateField(user, "webauthn_credentials", user.WebAuthn.CredentialsJSON)
	if err != nil {
		fmt.Printf("DEBUG: WebAuthnFinishRegistration - failed to save credentials: %v\n", err)
		return fmt.Errorf("failed to save credential: %w", err)
	}
	fmt.Printf("DEBUG: WebAuthnFinishRegistration - credentials saved successfully\n")

	// Clear cache
	srv.cache.Delete(user.ID)

	return nil
}

func (srv *UserService) WebAuthnBeginLogin(username string) (interface{}, interface{}, error) {
	if srv.webAuthn == nil {
		return nil, nil, errors.New("WebAuthn is not enabled")
	}

	user, err := srv.GetUserById(username)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.HasWebAuthnCredentials() {
		return nil, nil, errors.New("user has no WebAuthn credentials registered")
	}

	options, sessionData, err := srv.webAuthn.BeginLogin(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin login: %w", err)
	}

	// Store session data in cache instead of database
	sessionBytes, err := json.Marshal(sessionData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Use cache with 5 minute expiry
	cacheKey := fmt.Sprintf("webauthn_session_%s", user.ID)
	srv.cache.Set(cacheKey, string(sessionBytes), 5*time.Minute)

	return options, string(sessionBytes), nil
}

func (srv *UserService) WebAuthnFinishLogin(username string, sessionData interface{}, credentialAssertionResponse interface{}) (*models.User, error) {
	if srv.webAuthn == nil {
		return nil, errors.New("WebAuthn is not enabled")
	}

	user, err := srv.GetUserById(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Parse session data
	sessionDataStruct := &webauthn.SessionData{}
	if err := json.Unmarshal([]byte(sessionData.(string)), sessionDataStruct); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// Verify session exists in cache
	cacheKey := fmt.Sprintf("webauthn_session_%s", user.ID)
	if _, found := srv.cache.Get(cacheKey); !found {
		return nil, errors.New("session expired or invalid")
	}

	// Parse credential assertion response
	car, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(credentialAssertionResponse.([]byte)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse credential assertion response: %w", err)
	}

	credential, err := srv.webAuthn.ValidateLogin(user, *sessionDataStruct, car)
	if err != nil {
		return nil, fmt.Errorf("failed to validate login: %w", err)
	}

	// Update credential sign count
	credentials := user.GetWebAuthnCredentials()
	credentialsUpdated := false
	for _, cred := range credentials {
		if bytes.Equal(cred.CredentialID, credential.ID) {
			cred.Authenticator.SignCount = credential.Authenticator.SignCount
			cred.Authenticator.CloneWarning = credential.Authenticator.CloneWarning
			cred.UpdatedAt = models.CustomTime(time.Now())
			user.UpdateWebAuthnCredential(credential.ID, &cred)
			credentialsUpdated = true
			break
		}
	}

	// Save updated credentials to database if they were changed
	if credentialsUpdated {
		err = srv.UpdateUserCredentials(user, user.WebAuthn.CredentialsJSON)
		if err != nil {
			config.Log().Error("failed to update WebAuthn credentials after login", "error", err)
			// Don't fail login for this
		}
	}

	// Clear session from cache
	srv.cache.Delete(cacheKey)

	// Update last login time and save only specific fields to avoid GORM serialization issues
	user.LastLoggedInAt = models.CustomTime(time.Now())
	
	// Update only the specific fields we need to avoid GORM trying to serialize WebAuthn methods
	_, err = srv.repository.UpdateField(user, "last_logged_in_at", user.LastLoggedInAt)
	if err != nil {
		config.Log().Error("failed to update user last login time after WebAuthn login", "error", err)
		// Don't fail login for this
	}

	// Clear cache
	srv.cache.Delete(user.ID)

	return user, nil
}

// UpdateUserCredentials updates the WebAuthn credentials for a user
func (srv *UserService) UpdateUserCredentials(user *models.User, credentialsJSON string) error {
	_, err := srv.repository.UpdateField(user, "webauthn_credentials", credentialsJSON)
	if err != nil {
		return err
	}
	// Clear cache after updating credentials
	srv.cache.Delete(user.ID)
	return nil
}

// WebAuthnBeginLoginUsernameless starts usernameless WebAuthn authentication
func (srv *UserService) WebAuthnBeginLoginUsernameless() (interface{}, interface{}, error) {
	if srv.webAuthn == nil {
		return nil, nil, errors.New("WebAuthn is not enabled")
	}

	// For usernameless flow, we can't use srv.webAuthn.BeginLogin() with a user
	// because it expects the user to have credentials. Instead, we'll create
	// the assertion options manually to allow any credential for this domain.
	
	// Generate a challenge
	challenge, err := protocol.CreateChallenge()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create challenge: %w", err)
	}

	// Create the assertion options manually for usernameless flow
	// This mimics what webauthn.BeginLogin() does but without user-specific credentials
	assertionOptions := protocol.PublicKeyCredentialRequestOptions{
		Challenge:        challenge,
		Timeout:          60000, // 60 seconds
		RelyingPartyID:   srv.webAuthn.Config.RPID,
		UserVerification: protocol.VerificationRequired,
		// AllowCredentials is empty to allow any registered credential
	}
	
	// Wrap in the expected structure
	options := protocol.CredentialAssertion{
		Response: assertionOptions,
	}
	
	// Create session data for validation later
	sessionData := &webauthn.SessionData{
		Challenge:            string(challenge),
		UserID:               []byte{}, // Empty for usernameless
		AllowedCredentialIDs: [][]byte{}, // Empty for usernameless
		UserVerification:     protocol.VerificationRequired,
	}

	// Store session data in cache with a generic key for usernameless flows
	sessionBytes, err := json.Marshal(sessionData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Use a temporary session key that will be replaced when we identify the user
	cacheKey := fmt.Sprintf("webauthn_usernameless_session_%d", time.Now().UnixNano())
	srv.cache.Set(cacheKey, string(sessionBytes), 5*time.Minute)

	// Store the cache key in the session data so we can retrieve it later
	sessionDataWithKey := map[string]interface{}{
		"session":  sessionData,
		"cacheKey": cacheKey,
	}

	return options, sessionDataWithKey, nil
}

// WebAuthnFinishLoginUsernameless completes usernameless WebAuthn authentication
func (srv *UserService) WebAuthnFinishLoginUsernameless(sessionData interface{}, credentialAssertionResponse interface{}) (*models.User, error) {
	if srv.webAuthn == nil {
		return nil, errors.New("WebAuthn is not enabled")
	}

	// Parse the session data to get both the webauthn session and cache key
	sessionMap, ok := sessionData.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid session data format")
	}

	cacheKey, ok := sessionMap["cacheKey"].(string)
	if !ok {
		return nil, errors.New("missing cache key in session data")
	}

	// Verify session exists in cache
	if _, found := srv.cache.Get(cacheKey); !found {
		return nil, errors.New("session expired or invalid")
	}

	// Parse credential assertion response first to identify the user
	car, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(credentialAssertionResponse.([]byte)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse credential assertion response: %w", err)
	}

	// For usernameless flow, we need to identify the user from the credential response
	// The userHandle in the response should contain the user ID
	if car.Response.UserHandle == nil || len(car.Response.UserHandle) == 0 {
		return nil, errors.New("usernameless authentication requires a user handle in the credential response")
	}

	// The user handle should contain the user ID
	userID := string(car.Response.UserHandle)
	user, err := srv.GetUserById(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found from credential: %w", err)
	}

	// Parse the actual WebAuthn session data and fix the UserID for validation
	sessionDataStruct := &webauthn.SessionData{}
	sessionBytes, err := json.Marshal(sessionMap["session"])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session data: %w", err)
	}
	
	if err := json.Unmarshal(sessionBytes, sessionDataStruct); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// Update the session data with the correct user ID for validation
	sessionDataStruct.UserID = []byte(userID)

	// Now validate the credential using the identified user and updated session
	credential, err := srv.webAuthn.ValidateLogin(user, *sessionDataStruct, car)
	if err != nil {
		return nil, fmt.Errorf("failed to validate usernameless login: %w", err)
	}

	// Update credential sign count
	credentials := user.GetWebAuthnCredentials()
	credentialsUpdated := false
	for _, cred := range credentials {
		if bytes.Equal(cred.CredentialID, credential.ID) {
			cred.Authenticator.SignCount = credential.Authenticator.SignCount
			cred.Authenticator.CloneWarning = credential.Authenticator.CloneWarning
			cred.UpdatedAt = models.CustomTime(time.Now())
			user.UpdateWebAuthnCredential(credential.ID, &cred)
			credentialsUpdated = true
			break
		}
	}

	// Save updated credentials to database if they were changed
	if credentialsUpdated {
		err = srv.UpdateUserCredentials(user, user.WebAuthn.CredentialsJSON)
		if err != nil {
			config.Log().Error("failed to update WebAuthn credentials after usernameless login", "error", err)
			// Don't fail login for this
		}
	}

	// Clear session from cache
	srv.cache.Delete(cacheKey)

	// Update last login time and save only specific fields to avoid GORM serialization issues
	user.LastLoggedInAt = models.CustomTime(time.Now())
	
	// Update only the specific fields we need to avoid GORM trying to serialize WebAuthn methods
	_, err = srv.repository.UpdateField(user, "last_logged_in_at", user.LastLoggedInAt)
	if err != nil {
		config.Log().Error("failed to update user last login time after usernameless WebAuthn login", "error", err)
		// Don't fail login for this
	}

	// Clear user cache
	srv.cache.Delete(user.ID)

	return user, nil
}
