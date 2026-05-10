package tests

import (
	"testing"

	test "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/statics/app"
	"github.com/thescaffold/gox-apps/libs/statics/app/list"
)

func TestListEntity(t *testing.T) {
	test.NewSuiteRunner(t, &ListEntitySuite{}).Run()
}

type ListEntitySuite struct {
	test.Suite
}

func (s *ListEntitySuite) TestList_Initialization() {
	l := &list.List{Key: list.ListKeyTypeCountry, Value: "Nigeria"}
	s.T.Expect(l.Key).ToEqual("country")
	s.T.Expect(l.Value).ToEqual("Nigeria")
}

func (s *ListEntitySuite) TestList_NullableFields() {
	l := &list.List{Key: list.ListKeyTypeCity, Value: "Lagos"}
	s.T.Expect(l.ParentId).ToBeNil()
	s.T.Expect(l.Code).ToBeNil()
	s.T.Expect(l.Status).ToBeNil()
}

func (s *ListEntitySuite) TestList_TableName() {
	l := list.List{}
	s.T.Expect(l.TableName()).ToEqual("StaticsLists")
}

func (s *ListEntitySuite) TestList_KeyTypes() {
	s.T.Expect(list.ListKeyTypeSetting).ToEqual("setting")
	s.T.Expect(list.ListKeyTypeCountry).ToEqual("country")
	s.T.Expect(list.ListKeyTypeCity).ToEqual("city")
	s.T.Expect(list.ListKeyTypeCurrency).ToEqual("currency")
	s.T.Expect(list.ListKeyTypeLanguage).ToEqual("language")
	s.T.Expect(list.ListKeyTypeTimezone).ToEqual("timezone")
	s.T.Expect(list.ListKeyTypeGender).ToEqual("gender")
	s.T.Expect(list.ListKeyTypeIndustry).ToEqual("industry")
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

func (s *AppServiceSuite) TestGetHello_MatchesTSLiteral() {
	// Mirrors TS app.service.ts which returns "Hello World!".
	svc := &app.AppService{}
	s.T.Expect(svc.GetHello()).ToEqual("Hello World!")
}

func TestAppModule(t *testing.T) {
	test.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct {
	test.Suite
}

func (s *AppModuleSuite) TestMigrations_NotEmpty() {
	s.T.Expect(app.Migrations).Not().ToBeNil()
	s.T.Expect(len(app.Migrations)).ToEqual(1)
}

func (s *AppModuleSuite) TestSeeders_NotEmpty() {
	s.T.Expect(app.Seeders).Not().ToBeNil()
	s.T.Expect(len(app.Seeders)).ToEqual(1)
}

func (s *AppModuleSuite) TestName_IsStatics() {
	s.T.Expect(app.Name).ToEqual("statics")
}
