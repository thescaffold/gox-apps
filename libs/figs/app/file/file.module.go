package file

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type FileModule struct{}

func (m *FileModule) Imports() []types.Module {
	return []types.Module{
		ROUTES,
		sql.Child(&sql.Config{}),
	}
}

func (m *FileModule) Exports() []any {
	return []any{&FileService{}, &FileEntity{}}
}

func (m *FileModule) Declarations() []any {
	return []any{
		&FileController{},
		&FileService{},
		&FileEntity{},
	}
}
