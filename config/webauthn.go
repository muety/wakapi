package config

import (
	"encoding/gob"
	"log/slog"
	"net/url"

	"github.com/go-webauthn/webauthn/webauthn"
)

var WebAuthn *webauthn.WebAuthn

func initWebAuthn(config *Config) {
	gob.Register(&webauthn.SessionData{})

	parsedURL, err := url.Parse(config.Server.PublicUrl)
	if err != nil {
		slog.Error("webauthn init error", "error", err)
	}

	webauthnConfig := &webauthn.Config{
		RPDisplayName: "Wakapi",
		RPID:          parsedURL.Hostname(),              // without "https://"
		RPOrigins:     []string{config.Server.PublicUrl}, // with "https://"
	}

	WebAuthn, err = webauthn.New(webauthnConfig)
	if err != nil {
		Log().Fatal("webauthn init error", "error", err)
	}
}
