package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	nUrl "net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	nerrors "github.com/pkg/errors"

	"github.com/duke-git/lancet/v2/strutil"
	"github.com/mileusna/useragent"
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
	re := regexp.MustCompile(`(?iU)^(?:(?:wakatime|chrome|firefox|edge)\/(?:v?[\d+.]+|unset)?\s)?(?:\(?(\w+)[-_].*\)?.+\s)?(?:([^\/\s]+)\/\w+\s)?([^\/\s]+)-wakatime\/.+$`)

	var (
		os, editor string
	)

	if groups := re.FindAllStringSubmatch(ua, -1); len(groups) > 0 && len(groups[0]) == 4 {
		// extract os
		os = groups[0][1]
		if os == "win" {
			os = "windows"
		}
		if os == "darwin" {
			os = "macos"
		}

		// parse editor
		if groups[0][2] == "" {
			editor = groups[0][3] // for most user agents
		} else {
			editor = groups[0][2] // for user agents sent by desktop-wakatime plugin, see https://github.com/muety/wakapi/issues/686
		}
		// special treatment for neovim
		if groups[0][2] == "vim" && strings.Contains(ua, "neovim/") {
			groups[0][2] = "neovim"
		}

		return strutil.Capitalize(os), editor, nil
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

type JsonHttpRequestConfig struct {
	Method      string
	Url         string
	BaseUrl     string
	Body        string
	AccessToken string
}

func MakeJSONHttpRequest[T any](config *JsonHttpRequestConfig) (T, error) {
	var parsedResponse T
	client := &http.Client{}

	var finalUrl = fmt.Sprintf("%s%s", config.BaseUrl, config.Url)
	parsedUrl, err := nUrl.Parse(config.Url)

	if err != nil {
		return parsedResponse, nerrors.Wrap(err, "error parsing url") // this really should not happen
	}

	if parsedUrl.IsAbs() {
		finalUrl = config.Url
	}

	req, err := http.NewRequest(config.Method, finalUrl, strings.NewReader(config.Body))
	if err != nil {
		return parsedResponse, nerrors.Wrap(err, "error making json api request")
	}

	if config.AccessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.AccessToken))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return parsedResponse, nerrors.Wrap(err, fmt.Sprintf("error getting response from url %s", finalUrl))
	}

	b, err := io.ReadAll(res.Body)
	if res.StatusCode != 200 && res.StatusCode != 201 {
		return parsedResponse, nerrors.New("Response with error; status_code=" + res.Status + "; body=" + string(b))
	}

	if err != nil {
		return parsedResponse, nerrors.Wrap(err, "error reading response body")
	}

	err = json.Unmarshal(b, &parsedResponse)
	if err != nil {
		return parsedResponse, nerrors.Wrap(err, "error unmarshalling api response"+string(b))
	}

	return parsedResponse, nil
}

func MakeQueryParams(params map[string]string) string {
	queryParams := nUrl.Values{}
	for key, value := range params {
		queryParams.Set(key, value)
	}
	return queryParams.Encode()
}

func GenerateRandomPassword(length int) (string, error) {
	// Create a slice to hold the random bytes
	bytes := make([]byte, length)

	// Generate random bytes
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Encode bytes to a hexadecimal string
	password := hex.EncodeToString(bytes)

	return password, nil
}
