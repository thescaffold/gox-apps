package provider

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Bitbucket mirrors ntx-apps/libs/identity/src/api/provider/providers/bitbucket.service.ts.
type Bitbucket struct {
	clientID     string
	clientSecret string
}

// NewBitbucket reads BITBUCKET_CLIENT_ID / BITBUCKET_CLIENT_SECRET from env.
func NewBitbucket() *Bitbucket {
	return &Bitbucket{
		clientID:     os.Getenv("BITBUCKET_CLIENT_ID"),
		clientSecret: os.Getenv("BITBUCKET_CLIENT_SECRET"),
	}
}

func (p *Bitbucket) Name() string { return "bitbucket" }

func (p *Bitbucket) Authorize(state, redirectURI string) string {
	return buildAuthorizeURL(
		"https://bitbucket.org/site/oauth2/authorize",
		p.clientID, redirectURI, "account email", state, nil,
	)
}

func (p *Bitbucket) Exchange(code, redirectURI string) (*Token, error) {
	body := url.Values{}
	body.Set("code", code)
	body.Set("grant_type", "authorization_code")
	body.Set("redirect_uri", redirectURI)
	req, err := http.NewRequest("POST",
		"https://bitbucket.org/site/oauth2/access_token",
		strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	basic := base64.StdEncoding.EncodeToString([]byte(p.clientID + ":" + p.clientSecret))
	req.Header.Set("Authorization", "Basic "+basic)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("bitbucket: token exchange failed")
	}
	raw, _ := io.ReadAll(resp.Body)
	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scopes"`
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

func (p *Bitbucket) Profile(token string) (*Profile, error) {
	req, _ := http.NewRequest("GET", "https://api.bitbucket.org/2.0/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("bitbucket: profile fetch failed")
	}
	raw, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	id, _ := data["account_id"].(string)
	name, _ := data["display_name"].(string)
	username, _ := data["username"].(string)
	avatar := ""
	if links, ok := data["links"].(map[string]any); ok {
		if av, ok := links["avatar"].(map[string]any); ok {
			avatar, _ = av["href"].(string)
		}
	}
	return &Profile{
		ID: id, Name: name, Username: username, AvatarURL: avatar,
		Provider: p.Name(), RawPayload: data,
	}, nil
}
