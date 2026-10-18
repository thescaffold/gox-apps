package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	identitypermission "github.com/thescaffold/gox-apps/libs/identity/app/permission"
	identityroletype "github.com/thescaffold/gox-apps/libs/identity/app/roletype"
	"gorm.io/gorm"
)

// NewAppServiceForTest builds an AppService over db without DI, wired only
// with the entities the RBAC registration needs (the other handlers are not
// exercised by it).
func NewAppServiceForTest(db *gorm.DB) *AppService {
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	rt := &identityroletype.RoleTypeEntity{Entity: (&sql.Entity[identityroletype.RoleType]{}).With(q)}
	rt.OnRegister()
	pe := &identitypermission.PermissionEntity{Entity: (&sql.Entity[identitypermission.Permission]{}).With(q)}
	pe.OnRegister()
	return &AppService{roleTypeEntity: rt, permissionEntity: pe}
}
