package message

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type MessageModule struct{}

func (m *MessageModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *MessageModule) Exports() []any { return []any{&MessageService{}, &MessageEntity{}} }

func (m *MessageModule) Declarations() []any {
	return []any{&MessageService{}, &MessageEntity{}}
}
