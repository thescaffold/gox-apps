package tests

import (
	"testing"

	test "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/audit/app"
	auditlog "github.com/thescaffold/gox-apps/libs/audit/app/log"
)

func TestLogEntity(t *testing.T) {
	test.NewSuiteRunner(t, &LogEntitySuite{}).Run()
}

type LogEntitySuite struct {
	test.Suite
}

func (s *LogEntitySuite) TestLog_Initialization() {
	l := &auditlog.Log{
		UserId:      "user-1",
		ClientId:    "client-1",
		WorkspaceId: "ws-1",
		Group:       "apps",
		Service:     "identity",
		EntityId:    "entity-1",
		EntityName:  "user",
		Action:      auditlog.LogActionTypeCreate,
	}
	s.T.Expect(l.UserId).ToEqual("user-1")
	s.T.Expect(l.Action).ToEqual("create")
}

func (s *LogEntitySuite) TestLog_TableName() {
	s.T.Expect(auditlog.Log{}.TableName()).ToEqual("AuditLogs")
}

func (s *LogEntitySuite) TestLogActionTypes() {
	s.T.Expect(auditlog.LogActionTypeCreate).ToEqual("create")
	s.T.Expect(auditlog.LogActionTypeRead).ToEqual("read")
	s.T.Expect(auditlog.LogActionTypeUpdate).ToEqual("update")
	s.T.Expect(auditlog.LogActionTypeDelete).ToEqual("delete")
}

func TestAppService(t *testing.T) {
	test.NewSuiteRunner(t, &AppServiceSuite{}).Run()
}

type AppServiceSuite struct {
	test.Suite
}

func (s *AppServiceSuite) TestGetHello_NotEmpty() {
	svc := &app.AppService{}
	s.T.Expect(svc.GetHello()).Not().ToBeEmpty()
}

func (s *AppServiceSuite) TestHandleEvent_SkipsUnsafeEntity() {
	svc := &app.AppService{}
	err := svc.HandleEvent(map[string]any{
		"entityName": "AuditLogs",
	})
	s.T.Expect(err).ToBeNil()
}

func (s *AppServiceSuite) TestHandleEvent_SkipsMissingMeta() {
	svc := &app.AppService{}
	err := svc.HandleEvent(map[string]any{
		"entityName": "user",
		"meta":       nil,
	})
	s.T.Expect(err).ToBeNil()
}

func (s *AppServiceSuite) TestHandleEvent_SkipsEmptyEntityName() {
	svc := &app.AppService{}
	err := svc.HandleEvent(map[string]any{"entityName": ""})
	s.T.Expect(err).ToBeNil()
}

func TestAppModule(t *testing.T) {
	test.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct {
	test.Suite
}

func (s *AppModuleSuite) TestName() {
	s.T.Expect(app.Name).ToEqual("audit")
}

func (s *AppModuleSuite) TestMigrations_NotEmpty() {
	s.T.Expect(len(app.Migrations)).ToEqual(1)
}

func (s *AppModuleSuite) TestSubscriptions_NotEmpty() {
	s.T.Expect(len(app.Subscriptions)).ToEqual(6)
}
