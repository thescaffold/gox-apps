package pkg

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages-core/response"
)

type LoginDto struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
	ClientId string `json:"clientId" binding:"required"`
}

type RefreshDto struct {
	Token string `json:"token" binding:"required"`
}

type LogoutDto struct {
	TokenId string `json:"tokenId" binding:"required"`
}

type AuthAppController struct {
	authService *AuthService `inject:""`
}

func (c *AuthAppController) Login(dto *LoginDto) types.Output {
	access, refresh, err := c.authService.Login(dto.Email, dto.Password, dto.ClientId)
	if err != nil {
		return response.Unauthorized("identity", err.Error())
	}
	return response.Success(map[string]any{
		"accessToken":  access,
		"refreshToken": refresh,
	}, "identity", "login successful", nil)
}

func (c *AuthAppController) Refresh(dto *RefreshDto) types.Output {
	access, err := c.authService.RefreshToken(dto.Token)
	if err != nil {
		return response.Unauthorized("identity", err.Error())
	}
	return response.Success(map[string]any{"accessToken": access}, "identity", "token refreshed", nil)
}

func (c *AuthAppController) Logout(dto *LogoutDto) types.Output {
	if err := c.authService.Logout(dto.TokenId); err != nil {
		return response.Error("identity", "logout failed", 500)
	}
	return response.Success(nil, "identity", "logged out", nil)
}

func (c *AuthAppController) Me(ctx types.Context) types.Output {
	claims := map[string]any{}
	if v := ctx.GetValue("auth"); v != nil {
		if m, ok := v.(map[string]any); ok {
			claims = m
		}
	}
	return response.Success(claims, "identity", "current user", nil)
}
