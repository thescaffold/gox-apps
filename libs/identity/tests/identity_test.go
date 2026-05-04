package tests

import (
	"testing"
	"time"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps-identity/app"
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
	"github.com/thescaffold/gox-apps-identity/pkg"
	"github.com/thescaffold/gox-packages-core/auth"
)

func TestUserEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &UserEntitySuite{}).Run()
}
type UserEntitySuite struct{ goosetest.Suite }
func (s *UserEntitySuite) TestTableName() {
	s.T.Expect(identityuser.User{}.TableName()).ToEqual("IdentityUsers")
}

func TestClientEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &ClientEntitySuite{}).Run()
}
type ClientEntitySuite struct{ goosetest.Suite }
func (s *ClientEntitySuite) TestTableName() {
	s.T.Expect(identityclient.Client{}.TableName()).ToEqual("IdentityClients")
}

func TestClientLogEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &ClientLogEntitySuite{}).Run()
}
type ClientLogEntitySuite struct{ goosetest.Suite }
func (s *ClientLogEntitySuite) TestTableName() {
	s.T.Expect(identityclientlog.ClientLog{}.TableName()).ToEqual("IdentityClientLogs")
}

func TestWorkspaceEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &WorkspaceEntitySuite{}).Run()
}
type WorkspaceEntitySuite struct{ goosetest.Suite }
func (s *WorkspaceEntitySuite) TestTableName() {
	s.T.Expect(identityworkspace.Workspace{}.TableName()).ToEqual("IdentityWorkspaces")
}

func TestUserClientWorkspaceEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &UserClientWorkspaceEntitySuite{}).Run()
}
type UserClientWorkspaceEntitySuite struct{ goosetest.Suite }
func (s *UserClientWorkspaceEntitySuite) TestTableName() {
	s.T.Expect(identityucw.UserClientWorkspace{}.TableName()).ToEqual("IdentityUserClientWorkspaces")
}

func TestRoleTypeEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &RoleTypeEntitySuite{}).Run()
}
type RoleTypeEntitySuite struct{ goosetest.Suite }
func (s *RoleTypeEntitySuite) TestTableName() {
	s.T.Expect(identityroletype.RoleType{}.TableName()).ToEqual("IdentityRoleTypes")
}

func TestRoleEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &RoleEntitySuite{}).Run()
}
type RoleEntitySuite struct{ goosetest.Suite }
func (s *RoleEntitySuite) TestTableName() {
	s.T.Expect(identityrole.Role{}.TableName()).ToEqual("IdentityRoles")
}

func TestPermissionTypeEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &PermissionTypeEntitySuite{}).Run()
}
type PermissionTypeEntitySuite struct{ goosetest.Suite }
func (s *PermissionTypeEntitySuite) TestTableName() {
	s.T.Expect(identitypermissiontype.PermissionType{}.TableName()).ToEqual("IdentityPermissionTypes")
}

func TestPermissionEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &PermissionEntitySuite{}).Run()
}
type PermissionEntitySuite struct{ goosetest.Suite }
func (s *PermissionEntitySuite) TestTableName() {
	s.T.Expect(identitypermission.Permission{}.TableName()).ToEqual("IdentityPermissions")
}

func TestTokenEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &TokenEntitySuite{}).Run()
}
type TokenEntitySuite struct{ goosetest.Suite }
func (s *TokenEntitySuite) TestTableName() {
	s.T.Expect(identitytoken.Token{}.TableName()).ToEqual("IdentityTokens")
}

func TestDeviceEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &DeviceEntitySuite{}).Run()
}
type DeviceEntitySuite struct{ goosetest.Suite }
func (s *DeviceEntitySuite) TestTableName() {
	s.T.Expect(identitydevice.Device{}.TableName()).ToEqual("IdentityDevices")
}

func TestDeviceLogEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &DeviceLogEntitySuite{}).Run()
}
type DeviceLogEntitySuite struct{ goosetest.Suite }
func (s *DeviceLogEntitySuite) TestTableName() {
	s.T.Expect(identitydevicelog.DeviceLog{}.TableName()).ToEqual("IdentityDeviceLogs")
}

func TestDeviceSessionEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &DeviceSessionEntitySuite{}).Run()
}
type DeviceSessionEntitySuite struct{ goosetest.Suite }
func (s *DeviceSessionEntitySuite) TestTableName() {
	s.T.Expect(identitydevicesession.DeviceSession{}.TableName()).ToEqual("IdentityDeviceSessions")
}

func TestAttributeEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AttributeEntitySuite{}).Run()
}
type AttributeEntitySuite struct{ goosetest.Suite }
func (s *AttributeEntitySuite) TestTableName() {
	s.T.Expect(identityattribute.Attribute{}.TableName()).ToEqual("IdentityAttributes")
}

func TestInviteEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &InviteEntitySuite{}).Run()
}
type InviteEntitySuite struct{ goosetest.Suite }
func (s *InviteEntitySuite) TestTableName() {
	s.T.Expect(identityinvite.Invite{}.TableName()).ToEqual("IdentityInvites")
}

func TestProviderEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &ProviderEntitySuite{}).Run()
}
type ProviderEntitySuite struct{ goosetest.Suite }
func (s *ProviderEntitySuite) TestTableName() {
	s.T.Expect(identityprovider.Provider{}.TableName()).ToEqual("IdentityProviders")
}

func TestProviderLogEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &ProviderLogEntitySuite{}).Run()
}
type ProviderLogEntitySuite struct{ goosetest.Suite }
func (s *ProviderLogEntitySuite) TestTableName() {
	s.T.Expect(identityproviderlog.ProviderLog{}.TableName()).ToEqual("IdentityProviderLogs")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}
type AppModuleSuite struct{ goosetest.Suite }
func (s *AppModuleSuite) TestMigrations_Count() {
	s.T.Expect(len(app.Migrations)).ToEqual(17)
}
func (s *AppModuleSuite) TestName() {
	s.T.Expect(app.Name).ToEqual("identity")
}

// Auth service unit tests (no DB — validates token sign/verify round-trip)

func TestAuthService_ValidateToken(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AuthServiceSuite{}).Run()
}

type AuthServiceSuite struct{ goosetest.Suite }

func (s *AuthServiceSuite) TestValidateToken_Valid() {
	svc := &pkg.AuthService{}
	token, err := auth.Sign(map[string]any{"sub": "user-1"}, "changeme", 15*time.Minute)
	s.T.Expect(err).ToBeNil()
	claims, err := svc.ValidateToken(token)
	s.T.Expect(err).ToBeNil()
	s.T.Expect(claims["sub"]).ToEqual("user-1")
}

func (s *AuthServiceSuite) TestValidateToken_Invalid() {
	svc := &pkg.AuthService{}
	_, err := svc.ValidateToken("not.a.token")
	s.T.Expect(err).Not().ToBeNil()
}
