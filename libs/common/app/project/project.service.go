package project

import "github.com/thescaffold/gox-packages/libs/core/crud"

// ProjectService exposes the TS-equivalent persistence helpers used by the
// sync/attach/detach flow in ProjectTypeController.
type ProjectService struct {
	entity *ProjectEntity `inject:""`
}

// Entity returns the underlying entity, useful when callers need access to a
// method not exposed on the service.
func (s *ProjectService) Entity() *ProjectEntity { return s.entity }

// UpdateOrCreate mirrors TS `projectService.withContext(ctx).updateOrCreate(match)`.
func (s *ProjectService) UpdateOrCreate(match map[string]any) (*Project, error) {
	row := &Project{}
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

// SoftDelete mirrors TS `projectService.withContext(ctx).softDelete(match)`.
func (s *ProjectService) SoftDelete(match map[string]any) (int64, error) {
	where, args := crud.MapToWhere(match)
	return crud.SoftDelete(s.entity.Entity, where, args...)
}

// Count mirrors TS `projectService.withContext(ctx).count({where:[match]})`.
func (s *ProjectService) Count(match map[string]any) (int64, error) {
	where, args := crud.MapToWhere(match)
	return crud.Count(s.entity.Entity, where, args...)
}
