package log

import "encoding/json"

type LogService struct {
	entity *LogEntity `inject:""`
}

// Create inserts a log entry. state is optional.
func (s *LogService) Create(serviceId string, state *string) (*Log, error) {
	entry := &Log{ServiceId: serviceId, State: state}
	if err := s.entity.Insert(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// CreateWithMeta is the richer variant used by POST /ping — accepts meta JSON.
func (s *LogService) CreateWithMeta(serviceId, state string, meta json.RawMessage) (*Log, error) {
	entry := &Log{ServiceId: serviceId, State: &state, Meta: meta}
	if err := s.entity.Insert(entry); err != nil {
		return nil, err
	}
	return entry, nil
}
