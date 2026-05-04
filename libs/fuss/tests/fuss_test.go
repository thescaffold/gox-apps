package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps-fuss/app"
	fusshist "github.com/thescaffold/gox-apps-fuss/app/history"
	fusstok "github.com/thescaffold/gox-apps-fuss/app/token"
	fusstoklog "github.com/thescaffold/gox-apps-fuss/app/tokenlog"
)

func TestTokenEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &TokenEntitySuite{}).Run()
}

type TokenEntitySuite struct{ goosetest.Suite }

func (s *TokenEntitySuite) TestToken_TableName() {
	s.T.Expect(fusstok.Token{}.TableName()).ToEqual("FussTokens")
}

func (s *TokenEntitySuite) TestToken_Fields() {
	tok := &fusstok.Token{UserId: "u-1", EntityName: "User", EntityId: "e-1"}
	s.T.Expect(tok.UserId).ToEqual("u-1")
	s.T.Expect(tok.EntityName).ToEqual("User")
}

func TestTokenLogEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &TokenLogEntitySuite{}).Run()
}

type TokenLogEntitySuite struct{ goosetest.Suite }

func (s *TokenLogEntitySuite) TestTokenLog_TableName() {
	s.T.Expect(fusstoklog.TokenLog{}.TableName()).ToEqual("FussTokenLogs")
}

func (s *TokenLogEntitySuite) TestTokenLog_Fields() {
	tl := &fusstoklog.TokenLog{TokenId: "tok-1", Value: "hello world"}
	s.T.Expect(tl.TokenId).ToEqual("tok-1")
	s.T.Expect(tl.Value).ToEqual("hello world")
}

func TestHistoryEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &HistoryEntitySuite{}).Run()
}

type HistoryEntitySuite struct{ goosetest.Suite }

func (s *HistoryEntitySuite) TestHistory_TableName() {
	s.T.Expect(fusshist.History{}.TableName()).ToEqual("FussHistories")
}

func (s *HistoryEntitySuite) TestHistory_Fields() {
	h := &fusshist.History{UserId: "u-1", Query: "test query", Type: "global"}
	s.T.Expect(h.Query).ToEqual("test query")
	s.T.Expect(h.Type).ToEqual("global")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ goosetest.Suite }

func (s *AppModuleSuite) TestMigrations_Count() {
	s.T.Expect(len(app.Migrations)).ToEqual(3)
}

func (s *AppModuleSuite) TestName_IsFuss() {
	s.T.Expect(app.Name).ToEqual("fuss")
}

func (s *AppModuleSuite) TestSubscriptions_HasAfterInsert() {
	_, ok := app.Subscriptions["*.*.*.after-insert"]
	s.T.Expect(ok).ToEqual(true)
}

func (s *AppModuleSuite) TestSubscriptions_HasAfterUpdate() {
	_, ok := app.Subscriptions["*.*.*.after-update"]
	s.T.Expect(ok).ToEqual(true)
}

func (s *AppModuleSuite) TestSubscriptions_HasAfterDelete() {
	_, ok := app.Subscriptions["*.*.*.after-delete"]
	s.T.Expect(ok).ToEqual(true)
}
