package app

import (
	"errors"

	"github.com/awesome-goose/goose/types"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// The caller of an inbox route comes only from the verified token
// (coreauth.Session + coreauth.FromClaims on the route), never from x-ntx-*
// headers or the body.
func inboxCaller(ctx ntxctx.NTXContext) (userID, workspaceID string) {
	return ctx.UserID, ctx.WorkspaceID
}

func (c *AppController) inboxTitle(ctx ntxctx.NTXContext) string {
	if c.lang == nil {
		return "Notification"
	}
	return c.lang.Translate("apps.notification.app.title", nil, ctx.Preference)
}

func (c *AppController) inboxError(title string, err error) types.Output {
	switch {
	case errors.Is(err, ErrInboxNoUser):
		return response.Unauthorized(title, "Sign in to see your notifications.")
	case errors.Is(err, ErrInboxNotFound):
		return response.NotFound(title, "That notification isn't in your inbox.")
	case errors.Is(err, ErrInboxBadCursor):
		return response.BadRequest(title, "The cursor is not valid. Start again from the first page.")
	}
	return response.InternalServerError(title, "Your notifications could not be loaded.")
}

// Inbox is GET /inbox?cursor=&limit=&unread=: a page of the caller's in-app
// messages, newest first. data: {items: [InboxItem], next: "<cursor>"|""}.
func (c *AppController) Inbox(dto *InboxDto) types.Output {
	title := c.inboxTitle(dto.Ctx)
	userID, workspaceID := inboxCaller(dto.Ctx)
	page, err := c.appService.Inbox(InboxQuery{
		UserID: userID, WorkspaceID: workspaceID,
		Cursor: dto.Cursor, Limit: dto.Limit, UnreadOnly: dto.Unread,
	})
	if err != nil {
		return c.inboxError(title, err)
	}
	return response.Success(page, title, "", nil)
}

// InboxUnreadCount is GET /inbox/unread-count. data: {count}.
func (c *AppController) InboxUnreadCount(dto *InboxCountDto) types.Output {
	title := c.inboxTitle(dto.Ctx)
	userID, workspaceID := inboxCaller(dto.Ctx)
	n, err := c.appService.UnreadCount(userID, workspaceID)
	if err != nil {
		return c.inboxError(title, err)
	}
	return response.Success(InboxCount{Count: n}, title, "", nil)
}

// InboxMarkRead is POST /inbox/:id/read: marks one message read (again is
// fine). data: the InboxItem. 404 for a message outside the caller's inbox.
func (c *AppController) InboxMarkRead(dto *InboxReadDto) types.Output {
	title := c.inboxTitle(dto.Ctx)
	userID, workspaceID := inboxCaller(dto.Ctx)
	item, err := c.appService.MarkRead(userID, workspaceID, dto.ID)
	if err != nil {
		return c.inboxError(title, err)
	}
	return response.Success(item, title, "", nil)
}

// InboxMarkAllRead is POST /inbox/read: marks every unread message read.
// data: {updated}.
func (c *AppController) InboxMarkAllRead(dto *InboxMarkAllDto) types.Output {
	title := c.inboxTitle(dto.Ctx)
	userID, workspaceID := inboxCaller(dto.Ctx)
	n, err := c.appService.MarkAllRead(userID, workspaceID)
	if err != nil {
		return c.inboxError(title, err)
	}
	return response.Success(InboxUpdated{Updated: n}, title, "", nil)
}
