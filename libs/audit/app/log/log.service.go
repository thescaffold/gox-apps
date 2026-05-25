package log

import "time"

type LogService struct {
	entity *LogEntity `inject:""`
}

func (s *LogService) Create(entry *Log) error {
	return s.entity.Insert(entry)
}

// LogsInRange returns every log whose created_at falls within [from, to],
// newest-first (the entity's default created_at desc sort). Used by the periodic
// summary aggregation.
func (s *LogService) LogsInRange(from, to time.Time) ([]Log, error) {
	return s.entity.Find(0, 0, `"created_at" >= ? AND "created_at" <= ?`, from, to)
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
