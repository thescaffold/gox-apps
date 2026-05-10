package channel

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(router.Resource("channels", ChannelController{}).All())
