package auth

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
	identityattribute "github.com/thescaffold/gox-apps/libs/identity/app/attribute"
	identityclient "github.com/thescaffold/gox-apps/libs/identity/app/client"
	identityclientlog "github.com/thescaffold/gox-apps/libs/identity/app/clientlog"
	identitydevice "github.com/thescaffold/gox-apps/libs/identity/app/device"
	identitydevicelog "github.com/thescaffold/gox-apps/libs/identity/app/devicelog"
	identitydevicesession "github.com/thescaffold/gox-apps/libs/identity/app/devicesession"
	identitypermission "github.com/thescaffold/gox-apps/libs/identity/app/permission"
	identitypermissiontype "github.com/thescaffold/gox-apps/libs/identity/app/permissiontype"
	identityroletype "github.com/thescaffold/gox-apps/libs/identity/app/roletype"
	identityuser "github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	identityworkspace "github.com/thescaffold/gox-apps/libs/identity/app/workspace"
	identitypkg "github.com/thescaffold/gox-apps/libs/identity/pkg"
	"github.com/thescaffold/gox-packages/libs/core/services"
	"github.com/thescaffold/gox-packages/libs/core/services/otp"
	"gorm.io/gorm"
	"time"
)

// NewAuthControllerForTest builds an AuthController over db without the DI
// container, wired for the handlers that need only the database (Login,
// Sessions, Logout, UserClients, RegisterClientLog) and, with its in-memory
// cache, the register steps that follow an OTP.
func NewAuthControllerForTest(db *gorm.DB) *AuthController {
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	c := &AuthController{auth: identitypkg.NewAuthServiceForTest(db)}

	c.userEntity = &identityuser.UserEntity{Entity: (&sql.Entity[identityuser.User]{}).With(q)}
	c.attributeEntity = &identityattribute.AttributeEntity{Entity: (&sql.Entity[identityattribute.Attribute]{}).With(q)}
	c.workspaceEntity = &identityworkspace.WorkspaceEntity{Entity: (&sql.Entity[identityworkspace.Workspace]{}).With(q)}
	c.ucwEntity = &identityucw.UserClientWorkspaceEntity{Entity: (&sql.Entity[identityucw.UserClientWorkspace]{}).With(q)}
	c.deviceEntity = &identitydevice.DeviceEntity{Entity: (&sql.Entity[identitydevice.Device]{}).With(q)}
	c.deviceSessionEntity = &identitydevicesession.DeviceSessionEntity{Entity: (&sql.Entity[identitydevicesession.DeviceSession]{}).With(q)}
	c.deviceLogEntity = &identitydevicelog.DeviceLogEntity{Entity: (&sql.Entity[identitydevicelog.DeviceLog]{}).With(q)}
	c.clientEntity = &identityclient.ClientEntity{Entity: (&sql.Entity[identityclient.Client]{}).With(q)}
	c.clientLogEntity = &identityclientlog.ClientLogEntity{Entity: (&sql.Entity[identityclientlog.ClientLog]{}).With(q)}

	c.roleTypeEntity = &identityroletype.RoleTypeEntity{Entity: (&sql.Entity[identityroletype.RoleType]{}).With(q)}
	c.permissionTypeEntity = &identitypermissiontype.PermissionTypeEntity{Entity: (&sql.Entity[identitypermissiontype.PermissionType]{}).With(q)}
	c.permissionEntity = &identitypermission.PermissionEntity{Entity: (&sql.Entity[identitypermission.Permission]{}).With(q)}
	c.cache = services.NewMemoryBackend()
	c.otpSvc = otp.New(c.cache, 7*time.Minute)
	c.roleTypeEntity.OnRegister()
	c.permissionTypeEntity.OnRegister()
	c.permissionEntity.OnRegister()

	c.userEntity.OnRegister()
	c.attributeEntity.OnRegister()
	c.workspaceEntity.OnRegister()
	c.ucwEntity.OnRegister()
	c.deviceEntity.OnRegister()
	c.deviceSessionEntity.OnRegister()
	c.deviceLogEntity.OnRegister()
	c.clientEntity.OnRegister()
	c.clientLogEntity.OnRegister()
	return c
}

// SeedRegistrationForTest leaves what Initiate and Verify leave behind (the OTP
// and the pending-registration record) and returns the token Secret takes, so a
// test can take the last step of sign-up without a delivered code.
func (c *AuthController) SeedRegistrationForTest(id, ref, clientID string) string {
	o := c.otpSvc.Initiate("email", id, ref, "register")
	entry, _ := json.Marshal(registerInitiateEntry{Id: id, Ref: ref, Type: "email", ClientId: clientID})
	c.cache.Set(registerVerifyPrefix+id, string(entry), tempEntryTTL)
	return o.Hashed
}
