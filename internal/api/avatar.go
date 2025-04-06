package api

import (
	"net/http"
	"time"

	"codeberg.org/Codeberg/avatars"
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/utils"
)

func (a *APIv1) GetAvatarHash(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	if utils.IsNoCache(r, 1*time.Hour) {
		a.lruCache.Remove(hash)
	}

	if !a.lruCache.Contains(hash) {
		a.lruCache.Add(hash, avatars.MakeAvatar(hash))
	}
	data, _ := a.lruCache.Get(hash)

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "max-age=2592000")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data.(string)))
}
