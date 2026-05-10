package file

// FileService is the per-file CRUD service. Mirrors TS FileService which
// extends BaseRepository<File> with permission-scope filters and adds no
// custom methods of its own. Upload/download responsibilities belong to the
// gox-packages-blobs FilesService HTTP client; controllers inject it directly.
type FileService struct {
	entity *FileEntity `inject:""`
}

// FindById is a convenience over the underlying entity.
func (s *FileService) FindById(id string) (*File, error) {
	return s.entity.First(id)
}
