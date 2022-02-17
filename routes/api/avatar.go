package api

import (
	"codeberg.org/Codeberg/avatars"
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"net/http"
)

type AvatarHandler struct {
	config *conf.Config
}

func NewAvatarHandler() *AvatarHandler {
	return &AvatarHandler{config: conf.Get()}
}

func (h *AvatarHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/avatar/{hash}.svg").Subrouter()
	r.Path("").Methods(http.MethodGet).HandlerFunc(h.Get)
}

func (h *AvatarHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars["hash"] == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte{})
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(avatars.MakeMaleAvatar(vars["hash"])))
}
