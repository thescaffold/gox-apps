// Package pushtoken serves an app install's push token (PLAN U5,
// pkg/pushtoken) under apps/identity/devices/push-token:
//
//	PUT    /devices/push-token  register or refresh {platform, token, appVersion, locale, deviceId?}
//	DELETE /devices/push-token  deregister {token} (the app does this on sign-out)
//	GET    /devices/push-token  the caller's registrations (tokens not echoed)
//
// Every route needs a signed-in person, taken from the verified token.
// Signing a device out (POST /logout with a device) or signing out
// everywhere also removes the tokens, so a forgotten DELETE cannot leave a
// signed-out phone receiving pushes.
package pushtoken

import (
	"context"
	"errors"
	"sync"

	"github.com/awesome-goose/goose/modules/router"
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/pushtoken"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

var ROUTES = router.ForRoutes(
	router.Put("/devices/push-token", []any{PushTokenController{}, "Register"}, auth.RequireUser()...),
	router.Delete("/devices/push-token", []any{PushTokenController{}, "Deregister"}, auth.RequireUser()...),
	router.Get("/devices/push-token", []any{PushTokenController{}, "List"}, auth.RequireUser()...),
)

// PushTokenModule mounts the routes.
type PushTokenModule struct{}

func (m *PushTokenModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}
func (m *PushTokenModule) Exports() []any { return []any{&PushTokenService{}} }
func (m *PushTokenModule) Declarations() []any {
	return []any{&PushTokenService{}, &PushTokenController{}}
}

// PushTokenService owns the store for the process.
type PushTokenService struct {
	db   *sql.Db `inject:""`
	once sync.Once
	svc  *pushtoken.Service
}

// Tokens returns the store.
func (s *PushTokenService) Tokens() *pushtoken.Service {
	s.once.Do(func() { s.svc = pushtoken.New(s.db.DB) })
	return s.svc
}

// Use replaces the store (hosts and tests).
func (s *PushTokenService) Use(svc *pushtoken.Service) { s.once.Do(func() {}); s.svc = svc }

// NewControllerForTest builds a controller over svc without the container.
func NewControllerForTest(svc *pushtoken.Service) *PushTokenController {
	s := &PushTokenService{}
	s.Use(svc)
	return &PushTokenController{svc: s}
}

// RegisterReq, DeregisterReq and ListReq are the requests.
type RegisterReq struct {
	Ctx        ntxctx.NTXContext `context:"ntx"`
	Platform   string            `json:"platform"`
	Token      string            `json:"token"`
	AppVersion string            `json:"appVersion"`
	Locale     string            `json:"locale"`
	DeviceId   string            `json:"deviceId"`
}

type DeregisterReq struct {
	Ctx   ntxctx.NTXContext `context:"ntx"`
	Token string            `json:"token"`
}

type ListReq struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// PushTokenController answers the routes.
type PushTokenController struct {
	svc *PushTokenService `inject:""`
}

const title = "push"

func fail(err error) types.Output {
	switch {
	case errors.Is(err, pushtoken.ErrNoUser):
		return response.Unauthorized(title, "sign in first")
	case errors.Is(err, pushtoken.ErrPlatform):
		return response.ErrorCode(title, err.Error(), 400, "push_platform")
	case errors.Is(err, pushtoken.ErrToken), errors.Is(err, pushtoken.ErrField):
		return response.ErrorCode(title, err.Error(), 400, "push_token_invalid")
	}
	return response.Internal(title, err)
}

// Register is PUT /devices/push-token. data: the registration
// {id, deviceId?, platform, provider, appVersion, locale, createdAt, updatedAt}.
func (c *PushTokenController) Register(dto *RegisterReq) types.Output {
	t, err := c.svc.Tokens().Register(context.Background(), dto.Ctx.UserID, pushtoken.Registration{
		Platform: dto.Platform, Token: dto.Token, AppVersion: dto.AppVersion, Locale: dto.Locale, DeviceId: dto.DeviceId,
	})
	if err != nil {
		return fail(err)
	}
	return response.Success(t, title, "push token registered", nil)
}

// Deregister is DELETE /devices/push-token. data: {removed}.
func (c *PushTokenController) Deregister(dto *DeregisterReq) types.Output {
	removed, err := c.svc.Tokens().Deregister(context.Background(), dto.Ctx.UserID, dto.Token)
	if err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"removed": removed}, title, "push token removed", nil)
}

// List is GET /devices/push-token. data: [registration].
func (c *PushTokenController) List(dto *ListReq) types.Output {
	if dto.Ctx.UserID == "" {
		return fail(pushtoken.ErrNoUser)
	}
	ts, err := c.svc.Tokens().For(context.Background(), dto.Ctx.UserID)
	if err != nil {
		return fail(err)
	}
	return response.Success(ts, title, "push tokens", nil)
}
