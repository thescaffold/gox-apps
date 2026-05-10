package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct {
	appService *AppService `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "flags", "ok", nil)
}

// Register handles POST /register. Mirrors TS AppController.register().
func (c *AppController) Register(dto *RegisterDto) types.Output {
	owner := FlagOwner{
		UserId:      dto.NTX.UserID,
		ClientId:    dto.NTX.ClientID,
		WorkspaceId: dto.NTX.WorkspaceID,
	}
	out, err := c.appService.Register(dto.EnvironmentType, dto.Environment, owner, dto.Flags)
	if err != nil {
		return response.InternalServerError("flags", err.Error())
	}
	return response.Success(out, "flags", "ok", nil)
}

// Log handles POST /log.
func (c *AppController) Log(dto *StatusDto) types.Output {
	q := buildLogQuery(dto)
	limit := 1
	if dto.Limit != nil && *dto.Limit > 0 {
		limit = *dto.Limit
	}
	out, err := c.appService.Log(dto.Environment, q, limit)
	if err != nil {
		return response.InternalServerError("flags", err.Error())
	}
	return response.Success(out, "flags", "ok", nil)
}

// Status handles POST /status. Mirrors TS AppController.status() AND-of-OR logic:
// names is an OR-list of AND-groups; success when any group's flags are all enabled.
func (c *AppController) Status(dto *StatusDto) types.Output {
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
				return response.InternalServerError("flags", err.Error())
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
	return response.Success(final, "flags", "ok", nil)
}

// Limit handles POST /limit.
func (c *AppController) Limit(dto *StatusDto) types.Output {
	q := buildLogQuery(dto)
	q.Name = dto.Name
	out, err := c.appService.Limit(dto.Environment, q)
	if err != nil {
		return response.InternalServerError("flags", err.Error())
	}
	return response.Success(out, "flags", "ok", nil)
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
