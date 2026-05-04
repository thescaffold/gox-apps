package preference
import "github.com/awesome-goose/goose/modules/router"
var ROUTES = router.ForRoutes(router.Resource("preferences", PreferenceController{}).All())
