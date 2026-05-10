package tests

import (
	"strings"
	"testing"

	test "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps-assets/app"
	assetsfile "github.com/thescaffold/gox-apps-assets/app/file"
)

func TestFileEntity(t *testing.T) {
	test.NewSuiteRunner(t, &FileEntitySuite{}).Run()
}

type FileEntitySuite struct {
	test.Suite
}

func (s *FileEntitySuite) TestFile_Initialization() {
	f := &assetsfile.File{Type: "svg", Bucket: "dynamic", Name: "test"}
	s.T.Expect(f.Type).ToEqual("svg")
	s.T.Expect(f.Bucket).ToEqual("dynamic")
}

func (s *FileEntitySuite) TestFile_TableName() {
	s.T.Expect(assetsfile.File{}.TableName()).ToEqual("AssetsFiles")
}

func (s *FileEntitySuite) TestFile_NullableFields() {
	f := &assetsfile.File{Type: "png", Bucket: "uploads", Name: "photo"}
	s.T.Expect(f.Raw).ToBeNil()
	s.T.Expect(f.Status).ToBeNil()
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

func (s *AppServiceSuite) TestGenerateSVG_ContainsSVGTag() {
	svc := &app.AppService{}
	svg := svc.GenerateDynamicSVG("test", "pixel", nil, 100, false)
	s.T.Expect(svg).ToContainString("<svg")
	s.T.Expect(svg).ToContainString("</svg>")
}

func (s *AppServiceSuite) TestGenerateSVG_NameAppearsInTitle() {
	// The new core image.Service inserts the name into <title>, mirroring
	// ntx-apps assets ImageService.new().
	svc := &app.AppService{}
	svg := svc.GenerateDynamicSVG("hello", "shapes", nil, 100, false)
	s.T.Expect(svg).ToContainString("<title>hello</title>")
}

func (s *AppServiceSuite) TestGenerateSVG_DefaultSize() {
	svc := &app.AppService{}
	svg := svc.GenerateDynamicSVG("x", "pixel", nil, 0, false)
	// Default size = 80 (matches TS imageService default).
	s.T.Expect(svg).ToContainString(`width="80"`)
}

func (s *AppServiceSuite) TestGenerateSVG_CustomColor() {
	svc := &app.AppService{}
	svg := svc.GenerateDynamicSVG("x", "pixel", []string{"#ff0000"}, 100, false)
	s.T.Expect(svg).ToContainString("#ff0000")
}

func (s *AppServiceSuite) TestRandomName_NotEmpty() {
	name := app.RandomName()
	s.T.Expect(name).Not().ToBeEmpty()
	s.T.Expect(strings.Contains(name, "-")).ToEqual(true)
}

func TestAppModule(t *testing.T) {
	test.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct {
	test.Suite
}

func (s *AppModuleSuite) TestName() {
	s.T.Expect(app.Name).ToEqual("assets")
}

func (s *AppModuleSuite) TestMigrations_NotEmpty() {
	s.T.Expect(len(app.Migrations)).ToEqual(1)
}
