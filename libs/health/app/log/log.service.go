package log

type LogService struct {
	entity *LogEntity `inject:""`
}

func (s *LogService) Create(serviceId string, state *string) (*Log, error) {
	entry := &Log{ServiceId: serviceId, State: state}
	if err := s.entity.Insert(entry); err != nil {
		return nil, err
	}
	return entry, nil
}
