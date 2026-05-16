package app

import (
	formsform "github.com/thescaffold/gox-apps/libs/forms/app/form"
	formsformfield "github.com/thescaffold/gox-apps/libs/forms/app/formfield"
	formsformlog "github.com/thescaffold/gox-apps/libs/forms/app/formlog"
	formsformtype "github.com/thescaffold/gox-apps/libs/forms/app/formtype"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "forms"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/forms/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest": {
		"guest:create:apps:forms:one:*:self",
		"guest:read:apps:forms:one:*:self",
	},
	"member": {
		"member:create:apps:forms:*:*:workspace",
		"member:update:apps:forms:*:*:self",
		"member:read:apps:forms:*:*:workspace",
		"member:delete:apps:forms:*:*:self",
	},
	"admin":         {"admin:*:apps:forms:*:*:workspace"},
	"global-member": {"global-member:*:apps:forms:*:*:global"},
	"global-admin":  {"global-admin:*:apps:forms:*:*:global"},
}

// Subscriptions mirrors TS subscriptions = []. The retained heartbeat is a Go
// stub kept for tests that assert presence.
var Subscriptions = map[string]events.EventHandler{}

// Top-level exports mirroring ntx-apps/libs/forms/src/index.ts.
var (
	Entities = []any{
		formsform.Form{},
		formsformfield.FormField{},
		formsformlog.FormLog{},
		formsformtype.FormType{},
	}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/forms.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
