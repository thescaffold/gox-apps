package file

import (
	"io"

	"github.com/thescaffold/gox-packages-blobs/files"
)

type FileService struct {
	entity      *FileEntity      `inject:""`
	blobsFiles  *files.FilesService `inject:""`
}

func (s *FileService) FindById(id string) (*File, error) {
	return s.entity.First(id)
}

func (s *FileService) Upload(r io.Reader, name, bucket string, tags []string) (string, error) {
	return s.blobsFiles.UploadReader(r, name, bucket, tags)
}

func (s *FileService) Download(id string) (io.ReadCloser, error) {
	return s.blobsFiles.Download(id)
}
