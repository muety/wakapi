package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/integrations/github"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
)

type SignUpParams struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	PasswordRepeat string `json:"password_repeat"`
}

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type OauthCode struct {
	Code string `json:"code"`
}

func (a *APIv1) Signin(w http.ResponseWriter, r *http.Request) {

	var params = &LoginParams{}
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)

	if err != nil || params.Email == "" || params.Password == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	params.Email = strings.ToLower(params.Email)
	user, err := a.services.Users().GetUserByEmail(params.Email)
	if err != nil || user == nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid credentials",
			"status":  http.StatusBadRequest,
		})
		return
	}

	if !utils.ComparePassword(user.Password, params.Password, a.config.Security.PasswordSalt) {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid credentials",
			"status":  http.StatusBadRequest,
		})
		return
	}

	user.LastLoggedInAt = models.CustomTime(time.Now())
	a.services.Users().Update(user)

	response, err := helpers.MakeAuthSuccessResponse(
		&helpers.AuthSuccessResponse{
			Message:   "Login Successful",
			User:      user,
			OauthUser: nil,
			IsNewUser: false,
		},
	)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Internal Server Error. Try again",
			"status":  http.StatusInternalServerError,
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (a *APIv1) GithubOauth(w http.ResponseWriter, r *http.Request) {
	var params = &OauthCode{}
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)

	if err != nil || params.Code == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	token, err := github.GetGithubAccessToken(context.Background(), params.Code)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid credentials",
			"status":  http.StatusBadRequest,
			"error":   err.Error(),
		})
		return
	}

	githubUser, err := github.GetGithubUser(token.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Error getting github user. Github might be down. Try again later.",
			"status":  http.StatusBadRequest,
			"error":   err.Error(),
		})
		return
	}

	primaryEmail, err := github.GetPrimaryGithubEmail(token.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Couldn't get a verified primary email for this github account	",
			"status":  http.StatusBadRequest,
			"error":   err.Error(),
		})
		return
	}

	providerId := strconv.Itoa(githubUser.ID)

	// get provider
	provider, err := a.services.OAuth().GetOne(models.UserOauth{Provider: "github", ProviderID: providerId})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Internal server error. Error fetching provider. Our services might be down. Try again later.",
			"status":  http.StatusBadRequest,
			"error":   err.Error(),
		})
		return
	}

	userOauthDetails := models.UserOauth{
		ID:         uuid.New().String(),
		Provider:   "github",
		ProviderID: providerId,
		Email:      &primaryEmail.Email,
		Handle:     &githubUser.Login,
		AvatarUrl:  &githubUser.AvatarURL,
	}

	findUser := func(db *gorm.DB, email string) (*models.User, error) {
		u := &models.User{}
		result := db.Where(models.User{Email: email}).First(u)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return nil, nil // No record found
			}
			return nil, result.Error
		}
		return u, nil
	}

	createUser := func(db *gorm.DB, email, password string) (*models.User, error) {
		hash, err := utils.HashPassword(password, conf.Get().Security.PasswordSalt)
		if err != nil {
			return nil, err
		}
		u := &models.User{
			ID:            uuid.New().String(),
			ApiKey:        uuid.New().String(),
			Email:         email,
			Password:      hash,
			IsAdmin:       false,
			EmailVerified: true,
		}
		result := db.Create(u)
		if err := result.Error; err != nil {
			return nil, err
		}
		return u, nil
	}

	findOrCreateUser := func(db *gorm.DB, email, password string) (*models.User, error) {
		user, err := findUser(db, email)
		if err != nil {
			return nil, err
		}
		if user != nil {
			return user, nil
		}
		return createUser(db, email, password)
	}

	if provider == nil {
		err := a.db.Transaction(func(tx *gorm.DB) error {
			// create a new user
			password, genErr := utils.GenerateRandomPassword(24)
			if genErr != nil {
				return genErr
			}

			user, err := findOrCreateUser(tx, primaryEmail.Email, password)

			if err != nil {
				return err
			}

			userOauthDetails.UserID = user.ID

			result := tx.Create(&userOauthDetails)
			if err := result.Error; err != nil {
				return errors.Wrap(err, "Error creating oauth entry")
			}

			return nil
		})
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": "Internal server error. Our services might be down. Try again later.",
				"status":  http.StatusBadRequest,
				"error":   err.Error(),
			})
			return
		}
	}

	oauthUser, err := a.services.Users().GetUserByEmail(primaryEmail.Email)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Internal server error. Our services might be down. Try again later.",
			"status":  http.StatusBadRequest,
			"error":   err.Error(),
		})
		return
	}

	oauthUser.LastLoggedInAt = models.CustomTime(time.Now())
	a.services.Users().Update(oauthUser)

	response, err := helpers.MakeAuthSuccessResponse(
		&helpers.AuthSuccessResponse{
			Message:   "Signup Successful",
			User:      oauthUser,
			OauthUser: &userOauthDetails,
			IsNewUser: provider == nil,
		},
	)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Internal server error.",
			"status":  http.StatusBadRequest,
			"error":   err.Error(),
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (a *APIv1) Signup(w http.ResponseWriter, r *http.Request) {

	var signup = &models.SignupJson{}
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(signup)

	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}
	if signup.Email == "" || signup.Password == "" || signup.PasswordRepeat == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Missing Parameters",
			"status":  http.StatusBadRequest,
		})
		return
	}
	if signup.Password != signup.PasswordRepeat {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Passwords do not match",
			"status":  http.StatusBadRequest,
		})
		return
	}

	if !a.config.IsDev() && !a.config.Security.AllowSignup && (!a.config.Security.InviteCodes || signup.InviteCode == "") {
		helpers.RespondJSON(w, r, http.StatusForbidden, map[string]interface{}{
			"message": "Registration is disabled on this server",
			"status":  http.StatusForbidden,
		})
		return
	}

	signup.Email = strings.ToLower(signup.Email)
	existing_user, _ := a.services.Users().GetUserByEmail(signup.Email)
	if existing_user != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "An account already exists with this email.",
			"status":  http.StatusBadRequest,
		})
		return
	}

	user, err := a.services.Users().Create(&models.Signup{Email: signup.Email, Password: signup.Password})
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Internal Server Error. Failed to create new user",
			"status":  http.StatusInternalServerError,
		})
		return
	}

	user.LastLoggedInAt = models.CustomTime(time.Now())
	a.services.Users().Update(user)

	response, err := helpers.MakeAuthSuccessResponse(
		&helpers.AuthSuccessResponse{
			Message:   "Signup Successful",
			User:      user,
			OauthUser: nil,
			IsNewUser: true,
		},
	)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Internal Server Error. Try again",
			"status":  http.StatusInternalServerError,
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (a *APIv1) ValidateAuthToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	if token == "" {
		helpers.RespondJSON(w, r, http.StatusUnauthorized, map[string]interface{}{
			"message": "Unauthorized",
			"status":  http.StatusUnauthorized,
		})
	}

	claim, err := utils.GetTokenClaims(token, a.config.Security.JWT_SECRET)
	if err != nil || claim.UID == "" {
		helpers.RespondJSON(w, r, http.StatusUnauthorized, map[string]interface{}{
			"message": "Unauthorized: Invalid or expired token",
			"status":  http.StatusUnauthorized,
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusAccepted, map[string]interface{}{
		"message": "Token is valid",
		"status":  http.StatusAccepted,
	})
}

func (a *APIv1) GetApiKey(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)

	if user == nil {
		helpers.RespondJSON(w, r, http.StatusUnauthorized, map[string]interface{}{
			"message": "Unauthorized",
			"status":  http.StatusUnauthorized,
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusAccepted, map[string]interface{}{
		"status": http.StatusAccepted,
		"apiKey": user.ApiKey,
	})
}

// this was a bad idea - careful not to use it
func (a *APIv1) RefreshApiKey(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	if user == nil {
		helpers.RespondJSON(w, r, http.StatusUnauthorized, map[string]interface{}{
			"message": "Unauthorized",
			"status":  http.StatusUnauthorized,
		})
		return
	}

	user.ApiKey = a.services.Users().MakeApiKey()
	user, err := a.services.Users().Update(user)

	if err != nil {
		helpers.RespondJSON(w, r, http.StatusUnauthorized, map[string]interface{}{
			"message": "Error generating new API key. Try again",
			"status":  http.StatusInternalServerError,
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusAccepted, map[string]interface{}{
		"status": http.StatusAccepted,
		"apiKey": user.ApiKey,
	})
}

func (a *APIv1) handlePasswordReset(user *models.User) error {
	updatedUser, err := a.services.Users().GenerateResetToken(user)
	if err != nil {
		return err
	}

	go a.sendPasswordResetEmail(updatedUser)
	return nil
}

func (a *APIv1) sendPasswordResetEmail(user *models.User) {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", a.config.Server.GetFrontendUri(), user.ResetToken)

	if err := a.mailService.SendPasswordReset(user, resetLink); err != nil {
		conf.Log().Error("failed to send password reset mail",
			"userID", user.ID,
			"error", err,
		)
		conf.Log().Info("Password reset link", resetLink, "userID")
		return
	}

	slog.Info("sent password reset mail", "userID", user.ID)
}

func (a *APIv1) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var resetRequest = &models.ResetPasswordRequest{}
	if err := json.NewDecoder(r.Body).Decode(resetRequest); err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	if resetRequest.Email == "" || !models.ValidateEmail(resetRequest.Email) {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Email is invalid",
			"status":  http.StatusBadRequest,
		})
		return
	}

	user, err := a.services.Users().GetUserByEmail(resetRequest.Email)
	if err != nil || user == nil {
		conf.Log().Request(r).Warn("password reset requested for unregistered address",
			"email", resetRequest.Email,
		)
	} else {
		if err := a.handlePasswordReset(user); err != nil {
			helpers.RespondJSON(w, r, http.StatusInternalServerError, map[string]interface{}{
				"message": "Failed to generate password reset token",
				"status":  http.StatusInternalServerError,
			})
			return
		}
	}

	helpers.RespondJSON(w, r, http.StatusAccepted, map[string]interface{}{
		"message": "An e-mail was sent to you. Follow the instructions to reset your password",
		"status":  http.StatusAccepted,
	})
}
