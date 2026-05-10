package tests

import (
	"testing"

	test "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/cache/app"
	"github.com/thescaffold/gox-apps/libs/cache/app/list"
)

func TestListEntity(t *testing.T) {
	test.NewSuiteRunner(t, &ListEntitySuite{}).Run()
}

type ListEntitySuite struct {
	test.Suite
}

func (s *ListEntitySuite) TestList_Initialization() {
	l := &list.List{Key: "user.theme", Group: "settings"}
	s.T.Expect(l.Key).ToEqual("user.theme")
	s.T.Expect(l.Group).ToEqual("settings")
}

func (s *ListEntitySuite) TestList_NullableFields() {
	l := &list.List{Key: "k", Group: "g"}
	s.T.Expect(l.ExpiredAt).ToBeNil()
	s.T.Expect(l.Status).ToBeNil()
}

func (s *ListEntitySuite) TestList_TableName() {
	s.T.Expect(list.List{}.TableName()).ToEqual("CacheLists")
}

func TestAppModule(t *testing.T) {
	test.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct {
	test.Suite
}

func (s *AppModuleSuite) TestMigrations_NotEmpty() {
	s.T.Expect(len(app.Migrations)).ToEqual(1)
}

func (s *AppModuleSuite) TestName_IsCache() {
	s.T.Expect(app.Name).ToEqual("cache")
}

func (s *AppModuleSuite) TestSubscriptions_Empty() {
	s.T.Expect(len(app.Subscriptions)).ToEqual(0)
}
