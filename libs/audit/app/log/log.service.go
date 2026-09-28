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
	// Two real bugs, both only visible against a real database (PLAN
	// M1-04's integration test is what caught them — this had never been
	// exercised outside a mock before):
	//
	// 1. "desc" is a reserved word in Postgres (ORDER BY ... DESC) —
	//    unquoted, it was parsed as the keyword, not the column, and this
	//    query failed outright every time.
	//
	// 2. Entity[T].Count(query, args...) treats a call with zero args as
	//    "query is an ID, do WHERE id = ?" (its shortcut for Count(someId)).
	//    A bare boolean WHERE clause with no bind params — exactly this one
	//    — falls into that branch by accident: it silently searched for a
	//    row whose id equals the literal string "desc" IS NOT NULL, always
	//    finding none. Entity[T].Page/Find don't have this special case, so
	//    only the *count* (not the actual page of rows) was ever wrong —
	//    worse than an outright failure, because it looked like it worked.
	//    Rewritten with a real (no-op) bind parameter so the same query
	//    takes the normal "raw WHERE clause" path in both.
	query := `"desc" IS DISTINCT FROM ?`
	logs, err := s.entity.Page(page, perPage, query, nil)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.entity.Count(query, nil)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
