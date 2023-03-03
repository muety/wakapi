package utils

import (
	"errors"
	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"net/http"
)

// CheckEffectiveUser extracts the requested user from a URL (like '/users/{user}'), compares it with the currently authorized user and writes an HTTP error if they differ.
// Fallback can be used to manually set a value for '{user}' if none is present.
func CheckEffectiveUser(w http.ResponseWriter, r *http.Request, userService services.IUserService, fallback string) (*models.User, error) {
	respondError := func(code int, text string) (*models.User, error) {
		err := errors.New(conf.ErrUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return nil, err
	}

	userParam := chi.URLParam(r, "user")
	if userParam == "" {
		userParam = fallback
	}

	authorizedUser := middlewares.GetPrincipal(r)
	if authorizedUser == nil {
		return respondError(http.StatusUnauthorized, conf.ErrUnauthorized)
	} else if userParam == "current" {
		return authorizedUser, nil
	}

	if authorizedUser.ID != userParam && !authorizedUser.IsAdmin {
		return respondError(http.StatusUnauthorized, conf.ErrUnauthorized)
	}

	requestedUser, err := userService.GetUserById(userParam)
	if err != nil {
		return respondError(http.StatusNotFound, "user not found")
	}

	return requestedUser, nil
}
