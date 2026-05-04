package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps-notification/app"
	notificationlog "github.com/thescaffold/gox-apps-notification/app/log"
	notificationmessage "github.com/thescaffold/gox-apps-notification/app/message"
	notificationrule "github.com/thescaffold/gox-apps-notification/app/rule"
	notificationtemplate "github.com/thescaffold/gox-apps-notification/app/template"
)

func TestTemplateEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &TemplateEntitySuite{}).Run()
}

type TemplateEntitySuite struct{ goosetest.Suite }

func (s *TemplateEntitySuite) TestTableName() {
	s.T.Expect(notificationtemplate.Template{}.TableName()).ToEqual("NotificationTemplates")
}

func TestRuleEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &RuleEntitySuite{}).Run()
}

type RuleEntitySuite struct{ goosetest.Suite }

func (s *RuleEntitySuite) TestTableName() {
	s.T.Expect(notificationrule.Rule{}.TableName()).ToEqual("NotificationRules")
}

func TestMessageEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &MessageEntitySuite{}).Run()
}

type MessageEntitySuite struct{ goosetest.Suite }

func (s *MessageEntitySuite) TestTableName() {
	s.T.Expect(notificationmessage.Message{}.TableName()).ToEqual("NotificationMessages")
}

func TestLogEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &LogEntitySuite{}).Run()
}

type LogEntitySuite struct{ goosetest.Suite }

func (s *LogEntitySuite) TestTableName() {
	s.T.Expect(notificationlog.Log{}.TableName()).ToEqual("NotificationLogs")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ goosetest.Suite }

func (s *AppModuleSuite) TestMigrations_Count() {
	s.T.Expect(len(app.Migrations)).ToEqual(4)
}

func (s *AppModuleSuite) TestName() {
	s.T.Expect(app.Name).ToEqual("notification")
}
