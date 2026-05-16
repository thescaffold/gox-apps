package webhook

import "github.com/thescaffold/gox-packages/libs/core/crud"

// WebhookService exposes the TS-equivalent persistence helpers used by
// LicenseController.process (the child-sync flow).
type WebhookService struct {
	entity *WebhookEntity `inject:""`
}

// Entity returns the underlying entity.
func (s *WebhookService) Entity() *WebhookEntity { return s.entity }

// Save inserts a row. Mirrors TS `webhookService.withContext(ctx).save(row)`.
func (s *WebhookService) Save(row *Webhook) error {
	return s.entity.Insert(row)
}

// Delete soft-deletes rows matching `match`. Mirrors TS
// `webhookService.withContext(ctx).delete(match)`.
func (s *WebhookService) Delete(match map[string]any) (int64, error) {
	where, args := crud.MapToWhere(match)
	return crud.SoftDelete(s.entity.Entity, where, args...)
}
