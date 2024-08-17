package github

import (
	"context"
	"encoding/json"
	"fmt"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/utils"

	"github.com/pkg/errors"
)

var (
	GithubAuthBaseUrl = "https://github.com"
	GithubApiBaseUrl  = "https://api.github.com"
)

type GithubTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type Plan struct {
	Name          string `json:"name"`
	Space         int    `json:"space"`
	Collaborators int    `json:"collaborators"`
	PrivateRepos  int    `json:"private_repos"`
}

// User represents the user information from the JSON
type GithubUser struct {
	Login                   string  `json:"login"`
	ID                      int     `json:"id"`
	NodeID                  string  `json:"node_id"`
	AvatarURL               string  `json:"avatar_url"`
	GravatarID              string  `json:"gravatar_id"`
	URL                     string  `json:"url"`
	HtmlUrl                 string  `json:"html_url"`
	FollowersURL            string  `json:"followers_url"`
	FollowingURL            string  `json:"following_url"`
	GistsURL                string  `json:"gists_url"`
	StarredURL              string  `json:"starred_url"`
	SubscriptionsURL        string  `json:"subscriptions_url"`
	OrganizationsURL        string  `json:"organizations_url"`
	ReposURL                string  `json:"repos_url"`
	EventsURL               string  `json:"events_url"`
	ReceivedEventsURL       string  `json:"received_events_url"`
	Type                    string  `json:"type"`
	SiteAdmin               bool    `json:"site_admin"`
	Name                    *string `json:"name"`    // Use pointer to handle null
	Company                 *string `json:"company"` // Use pointer to handle null
	Blog                    string  `json:"blog"`
	Location                *string `json:"location"`           // Use pointer to handle null
	Email                   *string `json:"email"`              // Use pointer to handle null
	Hireable                *bool   `json:"hireable"`           // Use pointer to handle null
	Bio                     *string `json:"bio"`                // Use pointer to handle null
	TwitterUsername         *string `json:"twitter_username"`   // Use pointer to handle null
	NotificationEmail       *string `json:"notification_email"` // Use pointer to handle null
	PublicRepos             int     `json:"public_repos"`
	PublicGists             int     `json:"public_gists"`
	Followers               int     `json:"followers"`
	Following               int     `json:"following"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at"`
	PrivateGists            int     `json:"private_gists"`
	TotalPrivateRepos       int     `json:"total_private_repos"`
	OwnedPrivateRepos       int     `json:"owned_private_repos"`
	DiskUsage               int     `json:"disk_usage"`
	Collaborators           int     `json:"collaborators"`
	TwoFactorAuthentication bool    `json:"two_factor_authentication"`
	Plan                    Plan    `json:"plan"`
}

type GithubUserEmail struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

type OauthConfigEndpoint struct {
	AuthURL  string
	TokenURL string
}

type OauthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Endpoint     OauthConfigEndpoint
}

var githubEndpoint = OauthConfigEndpoint{
	AuthURL:  fmt.Sprintf("%s/oauth/authorize", GithubAuthBaseUrl),
	TokenURL: fmt.Sprintf("%s/login/oauth/access_token", GithubAuthBaseUrl),
}

func GetOAuthConfig() (*OauthConfig, error) {
	var (
		FrontendUri        = conf.Get().Security.FrontendUri
		GithubClientSecret = conf.Get().Security.GithubClientSecret
		GithubClientId     = conf.Get().Security.GithubClientId
	)
	if GithubClientId == "" {
		return nil, errors.New("GITHUB_CLIENT_ID not set")
	}
	if GithubClientSecret == "" {
		return nil, errors.New("GITHUB_CLIENT_SECRET not set")
	}
	if FrontendUri == "" {
		return nil, errors.New("FRONTEND_URI not set")
	}

	return &OauthConfig{
		ClientID:     GithubClientId,
		ClientSecret: GithubClientSecret,
		Endpoint:     githubEndpoint,
		RedirectURL:  fmt.Sprintf("%s/api/oauth/callback/github", FrontendUri),
	}, nil
}

type GithubAccessTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
}

func GetGithubUserEmails(ctx context.Context, accessToken string) ([]*GithubUserEmail, error) {
	config := utils.JsonHttpRequestConfig{
		Method:      "GET",
		Url:         fmt.Sprintf("%s/user/emails", GithubApiBaseUrl),
		AccessToken: accessToken,
	}

	response, err := utils.MakeJSONHttpRequest[[]*GithubUserEmail](&config)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func GetPrimaryVerifiedGithubUserEmail(emails []*GithubUserEmail) (*GithubUserEmail, error) {
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email, nil
		}
	}
	return nil, errors.New("no primary verified email found")
}

func GetPrimaryGithubEmail(accessToken string) (*GithubUserEmail, error) {
	// Get the user's email address
	emails, err := GetGithubUserEmails(context.Background(), accessToken)
	if err != nil {
		return nil, err
	}
	if len(emails) == 0 {
		return nil, errors.New("no emails found")
	}
	return GetPrimaryVerifiedGithubUserEmail(emails)
}

func GetGithubUser(accessToken string) (*GithubUser, error) {
	// Get the user's email address
	config := utils.JsonHttpRequestConfig{
		Method:      "GET",
		Url:         fmt.Sprintf("%s/user", GithubApiBaseUrl),
		AccessToken: accessToken,
	}

	response, err := utils.MakeJSONHttpRequest[*GithubUser](&config)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func GetGithubAccessToken(ctx context.Context, code string) (*GithubAccessTokenResponse, error) {
	conf, err := GetOAuthConfig()
	if err != nil {
		return nil, err
	}

	payload := struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
		RedirectURI  string `json:"redirect_uri"`
	}{
		ClientId:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		Code:         code,
		RedirectURI:  conf.RedirectURL,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	tokenUrl := conf.Endpoint.TokenURL

	config := utils.JsonHttpRequestConfig{
		Method: "POST",
		Url:    tokenUrl,
		Body:   string(body),
	}

	response, err := utils.MakeJSONHttpRequest[*GithubAccessTokenResponse](&config)
	if err != nil {
		return nil, err
	}

	return response, nil
}
