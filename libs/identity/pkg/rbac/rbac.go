// Package rbac decides who may change role and permission data. Roles reach
// signed tokens, so these writes are the way to platform-wide power.
//
// Three levels:
//   - a platform admin (ROOT_USER_ID, or a token holding a platform admin role)
//     may do anything;
//   - a workspace admin (owner or admin of the token's workspace) may manage
//     roles inside that workspace, never platform roles;
//   - everyone else may read.
//
// Role types, permission types and permission links are global tables, so a
// workspace admin who edits them affects every tenant. Consoles built on the
// identity screens still do that, so tightening it is opt-in: set
// IDENTITY_RBAC_STRICT=true to make every RBAC write platform-admin only.
package rbac

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/awesome-goose/goose/types"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// platformRoles are the role names that act across every workspace.
var platformRoles = map[string]bool{"root": true, "global-admin": true, "global-member": true}

// platformAdminRoles are the ones that may administer RBAC.
var platformAdminRoles = map[string]bool{"root": true, "global-admin": true}

// normalize folds the spellings in use: global_admin, Global-Admin.
func normalize(role string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(role)), "_", "-")
}

// Privileged reports whether a role name is a platform role.
func Privileged(role string) bool { return platformRoles[normalize(role)] }

func holds(ntx ntxctx.NTXContext, match func(string) bool) bool {
	for _, r := range ntx.Roles {
		if name, ok := r.(string); ok && match(name) {
			return true
		}
	}
	return false
}

// PlatformAdmin reports whether a verified caller administers the platform.
func PlatformAdmin(ntx ntxctx.NTXContext) bool {
	if ntx.UserID == "" {
		return false
	}
	if root := os.Getenv("ROOT_USER_ID"); root != "" && ntx.UserID == root {
		return true
	}
	return holds(ntx, func(r string) bool { return platformAdminRoles[normalize(r)] })
}

// WorkspaceAdmin reports whether a verified caller owns or administers the
// workspace their token is for.
func WorkspaceAdmin(ntx ntxctx.NTXContext) bool {
	if ntx.UserID == "" || ntx.WorkspaceID == "" {
		return false
	}
	return holds(ntx, func(r string) bool { return r == "owner" || r == "admin" })
}

// Strict reports whether every RBAC write is platform-admin only.
func Strict() bool {
	v := strings.ToLower(os.Getenv("IDENTITY_RBAC_STRICT"))
	return v == "true" || v == "1"
}

// Allowed decides a write by a verified caller. Reads never reach it.
func Allowed(ntx ntxctx.NTXContext) bool {
	if PlatformAdmin(ntx) {
		return true
	}
	return !Strict() && WorkspaceAdmin(ntx)
}

var errForbidden = errors.New("rbac: forbidden")

// Guard refuses RBAC writes the caller may not make. It runs after the session
// and FromClaims middleware, which supply the verified context it reads.
type Guard struct{}

func (g *Guard) Handle(ctx types.Context) error {
	switch strings.ToUpper(string(ctx.Request().Method())) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return nil
	}
	if Allowed(ntxctx.Get(ctx)) {
		return nil
	}
	body, _ := json.Marshal(response.Forbidden("Forbidden", "you may not change roles or permissions").Data())
	_ = ctx.Response().Write(types.SerialTypeObject, body, http.StatusForbidden)
	return errForbidden
}

// Middleware is what an RBAC route needs: a session, its verified context, and
// the guard. Apply it to routes outside a Resource.
func Middleware() []types.Middleware {
	return append(coreauth.Session(), &coreauth.FromClaims{}, &Guard{})
}

// Secure protects a Resource: every route needs a session and verified context,
// and its writes go through the guard.
func Secure(route types.Route) types.Route {
	secured := make(types.Routes, 0, len(route.Children))
	for _, child := range route.Children {
		child.Middlewares = append(append(types.Middlewares{}, Middleware()...), child.Middlewares...)
		secured = append(secured, child)
	}
	route.Children = secured
	return route
}
