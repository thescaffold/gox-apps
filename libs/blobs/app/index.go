package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "blobs"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {"guest:read:apps:blobs:remote:*:self", "guest:create:apps:blobs:file:*:self", "guest:update:apps:blobs:file:*:self", "guest:read:apps:blobs:file:*:self"},
	"member":        {"member:create:apps:blobs:*:*:workspace", "member:update:apps:blobs:*:*:self", "member:read:apps:blobs:*:*:workspace", "member:delete:apps:blobs:*:*:self"},
	"admin":         {"admin:*:apps:blobs:*:*:workspace"},
	"global-member": {"global-member:*:apps:blobs:*:*:global"},
	"global-admin":  {"global-admin:*:apps:blobs:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{}
