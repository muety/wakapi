package helpers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/muety/wakapi/config"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

var JWT_TOKEN_DURATION = time.Hour * 24

func ExtractCookieAuth(r *http.Request, config *config.Config) (username *string, err error) {
	cookie, err := r.Cookie(models.AuthCookieKey)
	if err != nil {
		return nil, errors.New("missing authentication")
	}

	if err := config.Security.SecureCookie.Decode(models.AuthCookieKey, cookie.Value, &username); err != nil {
		return nil, errors.New("cookie is invalid")
	}

	return username, nil
}

func RespondJSON(w http.ResponseWriter, r *http.Request, status int, object interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(object); err != nil {
		config.Log().Request(r).Error("error while writing json response", "error", err)
	}
}

func ExtractUser(r *http.Request) *models.User {
	type principalGetter interface {
		GetPrincipal() *models.User
	}
	if p := r.Context().Value("principal"); p != nil {
		return p.(principalGetter).GetPrincipal()
	}
	return nil
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

type AuthSuccessResponse struct {
	Message   string
	User      *models.User
	OauthUser *models.UserOauth
}

func MakeAuthSuccessResponse(payload *AuthSuccessResponse) (map[string]interface{}, error) {
	conf := conf.Get()
	user := payload.User
	avatar := conf.Server.PublicUrl + "/" + user.AvatarURL(conf.App.AvatarURLTemplate)

	if payload.OauthUser != nil {
		avatar = *payload.OauthUser.AvatarUrl
	}

	token, _, err := MakeLoginJWT(user.ID, conf)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": payload.Message,
		"status":  http.StatusCreated,
		"data": map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":                       user.ID,
				"email":                    user.Email,
				"has_wakatime_integration": user.WakatimeApiKey != "",
				"avatar":                   avatar,
			},
		},
	}, nil
}
