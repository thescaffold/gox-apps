package summary

type SummaryService struct {
	entity *SummaryEntity `inject:""`
}

func (s *SummaryService) FindByServiceId(serviceId string) ([]Summary, error) {
	return s.entity.Find(0, 0, "service_id = ?", serviceId)
}
