package event
import "github.com/awesome-goose/goose/modules/router"
var ROUTES = router.ForRoutes(router.Resource("events", EventController{}).All())
