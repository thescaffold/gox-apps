package tests

import (
	"testing"

	"github.com/thescaffold/gox-apps-blobs/app"
	"github.com/thescaffold/gox-apps-blobs/app/file"
	"github.com/thescaffold/gox-apps-blobs/app/page"
	test "github.com/awesome-goose/goose/testing"
)

func TestFileEntity(t *testing.T) {
	test.NewSuiteRunner(t, &FileEntitySuite{}).Run()
}

type FileEntitySuite struct {
	test.Suite
}

func (s *FileEntitySuite) TestFile_Initialization() {
	f := &file.File{Name: "photo.jpg", Type: "image", Bucket: "avatars"}
	s.T.Expect(f.Name).ToEqual("photo.jpg")
	s.T.Expect(f.Type).ToEqual("image")
	s.T.Expect(f.Bucket).ToEqual("avatars")
}

func (s *FileEntitySuite) TestFile_NullableFields() {
	f := &file.File{Name: "doc.pdf", Type: "document", Bucket: "docs"}
	s.T.Expect(f.Url).ToBeNil()
	s.T.Expect(f.ParentId).ToBeNil()
	s.T.Expect(f.Status).ToBeNil()
}

func (s *FileEntitySuite) TestFile_TableName() {
	s.T.Expect(file.File{}.TableName()).ToEqual("BlobsFiles")
}

func TestPageEntity(t *testing.T) {
	test.NewSuiteRunner(t, &PageEntitySuite{}).Run()
}

type PageEntitySuite struct {
	test.Suite
}

func (s *PageEntitySuite) TestPage_Initialization() {
	p := &page.Page{FileId: "file-1", Index: 0}
	s.T.Expect(p.FileId).ToEqual("file-1")
	s.T.Expect(p.Index).ToEqual(0)
}

func (s *PageEntitySuite) TestPage_TableName() {
	s.T.Expect(page.Page{}.TableName()).ToEqual("BlobsPages")
}

func TestAppModule(t *testing.T) {
	test.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct {
	test.Suite
}

func (s *AppModuleSuite) TestMigrations_NotEmpty() {
	s.T.Expect(len(app.Migrations)).ToEqual(2)
}

func (s *AppModuleSuite) TestName_IsBlobs() {
	s.T.Expect(app.Name).ToEqual("blobs")
}

func (s *AppModuleSuite) TestSubscriptions_Empty() {
	s.T.Expect(len(app.Subscriptions)).ToEqual(0)
}
