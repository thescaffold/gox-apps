package page

type PageService struct {
	entity *PageEntity `inject:""`
}

func (s *PageService) FindByFileId(fileId string) ([]Page, error) {
	return s.entity.Find(0, 0, "file_id = ?", fileId)
}
