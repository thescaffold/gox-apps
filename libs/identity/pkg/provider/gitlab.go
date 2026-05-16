package provider

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// GitLab mirrors ntx-apps/libs/identity/src/api/provider/providers/gitlab.service.ts.
type GitLab struct {
	clientID     string
	clientSecret string
	host         string
}

// NewGitLab reads GITLAB_CLIENT_ID / GITLAB_CLIENT_SECRET / GITLAB_HOST from env.
func NewGitLab() *GitLab {
	host := os.Getenv("GITLAB_HOST")
	if host == "" {
		host = "https://gitlab.com"
	}
	return &GitLab{
		clientID:     os.Getenv("GITLAB_CLIENT_ID"),
		clientSecret: os.Getenv("GITLAB_CLIENT_SECRET"),
		host:         host,
	}
}

func (p *GitLab) Name() string { return "gitlab" }

func (p *GitLab) Authorize(state, redirectURI string) string {
	return buildAuthorizeURL(
		p.host+"/oauth/authorize",
		p.clientID, redirectURI, "read_user email", state, nil,
	)
}

func (p *GitLab) Exchange(code, redirectURI string) (*Token, error) {
	body := url.Values{}
	body.Set("code", code)
	body.Set("client_id", p.clientID)
	body.Set("client_secret", p.clientSecret)
	body.Set("redirect_uri", redirectURI)
	body.Set("grant_type", "authorization_code")
	resp, err := http.Post(p.host+"/oauth/token",
		"application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("gitlab: token exchange failed")
	}
	raw, _ := io.ReadAll(resp.Body)
	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken: out.AccessToken, RefreshToken: out.RefreshToken,
		ExpiresIn: out.ExpiresIn, Scope: out.Scope, TokenType: out.TokenType,
	}, nil
}

func (p *GitLab) Profile(token string) (*Profile, error) {
	req, _ := http.NewRequest("GET", p.host+"/api/v4/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("gitlab: profile fetch failed")
	}
	raw, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	id := ""
	if v, ok := data["id"].(float64); ok {
		id = strconv.Itoa(int(v))
	}
	email, _ := data["email"].(string)
	name, _ := data["name"].(string)
	username, _ := data["username"].(string)
	avatar, _ := data["avatar_url"].(string)
	return &Profile{
		ID: id, Email: email, Name: name, Username: username, AvatarURL: avatar,
		Provider: p.Name(), RawPayload: data,
	}, nil
}
