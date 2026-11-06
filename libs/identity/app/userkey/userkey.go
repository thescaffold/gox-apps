// Package userkey serves a person's own SSH and signing keys (U-S22,
// pkg/userkey) under apps/identity/user-keys:
//
//	GET    /user-keys       the caller's keys (public parts only)
//	POST   /user-keys       add one {title, usage: auth|signing, key}
//	DELETE /user-keys/:id   remove one of the caller's own keys
//
// Every route needs a signed-in person; nobody reads or changes another
// person's keys.
package userkey

import (
	"context"
	"errors"
	"sync"

	"github.com/awesome-goose/goose/modules/router"
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/userkey"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

var ROUTES = router.ForRoutes(
	router.Get("/user-keys", []any{UserKeyController{}, "List"}, auth.RequireUser()...),
	router.Post("/user-keys", []any{UserKeyController{}, "Add"}, auth.RequireUser()...),
	router.Delete("/user-keys/:id", []any{UserKeyController{}, "Remove"}, auth.RequireUser()...),
)

// UserKeyModule mounts the routes.
type UserKeyModule struct{}

func (m *UserKeyModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}
func (m *UserKeyModule) Exports() []any { return []any{&UserKeyService{}} }
func (m *UserKeyModule) Declarations() []any {
	return []any{&UserKeyService{}, &UserKeyController{}}
}

// UserKeyService owns the store for the process.
type UserKeyService struct {
	db   *sql.Db `inject:""`
	once sync.Once
	svc  *userkey.Service
}

// Keys returns the store.
func (s *UserKeyService) Keys() *userkey.Service {
	s.once.Do(func() { s.svc = &userkey.Service{DB: s.db.DB} })
	return s.svc
}

// Use replaces the store (hosts and tests).
func (s *UserKeyService) Use(svc *userkey.Service) { s.once.Do(func() {}); s.svc = svc }

// NewControllerForTest builds a controller over svc without the container.
func NewControllerForTest(svc *userkey.Service) *UserKeyController {
	s := &UserKeyService{}
	s.Use(svc)
	return &UserKeyController{svc: s}
}

// ListReq, AddReq and IDReq are the requests.
type ListReq struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

type AddReq struct {
	Ctx   ntxctx.NTXContext `context:"ntx"`
	Title string            `json:"title"`
	Usage string            `json:"usage" binding:"required"`
	Key   string            `json:"key"   binding:"required"`
}

type IDReq struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
	Id  string            `param:"id"`
}

// UserKeyController answers the routes.
type UserKeyController struct {
	svc *UserKeyService `inject:""`
}

const title = "keys"

func fail(err error) types.Output {
	switch {
	case errors.Is(err, userkey.ErrNotFound):
		return response.NotFound(title, err.Error())
	case errors.Is(err, userkey.ErrTaken), errors.Is(err, userkey.ErrDuplicate):
		return response.ErrorCode(title, err.Error(), 409, "key_in_use")
	case errors.Is(err, userkey.ErrTooMany):
		return response.ErrorCode(title, err.Error(), 409, "key_limit")
	case errors.Is(err, userkey.ErrPrivate):
		return response.ErrorCode(title, err.Error(), 400, "key_private")
	case errors.Is(err, userkey.ErrInvalid), errors.Is(err, userkey.ErrWeak), errors.Is(err, userkey.ErrGPGAuth), errors.Is(err, userkey.ErrGPGUnusable):
		return response.ErrorCode(title, err.Error(), 400, "key_invalid")
	}
	return response.Internal(title, err)
}

func signedIn(ctx ntxctx.NTXContext) bool { return ctx.UserID != "" }

func (c *UserKeyController) List(dto *ListReq) types.Output {
	if !signedIn(dto.Ctx) {
		return response.Unauthorized(title, "sign in first")
	}
	ks, err := c.svc.Keys().List(context.Background(), dto.Ctx.UserID)
	if err != nil {
		return fail(err)
	}
	return response.Success(ks, title, "keys", nil)
}

func (c *UserKeyController) Add(dto *AddReq) types.Output {
	if !signedIn(dto.Ctx) {
		return response.Unauthorized(title, "sign in first")
	}
	k, err := c.svc.Keys().Add(context.Background(), dto.Ctx.UserID, dto.Usage, dto.Title, dto.Key)
	if err != nil {
		return fail(err)
	}
	return response.Created(k, title, "key added")
}

func (c *UserKeyController) Remove(dto *IDReq) types.Output {
	if !signedIn(dto.Ctx) {
		return response.Unauthorized(title, "sign in first")
	}
	if err := c.svc.Keys().Remove(context.Background(), dto.Ctx.UserID, dto.Id); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"removed": true}, title, "key removed", nil)
}
