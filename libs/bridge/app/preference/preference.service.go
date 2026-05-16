package preference

import "github.com/thescaffold/gox-packages/libs/core/crud"

// PreferenceService exposes the TS-equivalent persistence helpers used by
// LicenseController.process (the child-sync flow). Each method delegates to
// the generic crud helpers (Phase 2).
type PreferenceService struct {
	entity *PreferenceEntity `inject:""`
}

// Entity returns the underlying entity for callers that need methods not
// surfaced on the service.
func (s *PreferenceService) Entity() *PreferenceEntity { return s.entity }

// Find mirrors TS `preferenceService.withContext(ctx).find({where:[match]})`.
func (s *PreferenceService) Find(match map[string]any) ([]Preference, error) {
	where, args := crud.MapToWhere(match)
	return s.entity.Some(where, args...)
}

// Save inserts a row. Mirrors TS `preferenceService.withContext(ctx).save(row)`.
func (s *PreferenceService) Save(row *Preference) error {
	return s.entity.Insert(row)
}

// Delete soft-deletes rows matching `match`. Mirrors TS
// `preferenceService.withContext(ctx).delete(match)`.
func (s *PreferenceService) Delete(match map[string]any) (int64, error) {
	where, args := crud.MapToWhere(match)
	return crud.SoftDelete(s.entity.Entity, where, args...)
}
