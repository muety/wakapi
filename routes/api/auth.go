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

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/integrations/github"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

var JWT_TOKEN_DURATION = time.Hour * 24

type AuthApiHandler struct {
	db                 *gorm.DB
	config             *conf.Config
	userService        services.IUserService
	mailService        services.IMailService
	oauthUserService   services.IUserOauthService
	aggregationService services.IAggregationService
	summaryService     services.ISummaryService
	otpService         *services.OTPService
}

func NewAuthApiHandler(db *gorm.DB, userService services.IUserService, oauthUserService services.IUserOauthService, mailService services.IMailService, aggregationService services.IAggregationService, summaryService services.ISummaryService) *AuthApiHandler {
	otpService := services.NewOTPService(db)
	return &AuthApiHandler{
		db:                 db,
		userService:        userService,
		oauthUserService:   oauthUserService,
		config:             conf.Get(),
		mailService:        mailService,
		aggregationService: aggregationService,
		summaryService:     summaryService,
		otpService:         otpService,
	}
}

func (h *AuthApiHandler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/oauth/github", h.GithubOauth)
	r.Post("/auth/login", h.Signin)
	r.Get("/auth/validate", h.ValidateAuthToken)
	r.Post("/auth/forgot-password", h.ForgotPassword)
	r.Post("/auth/otp/create", services.CreateOTPHandler(h.otpService))
	r.Post("/auth/otp/verify", services.VerifyOTPHandler(h.otpService))

	r.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userService).Handler)
		r.Get("/auth/api-key", h.GetApiKey)
		r.Post("/auth/api-key/refresh", h.RefreshApiKey)
	})
}

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

func (h *AuthApiHandler) regenerateSummaries(user *models.User) error {
	slog.Info("clearing summaries for user", "userID", user.ID)
	if err := h.summaryService.DeleteByUser(user.ID); err != nil {
		conf.Log().Error("failed to clear summaries", "error", err)
		return err
	}

	if err := h.aggregationService.AggregateSummaries(datastructure.New(user.ID)); err != nil {
		conf.Log().Error("failed to regenerate summaries", "error", err)
		return err
	}

	return nil
}

func (h *AuthApiHandler) Signin(w http.ResponseWriter, r *http.Request) {

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
	user, err := h.userService.GetUserByEmail(params.Email)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusNotFound)
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid credentials",
			"status":  http.StatusBadRequest,
		})
		return
	}

	// go h.regenerateSummaries(user)

	if !utils.ComparePassword(user.Password, params.Password, h.config.Security.PasswordSalt) {
		w.WriteHeader(http.StatusUnauthorized)
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid credentials",
			"status":  http.StatusBadRequest,
		})
		return
	}

	user.LastLoggedInAt = models.CustomTime(time.Now())
	h.userService.Update(user)

	token, _, err := MakeLoginJWT(user.ID, h.config)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Internal Server Error. Try again",
			"status":  http.StatusInternalServerError,
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusCreated, makeAuthSuccessResponse(
		&MakeAuthSuccessResponse{
			token:   token,
			message: "Login Successful",
			user:    user,
		},
	))
}

func (h *AuthApiHandler) GithubOauth(w http.ResponseWriter, r *http.Request) {
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
	provider, err := h.oauthUserService.GetOne(models.UserOauth{Provider: "github", ProviderID: providerId})
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
		Provider:   "github",
		ProviderID: providerId,
		Email:      &primaryEmail.Email,
		// UserID:     user.ID,
		Handle:    &githubUser.Login,
		AvatarUrl: &githubUser.AvatarURL,
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
			ID:            uuid.NewV4().String(),
			ApiKey:        uuid.NewV4().String(),
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
		err := h.db.Transaction(func(tx *gorm.DB) error {
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

	oauthUser, err := h.userService.GetUserByEmail(primaryEmail.Email)
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
	h.userService.Update(oauthUser)

	accessToken, _, err := MakeLoginJWT(oauthUser.ID, h.config)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Internal Server Error. Try again",
			"status":  http.StatusInternalServerError,
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusCreated, makeAuthSuccessResponse(
		&MakeAuthSuccessResponse{
			token:     accessToken,
			message:   "Signup Successful",
			user:      oauthUser,
			oauthUser: &userOauthDetails,
		},
	))
}

func (h *AuthApiHandler) Signup(w http.ResponseWriter, r *http.Request) {

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

	if !h.config.IsDev() && !h.config.Security.AllowSignup && (!h.config.Security.InviteCodes || signup.InviteCode == "") {
		helpers.RespondJSON(w, r, http.StatusForbidden, map[string]interface{}{
			"message": "Registration is disabled on this server",
			"status":  http.StatusForbidden,
		})
		return
	}

	signup.Email = strings.ToLower(signup.Email)
	existing_user, err := h.userService.GetUserByEmail(signup.Email)
	if existing_user != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "An account already exists with this email.",
			"status":  http.StatusBadRequest,
		})
		return
	}

	user, err := h.userService.Create(&models.Signup{Email: signup.Email, Password: signup.Password})
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Internal Server Error. Failed to create new user",
			"status":  http.StatusInternalServerError,
		})
		return
	}

	user.LastLoggedInAt = models.CustomTime(time.Now())
	h.userService.Update(user)

	token, _, err := MakeLoginJWT(user.ID, h.config)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Internal Server Error. Try again",
			"status":  http.StatusInternalServerError,
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusCreated, makeAuthSuccessResponse(
		&MakeAuthSuccessResponse{
			token:     token,
			message:   "Signup Successful",
			user:      user,
			config:    h.config,
			oauthUser: nil,
		},
	))
}

type MakeAuthSuccessResponse struct {
	token     string
	message   string
	user      *models.User
	config    *conf.Config
	oauthUser *models.UserOauth
}

func makeAuthSuccessResponse(payload *MakeAuthSuccessResponse) map[string]interface{} {
	conf := conf.Get()
	user := payload.user
	avatar := conf.Server.PublicUrl + "/" + user.AvatarURL(conf.App.AvatarURLTemplate)

	if payload.oauthUser != nil {
		avatar = *payload.oauthUser.AvatarUrl
	}
	return map[string]interface{}{
		"message": payload.message,
		"status":  http.StatusCreated,
		"data": map[string]interface{}{
			"token": payload.token,
			"user": map[string]interface{}{
				"id":                       user.ID,
				"email":                    user.Email,
				"has_wakatime_integration": user.WakatimeApiKey != "",
				"avatar":                   avatar,
			},
		},
	}
}

func MakeLoginJWT(userId string, conf *conf.Config) (string, int64, error) {
	ttl := time.Now().Add(JWT_TOKEN_DURATION).Unix()
	atClaims := jwt.MapClaims{}
	atClaims["exp"] = ttl
	atClaims["uid"] = userId
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	token, err := at.SignedString([]byte(conf.Security.JWT_SECRET))
	if err != nil {
		return "", 0, err
	}

	return token, ttl / 1000000, nil // kinda wonder if its bad ide to return ttl in seconds
}

func (h *AuthApiHandler) ValidateAuthToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	if token == "" {
		helpers.RespondJSON(w, r, http.StatusUnauthorized, map[string]interface{}{
			"message": "Unauthorized",
			"status":  http.StatusUnauthorized,
		})
	}

	claim, err := utils.GetTokenClaims(token, h.config.Security.JWT_SECRET)
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

func (h *AuthApiHandler) GetApiKey(w http.ResponseWriter, r *http.Request) {
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
func (h *AuthApiHandler) RefreshApiKey(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	if user == nil {
		helpers.RespondJSON(w, r, http.StatusUnauthorized, map[string]interface{}{
			"message": "Unauthorized",
			"status":  http.StatusUnauthorized,
		})
		return
	}

	user.ApiKey = h.userService.MakeApiKey()
	user, err := h.userService.Update(user)

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

func (h *AuthApiHandler) handlePasswordReset(user *models.User) error {
	updatedUser, err := h.userService.GenerateResetToken(user)
	if err != nil {
		return err
	}

	go h.sendPasswordResetEmail(updatedUser)
	return nil
}

func (h *AuthApiHandler) sendPasswordResetEmail(user *models.User) {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", h.config.Server.GetFrontendUri(), user.ResetToken)

	if err := h.mailService.SendPasswordReset(user, resetLink); err != nil {
		conf.Log().Error("failed to send password reset mail",
			"userID", user.ID,
			"error", err,
		)
		conf.Log().Info("Password reset link", resetLink, "userID")
		return
	}

	slog.Info("sent password reset mail", "userID", user.ID)
}

func (h *AuthApiHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.userService.GetUserByEmail(resetRequest.Email)
	if err != nil || user == nil {
		conf.Log().Request(r).Warn("password reset requested for unregistered address",
			"email", resetRequest.Email,
		)
	} else {
		if err := h.handlePasswordReset(user); err != nil {
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
