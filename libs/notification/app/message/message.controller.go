package message

import "github.com/thescaffold/gox-packages/libs/core/crud"

type MessageController struct {
	crud.CrudResource[Message, CreateMessageDto, UpdateMessageDto]
	entity *MessageEntity `inject:""`
}

func (c *MessageController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Message, CreateMessageDto, UpdateMessageDto]{
		Name:       "NotificationMessage",
		Searchable: []string{"reference", "user_id", "client_id", "workspace_id", "key", "subject", "channel"},
	})
}
