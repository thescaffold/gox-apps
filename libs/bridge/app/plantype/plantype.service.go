package plantype

import "github.com/thescaffold/gox-packages/libs/core/crud"

// PlanTypeService exposes the TS-equivalent persistence helpers used by
// LicenseController.process and the PlanTypeController.AfterList sort.
type PlanTypeService struct {
	entity *PlanTypeEntity `inject:""`
}

// Entity returns the underlying entity.
func (s *PlanTypeService) Entity() *PlanTypeEntity { return s.entity }

// Save inserts a row. Mirrors TS `planTypeService.withContext(ctx).save(row)`.
func (s *PlanTypeService) Save(row *PlanType) error {
	return s.entity.Insert(row)
}

// Delete soft-deletes rows matching `match`. Mirrors TS
// `planTypeService.withContext(ctx).delete(match)`.
func (s *PlanTypeService) Delete(match map[string]any) (int64, error) {
	where, args := crud.MapToWhere(match)
	return crud.SoftDelete(s.entity.Entity, where, args...)
}
