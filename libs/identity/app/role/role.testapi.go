package role

import (
	"github.com/awesome-goose/goose/modules/sql"
	"gorm.io/gorm"
)

// NewControllerForTest builds a RoleController for the checks that need only
// the database.
func NewControllerForTest(db *gorm.DB) *RoleController {
	return &RoleController{db: &sql.Db{DB: db}}
}

// ScopeLinkForTest exposes scopeLink.
func (c *RoleController) ScopeLinkForTest(dto *LinkRoleDto) string { return c.scopeLink(dto) }

// AnyPlatformRoleForTest exposes anyPlatformRole.
func (c *RoleController) AnyPlatformRoleForTest(ids []string) bool { return c.anyPlatformRole(ids) }
