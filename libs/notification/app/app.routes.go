package app

import (
	"github.com/awesome-goose/goose/modules/router"
	"github.com/awesome-goose/goose/types"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/notification/src/app.controller.ts route paths:
//
//	GET    /                → Health
//	GET    /scope           → FindByScope
//	GET    /priority        → FindByPriority
//	GET    /subscription    → GetSubscription
//	PATCH  /subscription    → UpdateSubscription
//	GET    /capture         → Capture (non-production, token-guarded)
//
// The in-app inbox of the signed-in person (PLAN U9), each behind a verified
// session whose claims are the only source of the caller:
//
//	GET    /inbox              → Inbox (?cursor=&limit=&unread=) {items, next}
//	GET    /inbox/unread-count → InboxUnreadCount {count}
//	POST   /inbox/read         → InboxMarkAllRead {updated}
//	POST   /inbox/:id/read     → InboxMarkRead InboxItem
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/scope", []any{AppController{}, "FindByScope"}),
	router.Get("/priority", []any{AppController{}, "FindByPriority"}),
	router.Get("/subscription", []any{AppController{}, "GetSubscription"}, coreauth.Session()...),
	router.Patch("/subscription", []any{AppController{}, "UpdateSubscription"}, coreauth.Session()...),
	router.Get("/capture", []any{AppController{}, "Capture"}),
	router.Get("/inbox", []any{AppController{}, "Inbox"}, inboxGuard()...),
	router.Get("/inbox/unread-count", []any{AppController{}, "InboxUnreadCount"}, inboxGuard()...),
	router.Post("/inbox/read", []any{AppController{}, "InboxMarkAllRead"}, inboxGuard()...),
	router.Post("/inbox/:id/read", []any{AppController{}, "InboxMarkRead"}, inboxGuard()...),
)

// inboxGuard is a verified session, then the caller taken from its claims.
func inboxGuard() []types.Middleware {
	return append(coreauth.Session(), &coreauth.FromClaims{})
}
