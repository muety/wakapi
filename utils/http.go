package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/mileusna/useragent"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	cacheMaxAgePattern = `max-age=(\d+)`
)

var (
	cacheMaxAgeRe *regexp.Regexp
)

func init() {
	cacheMaxAgeRe = regexp.MustCompile(cacheMaxAgePattern)
}

type PageParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (p *PageParams) Limit() int {
	if p.PageSize < 0 {
		return 0
	}
	return p.PageSize
}

func (p *PageParams) Offset() int {
	if p.PageSize <= 0 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

// IsNoCache checks whether returning a cached resource no older than cacheTtl is allowed given the incoming request
func IsNoCache(r *http.Request, cacheTtl time.Duration) bool {
	cacheControl := r.Header.Get("cache-control")
	if strings.Contains(cacheControl, "no-cache") {
		return true
	}
	if match := cacheMaxAgeRe.FindStringSubmatch(cacheControl); match != nil && len(match) > 1 {
		if maxAge, _ := strconv.Atoi(match[1]); time.Duration(maxAge)*time.Second <= cacheTtl {
			return true
		}
	}
	return false
}

func ParsePageParams(r *http.Request) *PageParams {
	pageParams := &PageParams{}
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("page_size")
	if p, err := strconv.Atoi(page); err == nil {
		pageParams.Page = p
	}
	if p, err := strconv.Atoi(pageSize); err == nil && pageParams.Page > 0 {
		pageParams.PageSize = p
	}
	return pageParams
}

func ParsePageParamsWithDefault(r *http.Request, page, size int) *PageParams {
	pageParams := ParsePageParams(r)
	if pageParams.Page == 0 {
		pageParams.Page = page
	}
	if pageParams.PageSize == 0 {
		pageParams.PageSize = size
	}
	return pageParams
}

func ParseUserAgent(ua string) (string, string, error) { // os, editor, err
	// try parse wakatime client user agents
	re := regexp.MustCompile(`(?iU)^(?:(?:wakatime|chrome|firefox|edge)\/(?:v?[\d+.]+|unset)?\s)?(?:\(?(\w+)[-_].*\)?.+\s)?([^\/\s]+)-wakatime\/.+$`)
	if groups := re.FindAllStringSubmatch(ua, -1); len(groups) > 0 && len(groups[0]) == 3 {
		if groups[0][1] == "win" {
			groups[0][1] = "windows"
		}
		return strutil.Capitalize(groups[0][1]), groups[0][2], nil
	}
	// try parse browser user agent as a fallback
	if parsed := useragent.Parse(ua); len(parsed.Name) > 0 && len(parsed.OS) > 0 {
		return strutil.Capitalize(parsed.OS), parsed.Name, nil
	}
	return "", "", errors.New("failed to parse user agent string")
}

func RaiseForStatus(res *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return res, err
	}
	if res.StatusCode >= 400 {
		message := "<body omitted or empty>"
		contentType := res.Header.Get("content-type")
		if strings.HasPrefix(contentType, "text/") || strings.HasPrefix(contentType, "application/json") {
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			res.Body = io.NopCloser(bytes.NewBuffer(body))
			message = string(body)
		}
		return res, fmt.Errorf("got response status %d for '%s %s' - %s", res.StatusCode, res.Request.Method, res.Request.URL.String(), message)
	}
	return res, nil
}
