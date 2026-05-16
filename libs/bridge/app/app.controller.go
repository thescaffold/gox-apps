package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/bridge/src/app.controller.ts. The TS
// controller has no public endpoints beyond `@Get() getHello` — everything
// else is driven via subscriptions (see index.go).
type AppController struct {
	appService *AppService `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) — plain string,
// no envelope title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}
