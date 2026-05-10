package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/controller/app"
	ctrlreq "github.com/thescaffold/gox-apps/libs/controller/app/request"
	ctrlroute "github.com/thescaffold/gox-apps/libs/controller/app/route"
)

func TestRouteEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &RouteEntitySuite{}).Run()
}

type RouteEntitySuite struct{ goosetest.Suite }

func (s *RouteEntitySuite) TestRoute_TableName() {
	s.T.Expect(ctrlroute.Route{}.TableName()).ToEqual("ControllerRoutes")
}

func (s *RouteEntitySuite) TestRoute_Fields() {
	r := &ctrlroute.Route{Group: "apps", Service: "identity", Upstream: "http://identity:3000"}
	s.T.Expect(r.Group).ToEqual("apps")
	s.T.Expect(r.Upstream).ToEqual("http://identity:3000")
}

func TestRequestEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &RequestEntitySuite{}).Run()
}

type RequestEntitySuite struct{ goosetest.Suite }

func (s *RequestEntitySuite) TestRequest_TableName() {
	s.T.Expect(ctrlreq.Request{}.TableName()).ToEqual("ControllerRequests")
}

func (s *RequestEntitySuite) TestRequest_Fields() {
	r := &ctrlreq.Request{Method: "GET", Url: "/api/users"}
	s.T.Expect(r.Method).ToEqual("GET")
	s.T.Expect(r.Url).ToEqual("/api/users")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ goosetest.Suite }

func (s *AppModuleSuite) TestMigrations_Count() {
	s.T.Expect(len(app.Migrations)).ToEqual(2)
}

func (s *AppModuleSuite) TestName_IsController() {
	s.T.Expect(app.Name).ToEqual("controller")
}

func (s *AppModuleSuite) TestSubscriptions_HasRouteRegister() {
	_, ok := app.Subscriptions["apps.controller.route.register"]
	s.T.Expect(ok).ToEqual(true)
}
