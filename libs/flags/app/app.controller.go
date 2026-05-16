package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/flags/src/app.controller.ts. Endpoints:
//
//	GET  /           getHello
//	POST /register   register (upsert flag definitions)
//	POST /log        log (record a flag usage)
//	POST /status     status (AND-of-OR allowed check)
//	POST /limit      limit (return allowed/limit/usage triple)
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) with no envelope.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// Register handles POST /register. Mirrors TS AppController.register():
// userId/clientId/workspaceId on each FlagDef are pulled from the request
// context (TS spreads `{ ...flag, userId: user.id, ... }`).
func (c *AppController) Register(dto *RegisterDto) types.Output {
	pref := dto.NTX.Preference
	owner := FlagOwner{
		UserId:      dto.NTX.UserID,
		ClientId:    dto.NTX.ClientID,
		WorkspaceId: dto.NTX.WorkspaceID,
	}
	out, err := c.appService.Register(dto.EnvironmentType, dto.Environment, owner, dto.Flags)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.flags.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(out,
		c.lang.Translate("apps.flags.app.title", nil, pref),
		c.lang.Translate("apps.flags.app.register.success", nil, pref),
		nil)
}

// Log handles POST /log. Mirrors TS AppController.log(): writes a FlagLog row.
func (c *AppController) Log(dto *StatusDto) types.Output {
	pref := dto.NTX.Preference
	q := buildLogQuery(dto)
	limit := 1
	if dto.Limit != nil && *dto.Limit > 0 {
		limit = *dto.Limit
	}
	out, err := c.appService.Log(dto.Environment, q, limit)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.flags.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(out,
		c.lang.Translate("apps.flags.app.title", nil, pref),
		c.lang.Translate("apps.flags.app.log.success", nil, pref),
		nil)
}

// Status handles POST /status. Mirrors TS AppController.status() AND-of-OR logic:
// `names` is an OR-list of AND-groups; the final allowed value is true when
// any group's flags are all individually allowed.
func (c *AppController) Status(dto *StatusDto) types.Output {
	pref := dto.NTX.Preference

	var groups [][]string
	if dto.Name != "" {
		groups = append(groups, []string{dto.Name})
	}
	if len(dto.Names) > 0 {
		groups = append(groups, dto.Names...)
	}

	final := false
	for _, group := range groups {
		andResult := true
		for _, name := range group {
			q := buildLogQuery(dto)
			q.Name = name
			ok, err := c.appService.Status(dto.Environment, q)
			if err != nil {
				return response.InternalServerError(
					c.lang.Translate("apps.flags.app.title", nil, pref),
					err.Error(),
				)
			}
			if !ok {
				andResult = false
				break
			}
		}
		if andResult {
			final = true
			break
		}
	}
	return response.Success(final,
		c.lang.Translate("apps.flags.app.title", nil, pref),
		c.lang.Translate("apps.flags.app.status.success", nil, pref),
		nil)
}

// Limit handles POST /limit. Returns the allowed/limit/usage triple per flag.
func (c *AppController) Limit(dto *StatusDto) types.Output {
	pref := dto.NTX.Preference
	q := buildLogQuery(dto)
	q.Name = dto.Name
	out, err := c.appService.Limit(dto.Environment, q)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.flags.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(out,
		c.lang.Translate("apps.flags.app.title", nil, pref),
		c.lang.Translate("apps.flags.app.limit.success", nil, pref),
		nil)
}

// buildLogQuery resolves the user/client/workspace from explicit DTO fields
// or falls back to the request context (mirroring TS body fallbacks).
func buildLogQuery(dto *StatusDto) LogQuery {
	q := LogQuery{Name: dto.Name}
	if dto.Level != nil {
		q.Level = *dto.Level
	}
	if dto.UserId != nil {
		q.UserId = *dto.UserId
	} else {
		q.UserId = dto.NTX.UserID
	}
	if dto.ClientId != nil {
		q.ClientId = *dto.ClientId
	} else {
		q.ClientId = dto.NTX.ClientID
	}
	if dto.WorkspaceId != nil {
		q.WorkspaceId = *dto.WorkspaceId
	} else {
		q.WorkspaceId = dto.NTX.WorkspaceID
	}
	return q
}
