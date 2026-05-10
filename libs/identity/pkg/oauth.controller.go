package pkg

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

type OAuthCallbackDto struct {
	Code  string `query:"code"`
	State string `query:"state"`
}

type OAuthController struct {
	authService *AuthService `inject:""`
}

// Redirect sends the user to the OAuth provider's authorization URL.
func (c *OAuthController) Redirect(ctx types.Context) types.Output {
	provider := ctx.Request().Params()["provider"]
	return response.Success(map[string]any{
		"provider": provider,
		"message":  "redirect to provider auth URL",
	}, "identity", "oauth redirect", nil)
}

// Callback handles the provider redirect with an auth code.
func (c *OAuthController) Callback(dto *OAuthCallbackDto, ctx types.Context) types.Output {
	provider := ctx.Request().Params()["provider"]
	return response.Success(map[string]any{
		"provider": provider,
		"code":     dto.Code,
		"message":  "exchange code for token",
	}, "identity", "oauth callback", nil)
}
