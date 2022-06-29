package utils

import (
	"errors"
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"net/http"
)

// CheckEffectiveUser extracts the requested user from a URL (like '/users/{user}'), compares it with the currently authorized user and writes an HTTP error if they differ.
// Fallback can be used to manually set a value for '{user}' if none is present.
func CheckEffectiveUser(w http.ResponseWriter, r *http.Request, userService services.IUserService, fallback string) (*models.User, error) {
	var vars = mux.Vars(r)
	var authorizedUser, requestedUser *models.User

	if vars["user"] == "" {
		vars["user"] = fallback
	}

	authorizedUser = middlewares.GetPrincipal(r)
	if authorizedUser != nil {
		if vars["user"] == "current" {
			vars["user"] = authorizedUser.ID
		}
	}

	requestedUser, err := userService.GetUserById(vars["user"])
	if err != nil {
		err := errors.New("user not found")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return nil, err
	}

	if authorizedUser == nil || authorizedUser.ID != requestedUser.ID && !authorizedUser.IsAdmin {
		err := errors.New(conf.ErrUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return nil, err
	}

	return requestedUser, nil
}
