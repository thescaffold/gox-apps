package tests

import (
	"testing"

	"github.com/thescaffold/gox-apps-figs/app"
	"github.com/thescaffold/gox-apps-figs/app/file"
	test "github.com/awesome-goose/goose/testing"
)

func TestFileEntity(t *testing.T) {
	test.NewSuiteRunner(t, &FileEntitySuite{}).Run()
}

type FileEntitySuite struct {
	test.Suite
}

func (s *FileEntitySuite) TestFile_Initialization() {
	f := &file.File{Name: "test.csv"}
	s.T.Expect(f.Name).ToEqual("test.csv")
}

func (s *FileEntitySuite) TestFile_NullableFields() {
	f := &file.File{Name: "test.csv"}
	s.T.Expect(f.Url).ToBeNil()
	s.T.Expect(f.Raw).ToBeNil()
	s.T.Expect(f.Status).ToBeNil()
}

func (s *FileEntitySuite) TestFile_TableName() {
	f := file.File{}
	s.T.Expect(f.TableName()).ToEqual("FigsFiles")
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

func (s *AppModuleSuite) TestName_IsFigs() {
	s.T.Expect(app.Name).ToEqual("figs")
}

func (s *AppModuleSuite) TestSubscriptions_NotEmpty() {
	s.T.Expect(len(app.Subscriptions)).ToEqual(1)
}
