package log

type LogService struct {
	entity *LogEntity `inject:""`
}

func (s *LogService) Create(entry *Log) error {
	return s.entity.Insert(entry)
}

func (s *LogService) Activities(page, perPage int) ([]Log, int64, error) {
	query := "desc IS NOT NULL"
	logs, err := s.entity.Page(page, perPage, query)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.entity.Count(query)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
