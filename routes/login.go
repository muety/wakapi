package routes

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dchest/captcha"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

type LoginHandler struct {
	config       *conf.Config
	userSrvc     services.IUserService
	mailSrvc     services.IMailService
	keyValueSrvc services.IKeyValueService
}

func NewLoginHandler(userService services.IUserService, mailService services.IMailService, keyValueService services.IKeyValueService) *LoginHandler {
	return &LoginHandler{
		config:       conf.Get(),
		userSrvc:     userService,
		mailSrvc:     mailService,
		keyValueSrvc: keyValueService,
	}
}

func (h *LoginHandler) RegisterRoutes(router chi.Router) {
	router.Get("/login", h.GetIndex)
	router.
		With(httprate.LimitByRealIP(h.config.Security.GetLoginMaxRate())).
		Post("/login", h.PostLogin)
	router.Get("/signup", h.GetSignup)
	router.
		With(httprate.LimitByRealIP(h.config.Security.GetSignupMaxRate())).
		Post("/signup", h.PostSignup)
	router.Get("/set-password", h.GetSetPassword)
	router.Post("/set-password", h.PostSetPassword)
	router.Get("/reset-password", h.GetResetPassword)
	router.
		With(httprate.LimitByRealIP(h.config.Security.GetPasswordResetMaxRate())).
		Post("/reset-password", h.PostResetPassword)

	authMiddleware := middlewares.NewAuthenticateMiddleware(h.userSrvc).
		WithRedirectTarget(defaultErrorRedirectTarget()).
		WithRedirectErrorMessage("unauthorized").
		WithOptionalFor("/logout")

	logoutRouter := chi.NewRouter()
	logoutRouter.Use(authMiddleware.Handler)
	logoutRouter.Post("/", h.PostLogout)
	router.Mount("/logout", logoutRouter)
}

func (h *LoginHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r, w, false))
}

func (h *LoginHandler) PostLogin(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	var login models.Login
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("missing parameters"))
		return
	}
	if err := loginDecoder.Decode(&login, r.PostForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("missing parameters"))
		return
	}

	user, err := h.userSrvc.GetUserById(login.Username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("resource not found"))
		return
	}

	if !utils.ComparePassword(user.Password, login.Password, h.config.Security.PasswordSalt) {
		w.WriteHeader(http.StatusUnauthorized)
		templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("invalid credentials"))
		return
	}

	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, login.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		conf.Log().Request(r).Error("failed to encode secure cookie", "error", err)
		templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("internal server error"))
		return
	}

	user.LastLoggedInAt = models.CustomTime(time.Now())
	h.userSrvc.Update(user)

	http.SetCookie(w, h.config.CreateCookie(models.AuthCookieKey, encoded))
	http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
}

func (h *LoginHandler) PostLogout(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if user := middlewares.GetPrincipal(r); user != nil {
		h.userSrvc.FlushUserCache(user.ID)
	}
	http.SetCookie(w, h.config.GetClearCookie(models.AuthCookieKey))
	http.Redirect(w, r, fmt.Sprintf("%s/", h.config.Server.BasePath), http.StatusFound)
}

func (h *LoginHandler) GetSignup(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r, w, h.config.Security.SignupCaptcha))
}

func (h *LoginHandler) PostSignup(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	var signup models.Signup
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r, w, h.config.Security.SignupCaptcha).WithError("missing parameters"))
		return
	}
	if err := signupDecoder.Decode(&signup, r.PostForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r, w, h.config.Security.SignupCaptcha).WithError("missing parameters"))
		return
	}

	if !h.config.IsDev() && !h.config.Security.AllowSignup && (!h.config.Security.InviteCodes || signup.InviteCode == "") {
		w.WriteHeader(http.StatusForbidden)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r, w, h.config.Security.SignupCaptcha).WithError("registration is disabled on this server"))
		return
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	var invitedBy string
	var invitedDate time.Time
	var inviteCodeKey = fmt.Sprintf("%s_%s", conf.KeyInviteCode, signup.InviteCode)

	if kv, _ := h.keyValueSrvc.GetString(inviteCodeKey); kv != nil && kv.Value != "" {
		if parts := strings.Split(kv.Value, ","); len(parts) == 2 {
			invitedBy = parts[0]
			invitedDate, _ = time.Parse(time.RFC3339, parts[1])
		}

		if err := h.keyValueSrvc.DeleteString(inviteCodeKey); err != nil {
			conf.Log().Error("failed to revoke invite code", "inviteCodeKey", inviteCodeKey, "error", err)
		}
	}

	if signup.InviteCode != "" && time.Since(invitedDate) > 24*time.Hour {
		w.WriteHeader(http.StatusForbidden)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r, w, h.config.Security.SignupCaptcha).WithError("invite code invalid or expired"))
		return
	}

	signup.InvitedBy = invitedBy

	if !signup.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		errMsg := "invalid parameters"
		if !models.ValidateUsername(signup.Username) {
			errMsg = "User name is invalid"
		}
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r, w, h.config.Security.SignupCaptcha).WithError(errMsg))
		return
	}

	numUsers, _ := h.userSrvc.Count()

	_, created, err := h.userSrvc.CreateOrGet(&signup, numUsers == 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		conf.Log().Request(r).Error("failed to create new user", "error", err)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r, w, h.config.Security.SignupCaptcha).WithError("failed to create new user"))
		return
	}
	if !created {
		w.WriteHeader(http.StatusConflict)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r, w, h.config.Security.SignupCaptcha).WithError("user already existing"))
		return
	}

	routeutils.SetSuccess(r, w, "account created successfully")
	http.Redirect(w, r, h.config.Server.BasePath, http.StatusFound)
}

func (h *LoginHandler) GetResetPassword(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
	templates[conf.ResetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false))
}

func (h *LoginHandler) GetSetPassword(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	values, _ := url.ParseQuery(r.URL.RawQuery)
	token := values.Get("token")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		templates[conf.SetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("invalid or missing token"))
		return
	}

	vm := &view.SetPasswordViewModel{
		LoginViewModel: *h.buildViewModel(r, w, false),
		Token:          token,
	}

	templates[conf.SetPasswordTemplate].Execute(w, vm)
}

func (h *LoginHandler) PostSetPassword(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	var setRequest models.SetPasswordRequest
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("missing parameters"))
		return
	}
	if err := signupDecoder.Decode(&setRequest, r.PostForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("missing parameters"))
		return
	}

	user, err := h.userSrvc.GetUserByResetToken(setRequest.Token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		templates[conf.SetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("invalid token"))
		return
	}

	if !setRequest.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("invalid parameters"))
		return
	}

	user.Password = setRequest.Password
	user.ResetToken = ""
	if hash, err := utils.HashPassword(user.Password, h.config.Security.PasswordSalt); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		conf.Log().Request(r).Error("failed to set new password", "error", err)
		templates[conf.SetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("failed to set new password"))
		return
	} else {
		user.Password = hash
	}

	if _, err := h.userSrvc.Update(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		conf.Log().Request(r).Error("failed to save new password", "error", err)
		templates[conf.SetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("failed to save new password"))
		return
	}

	routeutils.SetSuccess(r, w, "password updated successfully")
	http.Redirect(w, r, fmt.Sprintf("%s/login", h.config.Server.BasePath), http.StatusFound)
}

func (h *LoginHandler) PostResetPassword(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if !h.config.Mail.Enabled {
		w.WriteHeader(http.StatusNotImplemented)
		templates[conf.ResetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("mailing is disabled on this server"))
		return
	}

	var resetRequest models.ResetPasswordRequest
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.ResetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("missing parameters"))
		return
	}
	if err := resetPasswordDecoder.Decode(&resetRequest, r.PostForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.ResetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("missing parameters"))
		return
	}

	if user, err := h.userSrvc.GetUserByEmail(resetRequest.Email); user != nil && err == nil {
		if u, err := h.userSrvc.GenerateResetToken(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			conf.Log().Request(r).Error("failed to generate password reset token", "error", err)
			templates[conf.ResetPasswordTemplate].Execute(w, h.buildViewModel(r, w, false).WithError("failed to generate password reset token"))
			return
		} else {
			go func(user *models.User, r *http.Request) {
				link := fmt.Sprintf("%s/set-password?token=%s", h.config.Server.GetPublicUrl(), user.ResetToken)
				if err := h.mailSrvc.SendPasswordReset(user, link); err != nil {
					conf.Log().Request(r).Error("failed to send password reset mail", "userID", user.ID, "error", err)
				} else {
					slog.Info("sent password reset mail", "userID", user.ID)
				}
			}(u, r)
		}
	} else {
		conf.Log().Request(r).Warn("password reset requested for unregistered address", "email", resetRequest.Email)
	}

	routeutils.SetSuccess(r, w, "an e-mail was sent to you in case your e-mail address was registered")
	http.Redirect(w, r, h.config.Server.BasePath, http.StatusFound)
}

func (h *LoginHandler) buildViewModel(r *http.Request, w http.ResponseWriter, withCaptcha bool) *view.LoginViewModel {
	numUsers, _ := h.userSrvc.Count()

	vm := &view.LoginViewModel{
		SharedViewModel: view.NewSharedViewModel(h.config, nil),
		TotalUsers:      int(numUsers),
		AllowSignup:     h.config.IsDev() || h.config.Security.AllowSignup,
		InviteCode:      r.URL.Query().Get("invite"),
	}

	if withCaptcha {
		vm.CaptchaId = captcha.New()
	}

	return routeutils.WithSessionMessages(vm, r, w)
}
