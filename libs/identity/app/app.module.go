package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	identityattribute "github.com/thescaffold/gox-apps/libs/identity/app/attribute"
	identityclient "github.com/thescaffold/gox-apps/libs/identity/app/client"
	identityclientlog "github.com/thescaffold/gox-apps/libs/identity/app/clientlog"
	identitydevice "github.com/thescaffold/gox-apps/libs/identity/app/device"
	identitydevicelog "github.com/thescaffold/gox-apps/libs/identity/app/devicelog"
	identitydevicesession "github.com/thescaffold/gox-apps/libs/identity/app/devicesession"
	identityinvite "github.com/thescaffold/gox-apps/libs/identity/app/invite"
	identitypermission "github.com/thescaffold/gox-apps/libs/identity/app/permission"
	identitypermissiontype "github.com/thescaffold/gox-apps/libs/identity/app/permissiontype"
	identityprovider "github.com/thescaffold/gox-apps/libs/identity/app/provider"
	identityproviderlog "github.com/thescaffold/gox-apps/libs/identity/app/providerlog"
	identityrole "github.com/thescaffold/gox-apps/libs/identity/app/role"
	identityroletype "github.com/thescaffold/gox-apps/libs/identity/app/roletype"
	identitytoken "github.com/thescaffold/gox-apps/libs/identity/app/token"
	identityuser "github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	identityworkspace "github.com/thescaffold/gox-apps/libs/identity/app/workspace"
	"github.com/thescaffold/gox-apps/libs/identity/migrations"
	"github.com/thescaffold/gox-packages/libs/core/module"
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
	// Phase 4 schema alignment — adds TS-side columns alongside existing gox
	// columns (User.ref+type, Permission.role_type_id+permission_type_id,
	// Provider.key+name+desc+..., Device.os+agent+engine+cpu, DeviceLog.ip+
	// country+region+city+area). Idempotent ALTER TABLE ADD COLUMN IF NOT
	// EXISTS — safe to rerun against partially-aligned databases.
	&migrations.AlignIdentityWithNtx{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		// TranslationPaths mirror TS index.ts `paths = [translations/en/ntx/apps/${name}.yaml]`;
		// CoreModule.Boot pre-loads them so translate() resolves real strings.
		module.New(module.CoreConfig{
			TranslationPaths: Paths,
		}),
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
