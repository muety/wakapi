package utils

import (
	"fmt"
	"net"
	"net/url"

	conf "github.com/muety/wakapi/config"
)

func ValidateWakatimeUrl(baseUrl string) error {
	cfg := conf.Get()

	if baseUrl == "" {
		baseUrl = conf.WakatimeApiUrl
	}

	// try to actually parse the url
	baseUrlParsed, err := url.Parse(baseUrl)
	if err != nil {
		return fmt.Errorf("failed to parse wakatime url (%v) – %v", baseUrl, err)
	}

	if !cfg.IsDev() && baseUrlParsed.Scheme != "https" {
		return fmt.Errorf("https is required for wakatime url (%v) – %v", baseUrl, err)
	}

	if baseUrlParsed.Host == cfg.Server.PublicNetUrl.Host {
		return fmt.Errorf("cannot use reference to own instance as wakatime url (%v) – %v", baseUrl, err)
	}

	// resolve ip and validate it's not internal
	ips, err := net.LookupIP(baseUrlParsed.Hostname())
	if err != nil {
		return fmt.Errorf("failed to resolve ip for wakatime url (%v) – %v", baseUrl, err)
	}

	if !cfg.IsDev() {
		for _, ip := range ips {
			if ip.IsPrivate() || ip.IsLoopback() || ip.IsMulticast() || ip.IsUnspecified() {
				return fmt.Errorf("cannot use private ip as wakatime url (%v) (ip: %v)", baseUrl, ip.String())
			}
			if ip.String() == baseUrlParsed.Hostname() {
				return fmt.Errorf("cannot use raw ip as wakatime url (%v) – %v", baseUrl, err)
			}
		}
	}

	return nil
}
