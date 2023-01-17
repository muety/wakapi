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
	respondError := func(code int, text string) (*models.User, error) {
		err := errors.New(conf.ErrUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return nil, err
	}

	var vars = mux.Vars(r)

	if vars["user"] == "" {
		vars["user"] = fallback
	}

	authorizedUser := middlewares.GetPrincipal(r)
	if authorizedUser == nil {
		return respondError(http.StatusUnauthorized, conf.ErrUnauthorized)
	} else if vars["user"] == "current" {
		return authorizedUser, nil
	}

	if authorizedUser.ID != vars["user"] && !authorizedUser.IsAdmin {
		return respondError(http.StatusUnauthorized, conf.ErrUnauthorized)
	}

	requestedUser, err := userService.GetUserById(vars["user"])
	if err != nil {
		return respondError(http.StatusNotFound, "user not found")
	}

	return requestedUser, nil
}
