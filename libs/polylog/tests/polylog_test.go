package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/polylog/app"
	polylogevent "github.com/thescaffold/gox-apps/libs/polylog/app/event"
	polylogsourcetype "github.com/thescaffold/gox-apps/libs/polylog/app/sourcetype"
)

func TestSourceTypeEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &SourceTypeEntitySuite{}).Run()
}

type SourceTypeEntitySuite struct{ goosetest.Suite }

func (s *SourceTypeEntitySuite) TestTableName() {
	s.T.Expect(polylogsourcetype.SourceType{}.TableName()).ToEqual("PolylogSourceTypes")
}

func TestEventEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &EventEntitySuite{}).Run()
}

type EventEntitySuite struct{ goosetest.Suite }

func (s *EventEntitySuite) TestTableName() {
	s.T.Expect(polylogevent.Event{}.TableName()).ToEqual("PolylogEvents")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ goosetest.Suite }

func (s *AppModuleSuite) TestMigrations_Count() {
	// 8 = source-types, sources, sink-types, sinks, channels, events, event-logs, configs.
	s.T.Expect(len(app.Migrations)).ToEqual(8)
}

func (s *AppModuleSuite) TestName() {
	s.T.Expect(app.Name).ToEqual("polylog")
}
