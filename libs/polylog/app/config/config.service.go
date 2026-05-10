package config

import "encoding/json"

// ConfigService is the persistence layer for polylog configs.
type ConfigService struct {
	entity *ConfigEntity `inject:""`
}

// FindByTuple returns the config row matching the full identifier tuple, or nil
// when none exists. Mirrors TS findOne({where: tuple}).
func (s *ConfigService) FindByTuple(userId, clientId, workspaceId, entityId, entityName, typ, category string) (*Config, error) {
	return s.entity.First(
		`user_id = ? AND client_id = ? AND workspace_id = ? AND entity_id = ? AND entity_name = ? AND type = ? AND category = ?`,
		userId, clientId, workspaceId, entityId, entityName, typ, category,
	)
}

// FindById returns a config row by primary key.
func (s *ConfigService) FindById(id string) (*Config, error) {
	return s.entity.First(`id = ?`, id)
}

// Upsert creates a new config row or updates the existing one matching the
// full identifier tuple. Mirrors TS updateOrCreate semantics.
func (s *ConfigService) Upsert(userId, clientId, workspaceId, entityId, entityName, typ, category string, value json.RawMessage, status *string) (*Config, error) {
	existing, _ := s.FindByTuple(userId, clientId, workspaceId, entityId, entityName, typ, category)
	if existing != nil {
		existing.Value = value
		if status != nil {
			existing.Status = status
		}
		if _, err := s.entity.Update(existing, `id = ?`, existing.Id); err != nil {
			return nil, err
		}
		return existing, nil
	}
	row := &Config{
		UserId:      userId,
		ClientId:    clientId,
		WorkspaceId: workspaceId,
		EntityId:    entityId,
		EntityName:  entityName,
		Type:        typ,
		Category:    category,
		Value:       value,
		Status:      status,
	}
	if err := s.entity.Insert(row); err != nil {
		return nil, err
	}
	return row, nil
}

// UpdateById patches an existing row's value/status.
func (s *ConfigService) UpdateById(id string, value json.RawMessage, status *string) (*Config, error) {
	existing, err := s.FindById(id)
	if err != nil || existing == nil {
		return nil, err
	}
	existing.Value = value
	if status != nil {
		existing.Status = status
	}
	if _, err := s.entity.Update(existing, `id = ?`, existing.Id); err != nil {
		return nil, err
	}
	return existing, nil
}
