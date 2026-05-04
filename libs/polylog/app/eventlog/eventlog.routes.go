package eventlog
import "github.com/awesome-goose/goose/modules/router"
var ROUTES = router.ForRoutes(router.Resource("eventlogs", EventLogController{}).All())
