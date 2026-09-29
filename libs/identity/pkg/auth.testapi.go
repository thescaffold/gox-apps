package pkg

import (
	identityattribute "github.com/thescaffold/gox-apps/libs/identity/app/attribute"
	identitypermission "github.com/thescaffold/gox-apps/libs/identity/app/permission"
	identityrole "github.com/thescaffold/gox-apps/libs/identity/app/role"
	"github.com/thescaffold/gox-apps/libs/identity/app/token"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/awesome-goose/goose/modules/sql"
	"gorm.io/gorm"
)

// NewAuthServiceForTest constructs an AuthService wired to the given
// gorm.DB, bypassing DI — mirrors gox-apps/libs/systems/app/system.testapi.go's
// NewSystemEntityForTest. AuthService's DI fields are unexported, so this
// constructor has to live in package pkg alongside it; an external test
// package couldn't set them directly.
//
// cache and tracker are left nil, matching GenerateOTP/VerifyOTP's own
// documented nil-tolerance — the methods this constructor exists to test
// (Login, VerifyCredentials, IssueTokens, ResolveRoles, ResolvePermissions,
// ResolvePreference) never touch either field.
func NewAuthServiceForTest(db *gorm.DB) *AuthService {
	q := (&sql.Query{}).With(&sql.Db{DB: db})

	userEntity := &user.UserEntity{Entity: (&sql.Entity[user.User]{}).With(q)}
	userEntity.OnRegister()

	tokenEntity := &token.TokenEntity{Entity: (&sql.Entity[token.Token]{}).With(q)}
	tokenEntity.OnRegister()

	roleEntity := &identityrole.RoleEntity{Entity: (&sql.Entity[identityrole.Role]{}).With(q)}
	roleEntity.OnRegister()

	permissionEntity := &identitypermission.PermissionEntity{Entity: (&sql.Entity[identitypermission.Permission]{}).With(q)}
	permissionEntity.OnRegister()

	attributeEntity := &identityattribute.AttributeEntity{Entity: (&sql.Entity[identityattribute.Attribute]{}).With(q)}
	attributeEntity.OnRegister()

	return &AuthService{
		userEntity:       userEntity,
		tokenEntity:      tokenEntity,
		roleEntity:       roleEntity,
		permissionEntity: permissionEntity,
		attributeEntity:  attributeEntity,
	}
}
