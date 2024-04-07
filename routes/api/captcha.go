package api

import (
	"github.com/dchest/captcha"
	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
)

type CaptchaHandler struct {
	config *conf.Config
}

func NewCaptchaHandler() *CaptchaHandler {
	return &CaptchaHandler{
		config: conf.Get(),
	}
}

func (h *CaptchaHandler) RegisterRoutes(router chi.Router) {
	router.Get("/captcha/{id}.png", captcha.Server(captcha.StdWidth, captcha.StdHeight).ServeHTTP)
}
