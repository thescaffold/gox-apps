package sinktype
import "github.com/awesome-goose/goose/modules/router"
var ROUTES = router.ForRoutes(router.Resource("sinktypes", SinkTypeController{}).All())
