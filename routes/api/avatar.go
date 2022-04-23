package api

import (
	"codeberg.org/Codeberg/avatars"
	"github.com/gorilla/mux"
	lru "github.com/hashicorp/golang-lru"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/utils"
	"net/http"
	"time"
)

type AvatarHandler struct {
	config *conf.Config
	cache  *lru.Cache
}

func NewAvatarHandler() *AvatarHandler {
	cache, err := lru.New(1 * 1000 * 64) // assuming an avatar is 1 kb, allocate up to 64 mb of memory for avatars cache
	if err != nil {
		panic(err)
	}

	return &AvatarHandler{
		config: conf.Get(),
		cache:  cache,
	}
}

func (h *AvatarHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/avatar/{hash}.svg").Subrouter()
	r.Path("").Methods(http.MethodGet).HandlerFunc(h.Get)
}

func (h *AvatarHandler) Get(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]

	if utils.IsNoCache(r, 1*time.Hour) {
		h.cache.Remove(hash)
	}

	if !h.cache.Contains(hash) {
		h.cache.Add(hash, avatars.MakeMaleAvatar(hash))
	}
	data, _ := h.cache.Get(hash)

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "max-age=2592000")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data.(string)))
}
