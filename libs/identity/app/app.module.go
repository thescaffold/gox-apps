package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	identityattribute "github.com/thescaffold/gox-apps-identity/app/attribute"
	identityclient "github.com/thescaffold/gox-apps-identity/app/client"
	identityclientlog "github.com/thescaffold/gox-apps-identity/app/clientlog"
	identitydevice "github.com/thescaffold/gox-apps-identity/app/device"
	identitydevicelog "github.com/thescaffold/gox-apps-identity/app/devicelog"
	identitydevicesession "github.com/thescaffold/gox-apps-identity/app/devicesession"
	identityinvite "github.com/thescaffold/gox-apps-identity/app/invite"
	identitypermission "github.com/thescaffold/gox-apps-identity/app/permission"
	identitypermissiontype "github.com/thescaffold/gox-apps-identity/app/permissiontype"
	identityprovider "github.com/thescaffold/gox-apps-identity/app/provider"
	identityproviderlog "github.com/thescaffold/gox-apps-identity/app/providerlog"
	identityrole "github.com/thescaffold/gox-apps-identity/app/role"
	identityroletype "github.com/thescaffold/gox-apps-identity/app/roletype"
	identitytoken "github.com/thescaffold/gox-apps-identity/app/token"
	identityuser "github.com/thescaffold/gox-apps-identity/app/user"
	identityucw "github.com/thescaffold/gox-apps-identity/app/userclientworkspace"
	identityworkspace "github.com/thescaffold/gox-apps-identity/app/workspace"
	"github.com/thescaffold/gox-apps-identity/migrations"
	"github.com/thescaffold/gox-packages-core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateIdentityUsers{},
	&migrations.CreateIdentityClients{},
	&migrations.CreateIdentityClientLogs{},
	&migrations.CreateIdentityWorkspaces{},
	&migrations.CreateIdentityUserClientWorkspaces{},
	&migrations.CreateIdentityRoleTypes{},
	&migrations.CreateIdentityRoles{},
	&migrations.CreateIdentityPermissionTypes{},
	&migrations.CreateIdentityPermissions{},
	&migrations.CreateIdentityTokens{},
	&migrations.CreateIdentityDevices{},
	&migrations.CreateIdentityDeviceLogs{},
	&migrations.CreateIdentityDeviceSessions{},
	&migrations.CreateIdentityAttributes{},
	&migrations.CreateIdentityInvites{},
	&migrations.CreateIdentityProviders{},
	&migrations.CreateIdentityProviderLogs{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{Migrations: Migrations}),
		&identityuser.UserModule{},
		&identityclient.ClientModule{},
		&identityclientlog.ClientLogModule{},
		&identityworkspace.WorkspaceModule{},
		&identityucw.UserClientWorkspaceModule{},
		&identityroletype.RoleTypeModule{},
		&identityrole.RoleModule{},
		&identitypermissiontype.PermissionTypeModule{},
		&identitypermission.PermissionModule{},
		&identitytoken.TokenModule{},
		&identitydevice.DeviceModule{},
		&identitydevicelog.DeviceLogModule{},
		&identitydevicesession.DeviceSessionModule{},
		&identityattribute.AttributeModule{},
		&identityinvite.InviteModule{},
		&identityprovider.ProviderModule{},
		&identityproviderlog.ProviderLogModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	svc := &AppService{}
	identityAppSvc = svc
	return []any{&AppController{}, svc}
}
