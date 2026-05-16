package projecttype

import "github.com/thescaffold/gox-packages/libs/core/crud"

// ProjectTypeService exposes the TS-equivalent persistence helpers used by
// the sync/attach/detach flow in ProjectTypeController.
type ProjectTypeService struct {
	entity *ProjectTypeEntity `inject:""`
}

// Entity returns the underlying entity, useful when callers need access to a
// method not exposed on the service.
func (s *ProjectTypeService) Entity() *ProjectTypeEntity { return s.entity }

// FindOne mirrors TS `projectTypeService.withContext(ctx).findOne({where:match})`.
func (s *ProjectTypeService) FindOne(match map[string]any) (*ProjectType, error) {
	where, args := crud.MapToWhere(match)
	return s.entity.First(where, args...)
}

// FindOrCreate mirrors TS `projectTypeService.withContext(ctx).findOrCreate(match, defaults)`.
// gox carries the optional thumbnailUrl in `defaults` so the new row picks up
// the dynamic-asset URL the caller (Phase 7.1 morph) computed.
func (s *ProjectTypeService) FindOrCreate(match map[string]any, defaults map[string]any) (*ProjectType, error) {
	row := &ProjectType{}
	if v, ok := match["user_id"].(string); ok {
		row.UserId = &v
	}
	if v, ok := match["client_id"].(string); ok {
		row.ClientId = &v
	}
	if v, ok := match["workspace_id"].(string); ok {
		row.WorkspaceId = &v
	}
	if v, ok := match["name"].(string); ok {
		row.Name = v
	}
	if defaults != nil {
		if v, ok := defaults["thumbnail_url"].(string); ok && v != "" {
			row.ThumbnailUrl = &v
		}
		if v, ok := defaults["desc"].(string); ok && v != "" {
			row.Desc = &v
		}
	}
	out, _, err := crud.FindOrCreate(s.entity.Entity, match, row)
	return out, err
}

// SoftDelete mirrors TS `projectTypeService.withContext(ctx).softDelete(match)`.
func (s *ProjectTypeService) SoftDelete(match map[string]any) (int64, error) {
	where, args := crud.MapToWhere(match)
	return crud.SoftDelete(s.entity.Entity, where, args...)
}
