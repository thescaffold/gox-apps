package file

type FileService struct {
	entity *FileEntity `inject:""`
}

func (s *FileService) FindById(id string) (*File, error) {
	return s.entity.First(id)
}

func (s *FileService) Save(f *File) error {
	return s.entity.Insert(f)
}
