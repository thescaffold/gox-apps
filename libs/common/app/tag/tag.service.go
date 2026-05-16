package tag

import "github.com/thescaffold/gox-packages/libs/core/crud"

// TagService exposes the TS-equivalent persistence helpers used by the
// sync/attach/detach flow in TagTypeController. Each method delegates to the
// generic crud helpers (gox-packages/libs/core/crud/helpers.go) which compose
// over goose's `*sql.Entity[Tag]`.
type TagService struct {
	entity *TagEntity `inject:""`
}

// Entity returns the underlying entity, useful when callers need access to a
// method not exposed on the service.
func (s *TagService) Entity() *TagEntity { return s.entity }

// UpdateOrCreate mirrors TS `tagService.withContext(ctx).updateOrCreate(match)`.
// match contains the row identifiers; the same map is used as the payload, so
// a freshly-created row carries those fields.
func (s *TagService) UpdateOrCreate(match map[string]any) (*Tag, error) {
	row := &Tag{}
	if v, ok := match["user_id"].(string); ok {
		row.UserId = &v
	}
	if v, ok := match["client_id"].(string); ok {
		row.ClientId = &v
	}
	if v, ok := match["workspace_id"].(string); ok {
		row.WorkspaceId = &v
	}
	if v, ok := match["group_name"].(string); ok {
		row.GroupName = v
	}
	if v, ok := match["service_name"].(string); ok {
		row.ServiceName = v
	}
	if v, ok := match["entity_name"].(string); ok {
		row.EntityName = v
	}
	if v, ok := match["entity_id"].(string); ok {
		row.EntityId = v
	}
	if v, ok := match["type_id"].(string); ok {
		row.TypeId = v
	}
	out, _, err := crud.UpdateOrCreate(s.entity.Entity, match, row)
	return out, err
}

// SoftDelete mirrors TS `tagService.withContext(ctx).softDelete(match)`.
// Returns the number of rows affected.
func (s *TagService) SoftDelete(match map[string]any) (int64, error) {
	where, args := crud.MapToWhere(match)
	return crud.SoftDelete(s.entity.Entity, where, args...)
}

// Count mirrors TS `tagService.withContext(ctx).count({where: [match]})`.
func (s *TagService) Count(match map[string]any) (int64, error) {
	where, args := crud.MapToWhere(match)
	return crud.Count(s.entity.Entity, where, args...)
}
