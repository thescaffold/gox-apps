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

// GitHub mirrors ntx-apps/libs/identity/src/api/provider/providers/github.service.ts.
type GitHub struct {
	clientID     string
	clientSecret string
}

// NewGitHub reads GITHUB_CLIENT_ID / GITHUB_CLIENT_SECRET from env.
func NewGitHub() *GitHub {
	return &GitHub{
		clientID:     os.Getenv("GITHUB_CLIENT_ID"),
		clientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	}
}

func (p *GitHub) Name() string { return "github" }

func (p *GitHub) Authorize(state, redirectURI string) string {
	return buildAuthorizeURL(
		"https://github.com/login/oauth/authorize",
		p.clientID, redirectURI, "read:user user:email", state, nil,
	)
}

func (p *GitHub) Exchange(code, redirectURI string) (*Token, error) {
	body := url.Values{}
	body.Set("code", code)
	body.Set("client_id", p.clientID)
	body.Set("client_secret", p.clientSecret)
	body.Set("redirect_uri", redirectURI)
	req, err := http.NewRequest("POST",
		"https://github.com/login/oauth/access_token",
		strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("github: token exchange failed")
	}
	raw, _ := io.ReadAll(resp.Body)
	var out struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &Token{
		AccessToken: out.AccessToken, Scope: out.Scope, TokenType: out.TokenType,
	}, nil
}

func (p *GitHub) Profile(token string) (*Profile, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("github: profile fetch failed")
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
	username, _ := data["login"].(string)
	avatar, _ := data["avatar_url"].(string)
	return &Profile{
		ID: id, Email: email, Name: name, Username: username, AvatarURL: avatar,
		Provider: p.Name(), RawPayload: data,
	}, nil
}
