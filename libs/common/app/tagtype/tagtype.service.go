package tagtype

import "github.com/thescaffold/gox-packages/libs/core/crud"

// TagTypeService exposes the TS-equivalent persistence helpers used by the
// sync/attach/detach flow in TagTypeController. Each method delegates to the
// generic crud helpers (gox-packages/libs/core/crud/helpers.go) which compose
// over goose's `*sql.Entity[TagType]`.
type TagTypeService struct {
	entity *TagTypeEntity `inject:""`
}

// Entity returns the underlying entity, useful when callers need access to a
// method not exposed on the service.
func (s *TagTypeService) Entity() *TagTypeEntity { return s.entity }

// FindOne mirrors TS `tagTypeService.withContext(ctx).findOne({where:match})`.
func (s *TagTypeService) FindOne(match map[string]any) (*TagType, error) {
	where, args := crud.MapToWhere(match)
	return s.entity.First(where, args...)
}

// FindOrCreate mirrors TS `tagTypeService.withContext(ctx).findOrCreate(match)`.
// match identifies the existing row; on miss a new row is inserted using the
// same fields.
func (s *TagTypeService) FindOrCreate(match map[string]any) (*TagType, error) {
	row := &TagType{}
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
	out, _, err := crud.FindOrCreate(s.entity.Entity, match, row)
	return out, err
}

// SoftDelete mirrors TS `tagTypeService.withContext(ctx).softDelete(match)`.
func (s *TagTypeService) SoftDelete(match map[string]any) (int64, error) {
	where, args := crud.MapToWhere(match)
	return crud.SoftDelete(s.entity.Entity, where, args...)
}
