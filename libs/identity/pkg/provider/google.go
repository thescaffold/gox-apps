package provider

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Google mirrors ntx-apps/libs/identity/src/api/provider/providers/google.service.ts.
type Google struct {
	clientID     string
	clientSecret string
}

// NewGoogle reads GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET from env. Missing
// credentials are tolerated — the returned provider's Authorize will produce
// a malformed URL but Exchange/Profile will fail predictably.
func NewGoogle() *Google {
	return &Google{
		clientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		clientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	}
}

func (p *Google) Name() string { return "google" }

func (p *Google) Authorize(state, redirectURI string) string {
	return buildAuthorizeURL(
		"https://accounts.google.com/o/oauth2/v2/auth",
		p.clientID, redirectURI,
		"openid email profile", state, nil,
	)
}

func (p *Google) Exchange(code, redirectURI string) (*Token, error) {
	body := url.Values{}
	body.Set("code", code)
	body.Set("client_id", p.clientID)
	body.Set("client_secret", p.clientSecret)
	body.Set("redirect_uri", redirectURI)
	body.Set("grant_type", "authorization_code")
	resp, err := http.Post("https://oauth2.googleapis.com/token",
		"application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("google: token exchange failed")
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

func (p *Google) Profile(token string) (*Profile, error) {
	req, err := http.NewRequest("GET", "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("google: profile fetch failed")
	}
	raw, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	id, _ := data["sub"].(string)
	email, _ := data["email"].(string)
	name, _ := data["name"].(string)
	picture, _ := data["picture"].(string)
	return &Profile{
		ID: id, Email: email, Name: name, AvatarURL: picture,
		Provider: p.Name(), RawPayload: data,
	}, nil
}
