package service

type ServiceService struct {
	entity *ServiceEntity `inject:""`
}

func (s *ServiceService) FindById(id string) (*Service, error) {
	return s.entity.First(id)
}

func (s *ServiceService) FindAll() ([]Service, error) {
	return s.entity.Find(0, 0, "")
}
