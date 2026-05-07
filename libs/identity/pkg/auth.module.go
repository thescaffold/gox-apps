package pkg

import (
	"os"

	identityattribute "github.com/thescaffold/gox-apps-identity/app/attribute"
	identitypermission "github.com/thescaffold/gox-apps-identity/app/permission"
	identityrole "github.com/thescaffold/gox-apps-identity/app/role"
	identitytoken "github.com/thescaffold/gox-apps-identity/app/token"
	identityuser "github.com/thescaffold/gox-apps-identity/app/user"
	coreauth "github.com/thescaffold/gox-packages-core/auth"

	"github.com/awesome-goose/goose/modules/router"
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

func jwtSecret() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	return "changeme"
}

// AuthMiddleware is the pre-built goose middleware that validates Bearer JWTs.
var AuthMiddleware = &coreauth.AuthMiddleware{Secret: jwtSecret()}

var authRoutes = router.ForRoutes(
	router.Post("/auth/login",   []any{AuthAppController{}, "Login"}),
	router.Post("/auth/logout",  []any{AuthAppController{}, "Logout"}),
	router.Post("/auth/refresh", []any{AuthAppController{}, "Refresh"}),
	router.Get("/auth/me",       []any{AuthAppController{}, "Me"}, AuthMiddleware),
	router.Get("/oauth/:provider",          []any{OAuthController{}, "Redirect"}),
	router.Get("/oauth/:provider/callback", []any{OAuthController{}, "Callback"}),
)

type AuthModule struct{}

func (m *AuthModule) Imports() []types.Module {
	return []types.Module{
		authRoutes,
		sql.Child(&sql.Config{}),
		&identityuser.UserModule{},
		&identitytoken.TokenModule{},
		&identityrole.RoleModule{},
		&identitypermission.PermissionModule{},
		&identityattribute.AttributeModule{},
	}
}

func (m *AuthModule) Exports() []any {
	return []any{&AuthService{}}
}

func (m *AuthModule) Declarations() []any {
	return []any{&AuthService{}, &AuthAppController{}, &OAuthController{}}
}
