package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/notification/app"
	notificationmessage "github.com/thescaffold/gox-apps/libs/notification/app/message"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"gorm.io/gorm"
)

func inboxMessage(t *testing.T, db *gorm.DB, user, ws string, at time.Time, mut func(*notificationmessage.Message)) *notificationmessage.Message {
	t.Helper()
	ch, key, subj, ref := "web", "k", "Hello", "ref-"+uuid.NewString()
	m := &notificationmessage.Message{UserId: &user, Channel: &ch, Key: &key, Subject: &subj, Reference: &ref}
	if ws != "" {
		m.WorkspaceId = &ws
	}
	m.Id = uuid.NewString()
	m.CreatedAt = &at
	if mut != nil {
		mut(m)
	}
	if err := db.Create(m).Error; err != nil {
		t.Fatal(err)
	}
	return m
}

// PLAN U9: the signed-in person's inbox — a cursor-paged list, the unread
// count, mark one read, mark all read, and a link per message.
func TestInbox_ListCountAndRead(t *testing.T) {
	svc, db, _ := deliverFixture(t)
	me, other, ws := "u-"+uuid.NewString()[:8], "u-"+uuid.NewString()[:8], "ws-"+uuid.NewString()[:8]
	base := time.Now().UTC().Add(-time.Hour).Truncate(time.Millisecond)

	a := inboxMessage(t, db, me, ws, base, nil)
	b := inboxMessage(t, db, me, "", base.Add(time.Minute), func(m *notificationmessage.Message) {
		link := "yan://post/1"
		m.Link = &link
	})
	c := inboxMessage(t, db, me, ws, base.Add(2*time.Minute), func(m *notificationmessage.Message) {
		m.Data = []byte(`{"link":"https://yansocial.com/n/2"}`)
	})
	// Same instant as c: the cursor breaks the tie by id.
	d := inboxMessage(t, db, me, ws, base.Add(2*time.Minute), nil)
	// Not in my inbox: someone else's, another workspace's, an email copy,
	// not yet published, expired.
	inboxMessage(t, db, other, ws, base, nil)
	inboxMessage(t, db, me, "ws-elsewhere", base, nil)
	inboxMessage(t, db, me, ws, base, func(m *notificationmessage.Message) { e := "email"; m.Channel = &e })
	inboxMessage(t, db, me, ws, base, func(m *notificationmessage.Message) { f := time.Now().UTC().Add(time.Hour); m.PublishAt = &f })
	inboxMessage(t, db, me, ws, base, func(m *notificationmessage.Message) { p := time.Now().UTC().Add(-time.Minute); m.ExpireAt = &p })

	seen := map[string]bool{}
	var order []string
	cursor := ""
	for pages := 0; ; pages++ {
		page, err := svc.Inbox(app.InboxQuery{UserID: me, WorkspaceID: ws, Cursor: cursor, Limit: 3})
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range page.Items {
			if seen[it.ID] {
				t.Fatalf("%s listed twice", it.ID)
			}
			seen[it.ID] = true
			order = append(order, it.ID)
		}
		if page.Next == "" {
			break
		}
		if pages > 3 {
			t.Fatal("the cursor never ends")
		}
		cursor = page.Next
	}
	if len(order) != 4 || order[3] != a.Id || order[2] != b.Id || !(seen[c.Id] && seen[d.Id]) {
		t.Fatalf("inbox order = %v (a=%s b=%s c=%s d=%s)", order, a.Id, b.Id, c.Id, d.Id)
	}

	first, _ := svc.Inbox(app.InboxQuery{UserID: me, WorkspaceID: ws, Limit: 50})
	links := map[string]string{}
	for _, it := range first.Items {
		links[it.ID] = it.Link
		if it.Read || it.ReadAt != nil || it.CreatedAt == nil {
			t.Fatalf("a new message: %+v", it)
		}
	}
	if links[b.Id] != "yan://post/1" || links[c.Id] != "https://yansocial.com/n/2" || links[a.Id] != "" {
		t.Fatalf("links = %v", links)
	}

	if n, err := svc.UnreadCount(me, ws); err != nil || n != 4 {
		t.Fatalf("unread = %d %v, want 4", n, err)
	}
	item, err := svc.MarkRead(me, ws, b.Id)
	if err != nil || !item.Read || item.ReadAt == nil {
		t.Fatalf("mark read: %+v %v", item, err)
	}
	if again, err := svc.MarkRead(me, ws, b.Id); err != nil || !again.ReadAt.Equal(*item.ReadAt) {
		t.Fatalf("marking read again changed it: %+v %v", again, err)
	}
	if n, _ := svc.UnreadCount(me, ws); n != 3 {
		t.Fatalf("unread after one read = %d, want 3", n)
	}
	unread, _ := svc.Inbox(app.InboxQuery{UserID: me, WorkspaceID: ws, UnreadOnly: true})
	if len(unread.Items) != 3 {
		t.Fatalf("unread filter = %d items", len(unread.Items))
	}
	// Someone else's message is not mine to read.
	theirs := inboxMessage(t, db, other, ws, base, nil)
	if _, err := svc.MarkRead(me, ws, theirs.Id); !errors.Is(err, app.ErrInboxNotFound) {
		t.Fatalf("marking another person's message: %v", err)
	}
	var stored notificationmessage.Message
	if err := db.First(&stored, "id = ?", theirs.Id).Error; err != nil {
		t.Fatal(err)
	}
	if stored.ReadAt != nil {
		t.Fatal("another person's message was marked read")
	}

	if n, err := svc.MarkAllRead(me, ws); err != nil || n != 3 {
		t.Fatalf("mark all read = %d %v, want 3", n, err)
	}
	if n, _ := svc.UnreadCount(me, ws); n != 0 {
		t.Fatalf("unread after mark all = %d", n)
	}
	if n, _ := svc.UnreadCount(other, ws); n != 2 {
		t.Fatalf("mark all read touched someone else's inbox: %d unread left, want 2", n)
	}

	if _, err := svc.Inbox(app.InboxQuery{UserID: me, Cursor: "not-a-cursor"}); !errors.Is(err, app.ErrInboxBadCursor) {
		t.Fatalf("bad cursor: %v", err)
	}
	if _, err := svc.Inbox(app.InboxQuery{}); !errors.Is(err, app.ErrInboxNoUser) {
		t.Fatalf("no user: %v", err)
	}
}

// A delivered notification lands in the inbox with its rendered web text as
// the body and data.link (or a top-level link) as its link.
func TestInbox_DeliveredMessageHasBodyAndLink(t *testing.T) {
	svc, db, _ := deliverFixture(t)
	key := "k-" + uuid.NewString()[:8]
	seedTemplate(t, db, key)
	p := payload(key)
	if err := svc.Deliver(p); err != nil {
		t.Fatal(err)
	}
	page, err := svc.Inbox(app.InboxQuery{UserID: p["userId"].(string), WorkspaceID: "ws-1"})
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("inbox = %+v %v", page, err)
	}
	it := page.Items[0]
	if it.Link != "https://app.example.test/a/1" || it.Body != "Deploy &lt;b&gt;v12&lt;/b&gt;" || it.Subject != "Approval needed" {
		t.Fatalf("delivered item = %+v", it)
	}

	// A top-level link is carried into data.link.
	p2 := payload(key)
	delete(p2["data"].(map[string]any), "link")
	p2["link"] = "yan://notifications/9"
	if err := svc.Deliver(p2); err != nil {
		t.Fatal(err)
	}
	page, _ = svc.Inbox(app.InboxQuery{UserID: p2["userId"].(string), WorkspaceID: "ws-1"})
	if len(page.Items) != 1 || page.Items[0].Link != "yan://notifications/9" {
		t.Fatalf("top-level link: %+v", page.Items)
	}
}

// A caller with no workspace sees and counts only the messages that belong to none (an empty
// workspace counted every workspace's), and a read is announced as apps.notification.inbox.read.
func TestInbox_NoWorkspaceAndTheReadEvent(t *testing.T) {
	svc, db, _ := deliverFixture(t)
	me := "u-" + uuid.NewString()[:8]
	base := time.Now().UTC().Add(-time.Hour)
	none := inboxMessage(t, db, me, "", base, nil)
	inboxMessage(t, db, me, "ws-a-"+uuid.NewString()[:6], base, nil)
	inboxMessage(t, db, me, "ws-b-"+uuid.NewString()[:6], base, nil)

	if n, err := svc.UnreadCount(me, ""); err != nil || n != 1 {
		t.Fatalf("unread with no workspace = %d %v, want 1 (only the workspace-less message)", n, err)
	}
	page, _ := svc.Inbox(app.InboxQuery{UserID: me})
	if len(page.Items) != 1 || page.Items[0].ID != none.Id {
		t.Fatalf("inbox with no workspace listed %d items", len(page.Items))
	}

	got := make(chan map[string]any, 4)
	events.Default().Subscribe(app.EventInboxRead, func(_ string, p any) {
		if m, ok := p.(map[string]any); ok && m["userId"] == me {
			got <- m
		}
	})
	if _, err := svc.MarkRead(me, "", none.Id); err != nil {
		t.Fatal(err)
	}
	select {
	case m := <-got:
		if ids, _ := m["ids"].([]string); len(ids) != 1 || ids[0] != none.Id {
			t.Fatalf("read event = %v", m)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no apps.notification.inbox.read after MarkRead")
	}
	if _, err := svc.MarkAllRead(me, ""); err != nil {
		t.Fatal(err)
	}
	select {
	case m := <-got:
		if m["updated"] != int64(0) {
			t.Fatalf("read-all event = %v (nothing was left unread)", m)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no apps.notification.inbox.read after MarkAllRead")
	}
}
