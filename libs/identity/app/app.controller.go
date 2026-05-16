package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/identity/src/app.controller.ts. TS
// declares a large surface (login/logout/initiate/verify/secret/reset-secret/
// session/oauth/provider authorize/etc) that depends on AuthService /
// ProviderService.authorize / TokenService.verify / many other primitives
// that are mostly unported in gox. The OAuth/auth surface lives in
// gox-apps/libs/identity/pkg/* (separate controllers). gox currently surfaces
// only GET / hello here.
type AppController struct {
	appService *AppService `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) — plain string,
// no envelope title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}
