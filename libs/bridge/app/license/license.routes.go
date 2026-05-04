package license
import "github.com/awesome-goose/goose/modules/router"
var ROUTES = router.ForRoutes(router.Resource("licenses", LicenseController{}).All())
