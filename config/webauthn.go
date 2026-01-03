package config

import (
	"encoding/gob"
	"log/slog"
	"net/url"

	"github.com/go-webauthn/webauthn/webauthn"
)

var WebAuthn *webauthn.WebAuthn

func initWebAuthn() {
	gob.Register(&webauthn.SessionData{})

	parsedURL, err := url.Parse(cfg.Server.PublicUrl)
	if err != nil {
		slog.Error("webauthn init error", "error", err)
	}

	webauthnConfig := &webauthn.Config{
		RPDisplayName: "Wakapi",
		RPID:          parsedURL.Hostname(),           // without "https://"
		RPOrigins:     []string{cfg.Server.PublicUrl}, // with "https://"
	}

	WebAuthn, err = webauthn.New(webauthnConfig)
	if err != nil {
		slog.Error("webauthn init error", "error", err)
	}
}
