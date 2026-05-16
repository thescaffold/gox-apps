// Package flag hosts the gox-apps FlagService — a thin DI wrapper around
// FlagEntity used by the lib's own controllers, and by cross-app callers
// (capital uses this from FindOrCreateUserPlan / ChangeUserPlan) to seed
// user-scoped flags out of plan-type meta.
package flag

import (
	"encoding/json"
)

// FlagService persists Flag rows for cross-app callers + drives the local
// CRUD surface. Register is the cross-app entry point — mirrors TS
// `flagService.register(flags)` from gox-packages/libs/flags.
type FlagService struct {
	entity *FlagEntity `inject:""`
}

// RegisterDto is the shape capital + identity pass through to Register.
// All fields mirror TS Flag — meta and reference are optional.
type RegisterDto struct {
	UserId      string
	ClientId    string
	WorkspaceId string
	Name        string
	Limit       int
	Priority    int
	Level       string
	Reference   string
	Meta        any
	Status      string
}

// Register upserts a batch of Flag rows. For each entry, lookup is by
// (user_id, client_id, workspace_id, name) — if exists, the row is updated
// in place; otherwise inserted fresh. Returns true when every row succeeded.
// Mirrors TS `flagService.register([...])` semantics: a single failure causes
// false to bubble up so callers can rollback.
func (s *FlagService) Register(flags []RegisterDto) (bool, error) {
	if s.entity == nil || len(flags) == 0 {
		return true, nil
	}
	for _, f := range flags {
		if f.Name == "" {
			continue
		}
		status := f.Status
		var statusPtr *string
		if status != "" {
			statusPtr = &status
		}
		var metaJSON json.RawMessage
		if f.Meta != nil {
			if b, err := json.Marshal(f.Meta); err == nil {
				metaJSON = b
			}
		}
		existing, _ := s.entity.First(
			`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "name" = ?`,
			f.UserId, f.ClientId, f.WorkspaceId, f.Name,
		)
		if existing != nil {
			existing.Limit = f.Limit
			existing.Priority = f.Priority
			existing.Level = f.Level
			existing.Meta = metaJSON
			existing.Status = statusPtr
			if _, err := s.entity.Update(existing, `"id" = ?`, existing.Id); err != nil {
				return false, err
			}
			continue
		}
		row := &Flag{
			UserId:      f.UserId,
			ClientId:    f.ClientId,
			WorkspaceId: f.WorkspaceId,
			Name:        f.Name,
			Limit:       f.Limit,
			Priority:    f.Priority,
			Level:       f.Level,
			Meta:        metaJSON,
			Status:      statusPtr,
		}
		if err := s.entity.Insert(row); err != nil {
			return false, err
		}
	}
	return true, nil
}

// Status mirrors TS flagService.status — looks up the flag row for the
// (name, level, owner-tuple) tuple and returns true when its status is
// `active`. Returns true (permissive) when no row is found so brand-new
// callers aren't blocked before a flag has been registered.
func (s *FlagService) Status(name, level, userId, clientId, workspaceId string) bool {
	if s.entity == nil || name == "" {
		return true
	}
	where := `"name" = ?`
	args := []any{name}
	if level != "" {
		where += ` AND "level" = ?`
		args = append(args, level)
	}
	if userId != "" {
		where += ` AND "user_id" = ?`
		args = append(args, userId)
	}
	if clientId != "" {
		where += ` AND "client_id" = ?`
		args = append(args, clientId)
	}
	if workspaceId != "" {
		where += ` AND "workspace_id" = ?`
		args = append(args, workspaceId)
	}
	row, _ := s.entity.First(where, args...)
	if row == nil {
		return true
	}
	if row.Status == nil {
		return true
	}
	return *row.Status == "active"
}

// FromMeta converts a planType.meta.flags map (TS-shaped) into a slice of
// RegisterDto for the supplied owner triple. Booleans become "active"/"inactive"
// with limit=0; numerics become the numeric limit with status="active".
// Mirrors the TS object-entries iteration in capital/app.service.ts.
func FromMeta(userId, clientId, workspaceId string, meta map[string]any) []RegisterDto {
	out := make([]RegisterDto, 0, len(meta))
	for name, value := range meta {
		limit := 0
		status := "active"
		switch v := value.(type) {
		case bool:
			if v {
				status = "active"
			} else {
				status = "inactive"
			}
		case float64:
			limit = int(v)
		case int:
			limit = v
		}
		out = append(out, RegisterDto{
			UserId:      userId,
			ClientId:    clientId,
			WorkspaceId: workspaceId,
			Name:        name,
			Limit:       limit,
			Priority:    100,
			Level:       "user",
			Status:      status,
		})
	}
	return out
}
