package tests

import (
	"testing"
	"time"

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

// TestSubscriptions_RouteRegisterIsBound and TestSubscriptions_HourlyIsBound
// guard against regressing to the old no-op stubs (`func(_,_){}`). The TS
// subscription handlers do real work — route.register inserts Http+Ws Route
// rows; hourly heartbeat deletes Request rows — so the gox handlers must be
// real function references wired to AppService methods.
func (s *AppModuleSuite) TestSubscriptions_RouteRegisterIsBound() {
	handler, ok := app.Subscriptions["apps.controller.route.register"]
	s.T.Expect(ok).ToEqual(true)
	s.T.Expect(handler == nil).ToEqual(false)
}

func (s *AppModuleSuite) TestSubscriptions_HourlyIsBound() {
	handler, ok := app.Subscriptions["apps.cron.heartbeat.hourly"]
	s.T.Expect(ok).ToEqual(true)
	s.T.Expect(handler == nil).ToEqual(false)
}

// TestAppService_Housekeep_DoesNotPanic_WithoutDeps and
// TestAppService_RegisterRoutePayload_NonMapInputs verify the subscription-
// reachable AppService methods tolerate uninitialized / malformed input.
// They run from the event bus, so panics would crash the loop.
func (s *AppModuleSuite) TestAppService_Housekeep_DoesNotPanic_WithoutDeps() {
	svc := &app.AppService{}
	defer func() { s.T.Expect(recover() == nil).ToEqual(true) }()
	svc.Housekeep(time.Now().UTC().Add(-24 * time.Hour))
}

func (s *AppModuleSuite) TestAppService_RegisterRoutePayload_NonMapInputs() {
	svc := &app.AppService{}
	// nil, non-map and empty-map payloads must all no-op without panicking.
	defer func() { s.T.Expect(recover() == nil).ToEqual(true) }()
	s.T.Expect(svc.RegisterRoutePayload(nil)).ToBeNil()
	s.T.Expect(svc.RegisterRoutePayload("not-a-map")).ToBeNil()
	s.T.Expect(svc.RegisterRoutePayload(map[string]any{})).ToBeNil()
}
