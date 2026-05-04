package file

type FileService struct {
	entity *FileEntity `inject:""`
}

func (s *FileService) FindById(id string) (*File, error) {
	return s.entity.First(id)
}

func (s *FileService) FindByName(name string) (*File, error) {
	return s.entity.First("name = ?", name)
}

func (s *FileService) Save(f *File) error {
	return s.entity.Insert(f)
}

func (s *FileService) Update(f *File) (int64, error) {
	return s.entity.Update(f, "id = ?", f.Id)
}
