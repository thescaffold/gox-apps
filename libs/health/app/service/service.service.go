package service

type ServiceService struct {
	entity *ServiceEntity `inject:""`
}

func (s *ServiceService) FindById(id string) (*Service, error) {
	return s.entity.First(id)
}

func (s *ServiceService) FindByName(name string) (*Service, error) {
	return s.entity.First(`name = ?`, name)
}

func (s *ServiceService) FindAll() ([]Service, error) {
	return s.entity.Find(0, 0, "")
}

// UpdateState updates the state column for the row id.
func (s *ServiceService) UpdateState(id string, state string) error {
	row := &Service{State: &state}
	_, err := s.entity.Update(row, `id = ?`, id)
	return err
}

// UpsertByName is the upsert primitive used by POST /register: finds an
// existing row by name or inserts a new one, then updates the optional fields.
// Returns the persisted Service. Mirrors TS serviceService.updateOrCreate.
func (s *ServiceService) UpsertByName(name string, desc, status *string) (*Service, error) {
	existing, _ := s.entity.First(`name = ?`, name)
	if existing != nil {
		if desc != nil {
			existing.Desc = desc
		}
		if status != nil {
			existing.Status = status
		}
		if _, err := s.entity.Update(existing, `id = ?`, existing.Id); err != nil {
			return nil, err
		}
		return existing, nil
	}
	svc := &Service{Name: name, Desc: desc, Status: status}
	if err := s.entity.Insert(svc); err != nil {
		return nil, err
	}
	return svc, nil
}
