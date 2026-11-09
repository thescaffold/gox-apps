package app

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/thescaffold/gox-apps/libs/notification/app/message"
)

// The in-app inbox: the signed-in person's own `web` messages (what the
// provider's web channel saves), newest first, with an unread count and
// mark-read. Every query is confined to the caller: a message belongs to the
// inbox of its user_id, in the caller's workspace or in none.
//
// Pagination is by cursor (TRD §7.6: ?cursor=&limit= → {items, next}). The
// cursor is opaque to clients; it encodes the last item's created_at and id.

const (
	inboxChannel      = "web"
	inboxDefaultLimit = 20
	inboxMaxLimit     = 100
)

// ErrInboxNoUser is returned when the caller is not known.
var ErrInboxNoUser = errors.New("notification: inbox needs a signed-in user")

// ErrInboxNotFound is returned for a message that is not in the caller's inbox.
var ErrInboxNotFound = errors.New("notification: no such message in this inbox")

// ErrInboxBadCursor is returned for a cursor this server did not issue.
var ErrInboxBadCursor = errors.New("notification: invalid cursor")

// InboxItem is one row of the inbox as clients see it.
type InboxItem struct {
	ID        string          `json:"id"`
	Key       string          `json:"key"`
	Type      string          `json:"type,omitempty"`
	Subject   string          `json:"subject"`
	Body      string          `json:"body,omitempty"`
	Link      string          `json:"link,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Priority  string          `json:"priority,omitempty"`
	Theme     string          `json:"theme,omitempty"`
	Read      bool            `json:"read"`
	ReadAt    *time.Time      `json:"readAt"`
	CreatedAt *time.Time      `json:"createdAt"`
}

// InboxPage is one page of the inbox; Next is "" on the last page.
type InboxPage struct {
	Items []InboxItem `json:"items"`
	Next  string      `json:"next"`
}

// InboxQuery selects a page of the inbox.
type InboxQuery struct {
	UserID      string
	WorkspaceID string
	Cursor      string
	Limit       int
	UnreadOnly  bool
}

// inboxScope is the WHERE clause every inbox query starts from: the caller's
// own, published, unexpired web messages, in their workspace or in none.
func inboxScope(userID, workspaceID string, now time.Time) (string, []any) {
	parts := []string{
		`"user_id" = ?`,
		`"channel" = ?`,
		`("publish_at" IS NULL OR "publish_at" <= ?)`,
		`("expire_at" IS NULL OR "expire_at" > ?)`,
	}
	args := []any{userID, inboxChannel, now, now}
	if workspaceID != "" {
		parts = append(parts, `("workspace_id" = ? OR "workspace_id" IS NULL OR "workspace_id" = '')`)
		args = append(args, workspaceID)
	}
	return strings.Join(parts, " AND "), args
}

type inboxCursor struct {
	At time.Time `json:"t"`
	ID string    `json:"i"`
}

func encodeCursor(m *message.Message) string {
	if m == nil || m.CreatedAt == nil {
		return ""
	}
	b, _ := json.Marshal(inboxCursor{At: m.CreatedAt.UTC(), ID: m.Id})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(s string) (*inboxCursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, ErrInboxBadCursor
	}
	var c inboxCursor
	if err := json.Unmarshal(b, &c); err != nil || c.ID == "" || c.At.IsZero() {
		return nil, ErrInboxBadCursor
	}
	return &c, nil
}

// Inbox returns a page of the caller's inbox, newest first.
func (s *AppService) Inbox(q InboxQuery) (*InboxPage, error) {
	if q.UserID == "" {
		return nil, ErrInboxNoUser
	}
	if s.messageEntity == nil {
		return nil, errors.New("notification: not configured")
	}
	limit := q.Limit
	if limit <= 0 {
		limit = inboxDefaultLimit
	}
	if limit > inboxMaxLimit {
		limit = inboxMaxLimit
	}
	where, args := inboxScope(q.UserID, q.WorkspaceID, time.Now().UTC())
	if q.UnreadOnly {
		where += ` AND "read_at" IS NULL`
	}
	if q.Cursor != "" {
		c, err := decodeCursor(q.Cursor)
		if err != nil {
			return nil, err
		}
		where += ` AND ("created_at" < ? OR ("created_at" = ? AND "id" < ?))`
		args = append(args, c.At, c.At, c.ID)
	}
	rows, err := s.messageEntity.Find(0, limit+1, where, args...)
	if err != nil {
		return nil, err
	}
	page := &InboxPage{Items: make([]InboxItem, 0, len(rows))}
	if len(rows) > limit {
		rows = rows[:limit]
		page.Next = encodeCursor(&rows[len(rows)-1])
	}
	for i := range rows {
		page.Items = append(page.Items, toInboxItem(&rows[i]))
	}
	return page, nil
}

// UnreadCount is how many of the caller's inbox messages are unread.
func (s *AppService) UnreadCount(userID, workspaceID string) (int64, error) {
	if userID == "" {
		return 0, ErrInboxNoUser
	}
	if s.messageEntity == nil {
		return 0, errors.New("notification: not configured")
	}
	where, args := inboxScope(userID, workspaceID, time.Now().UTC())
	return s.messageEntity.Count(where+` AND "read_at" IS NULL`, args...)
}

// MarkRead marks one of the caller's messages read (again is a no-op) and
// returns it. A message outside the caller's inbox is ErrInboxNotFound.
func (s *AppService) MarkRead(userID, workspaceID, id string) (*InboxItem, error) {
	if userID == "" {
		return nil, ErrInboxNoUser
	}
	if s.messageEntity == nil {
		return nil, errors.New("notification: not configured")
	}
	now := time.Now().UTC()
	where, args := inboxScope(userID, workspaceID, now)
	where = `"id" = ? AND ` + where
	args = append([]any{id}, args...)
	if _, err := s.messageEntity.Update(&message.Message{ReadAt: &now}, where+` AND "read_at" IS NULL`, args...); err != nil {
		return nil, err
	}
	m, err := s.messageEntity.First(where, args...)
	if err != nil || m == nil {
		return nil, ErrInboxNotFound
	}
	item := toInboxItem(m)
	return &item, nil
}

// MarkAllRead marks every unread message in the caller's inbox read and says
// how many it marked.
func (s *AppService) MarkAllRead(userID, workspaceID string) (int64, error) {
	if userID == "" {
		return 0, ErrInboxNoUser
	}
	if s.messageEntity == nil {
		return 0, errors.New("notification: not configured")
	}
	now := time.Now().UTC()
	where, args := inboxScope(userID, workspaceID, now)
	return s.messageEntity.Update(&message.Message{ReadAt: &now}, where+` AND "read_at" IS NULL`, args...)
}

func toInboxItem(m *message.Message) InboxItem {
	item := InboxItem{
		ID:        m.Id,
		Key:       deref(m.Key),
		Type:      deref(m.Type),
		Subject:   deref(m.Subject),
		Link:      deref(m.Link),
		Priority:  deref(m.Priority),
		Theme:     deref(m.Theme),
		Read:      m.ReadAt != nil,
		ReadAt:    m.ReadAt,
		CreatedAt: m.CreatedAt,
	}
	if len(m.Data) > 0 && string(m.Data) != "null" {
		item.Data = m.Data
	}
	// The web channel stores the rendered web template as a JSON string.
	if len(m.Message) > 0 {
		var body string
		if json.Unmarshal(m.Message, &body) == nil {
			item.Body = body
		}
	}
	if item.Link == "" && len(item.Data) > 0 {
		var d map[string]any
		if json.Unmarshal(item.Data, &d) == nil {
			if l, ok := d["link"].(string); ok {
				item.Link = l
			}
		}
	}
	return item
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
