package list

type ListService struct {
	entity *ListEntity `inject:""`
}

func (s *ListService) FilterByKey(key string, parentId *string) ([]List, error) {
	if parentId != nil {
		return s.entity.Some("key = ? AND parent_id = ?", key, *parentId)
	}
	return s.entity.Some("key = ?", key)
}

func (s *ListService) FindByCode(code string, parentId *string) (*List, error) {
	if parentId != nil {
		return s.entity.First("code = ? AND parent_id = ?", code, *parentId)
	}
	return s.entity.First("code = ?", code)
}

func (s *ListService) FindByValue(value string, parentId *string) (*List, error) {
	if parentId != nil {
		return s.entity.First("value = ? AND parent_id = ?", value, *parentId)
	}
	return s.entity.First("value = ?", value)
}
