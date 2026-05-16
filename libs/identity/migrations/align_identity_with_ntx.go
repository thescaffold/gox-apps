package migrations

import "github.com/awesome-goose/goose/modules/sql"

// AlignIdentityWithNtx is the Phase 4 schema-alignment migration. It adds
// TS-side columns (User.ref+type, Permission.role_type_id+permission_type_id,
// Provider.key+name+desc+detail+logo_url+site_url+redirect_url+scope+client_id+
// client_secret, Device.os+agent+engine+cpu, DeviceLog.session_id+ip+country+
// region+city+area) alongside the existing gox columns. Shadow-column strategy:
// existing gox columns (User.email/phone/name/secret, Permission.role_id+key,
// Provider.type+reference+email+token, Device.fingerprint+type+platform,
// DeviceLog.device_id+event+meta) stay in place; a follow-up cleanup migration
// drops them once all callers have cut over.
//
// All statements use IF NOT EXISTS so the migration is idempotent and safe to
// rerun against partially-aligned databases.
type AlignIdentityWithNtx struct{ sql.BaseMigration }

func (m *AlignIdentityWithNtx) Run(q *sql.Query) error {
	for _, stmt := range []string{
		// User: TS uses `ref` (canonical handle: email/phone/etc) + `type`
		// (the UserRefType enum). gox already has email/phone/name as the
		// equivalent fields — keep both for now.
		`ALTER TABLE "IdentityUsers" ADD COLUMN IF NOT EXISTS "ref"  varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityUsers" ADD COLUMN IF NOT EXISTS "type" varchar(255) DEFAULT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "idx_identity_users_ref" ON "IdentityUsers" ("ref")`,

		// Permission: TS uses (role_type_id, permission_type_id) — gox has
		// (role_id, key) which the existing app code relied on. Add both new
		// columns nullable so existing rows keep working.
		`ALTER TABLE "IdentityPermissions" ADD COLUMN IF NOT EXISTS "role_type_id"       varchar(36) DEFAULT NULL`,
		`ALTER TABLE "IdentityPermissions" ADD COLUMN IF NOT EXISTS "permission_type_id" varchar(36) DEFAULT NULL`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_permissions_role_type_id"       ON "IdentityPermissions" ("role_type_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_permissions_permission_type_id" ON "IdentityPermissions" ("permission_type_id")`,

		// Provider: TS uses key+name+desc+detail+logo_url+site_url+redirect_url+
		// scope+client_id+client_secret as the registry shape. gox has
		// type+reference+email+token as the per-user-provider shape. The two
		// shapes coexist — TS's Provider is a global OAuth provider registry
		// row; gox's is a per-user link. Both layers persist now; the app
		// code will use whichever fits the call site.
		`ALTER TABLE "IdentityProviders" ADD COLUMN IF NOT EXISTS "key"           varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityProviders" ADD COLUMN IF NOT EXISTS "name"          varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityProviders" ADD COLUMN IF NOT EXISTS "desc"          varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityProviders" ADD COLUMN IF NOT EXISTS "detail"        text         DEFAULT NULL`,
		`ALTER TABLE "IdentityProviders" ADD COLUMN IF NOT EXISTS "logo_url"      varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityProviders" ADD COLUMN IF NOT EXISTS "site_url"      varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityProviders" ADD COLUMN IF NOT EXISTS "redirect_url"  varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityProviders" ADD COLUMN IF NOT EXISTS "scope"         varchar(255) DEFAULT NULL`,
		// NOTE: TS Provider entity calls these `client_id` / `client_secret` but
		// gox already uses `client_id` on this table for the per-user-link's
		// clientId. Storing the OAuth credentials under disambiguated names
		// (`oauth_client_id` / `oauth_client_secret`) so both shapes coexist
		// without overwriting each other.
		`ALTER TABLE "IdentityProviders" ADD COLUMN IF NOT EXISTS "oauth_client_id"     varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityProviders" ADD COLUMN IF NOT EXISTS "oauth_client_secret" varchar(255) DEFAULT NULL`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_providers_key"  ON "IdentityProviders" ("key")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_providers_name" ON "IdentityProviders" ("name")`,

		// Device: TS uses os/agent/engine/cpu — gox has fingerprint/type/
		// platform. The columns are mostly orthogonal (TS describes browser
		// User-Agent components, gox describes device identity); keep both.
		`ALTER TABLE "IdentityDevices" ADD COLUMN IF NOT EXISTS "os"     varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityDevices" ADD COLUMN IF NOT EXISTS "agent"  varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityDevices" ADD COLUMN IF NOT EXISTS "engine" varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityDevices" ADD COLUMN IF NOT EXISTS "cpu"    varchar(255) DEFAULT NULL`,

		// DeviceLog: TS has session_id + ip/country/region/city/area as request
		// telemetry; gox previously only had device_id/event/meta.
		`ALTER TABLE "IdentityDeviceLogs" ADD COLUMN IF NOT EXISTS "session_id" varchar(36)  DEFAULT NULL`,
		`ALTER TABLE "IdentityDeviceLogs" ADD COLUMN IF NOT EXISTS "ip"         varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityDeviceLogs" ADD COLUMN IF NOT EXISTS "country"    varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityDeviceLogs" ADD COLUMN IF NOT EXISTS "region"     varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityDeviceLogs" ADD COLUMN IF NOT EXISTS "city"       varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityDeviceLogs" ADD COLUMN IF NOT EXISTS "area"       varchar(255) DEFAULT NULL`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_device_logs_session_id" ON "IdentityDeviceLogs" ("session_id")`,

		// Role: TS pairs (entity_name, entity_id) with role_type_id — gox's
		// original shape was (name, type, workspace_id). Add shadow columns so
		// attach/detach/sync endpoints can write against the TS-aligned schema.
		`ALTER TABLE "IdentityRoles" ADD COLUMN IF NOT EXISTS "entity_name"  varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityRoles" ADD COLUMN IF NOT EXISTS "entity_id"    varchar(36)  DEFAULT NULL`,
		`ALTER TABLE "IdentityRoles" ADD COLUMN IF NOT EXISTS "role_type_id" varchar(36)  DEFAULT NULL`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_roles_entity"        ON "IdentityRoles" ("entity_name", "entity_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_roles_role_type_id"  ON "IdentityRoles" ("role_type_id")`,

		// Token: TS uses (name, desc, value, check, expired_at) for the API-key
		// shape; gox originally had (type, token, expires_at). Add the TS
		// columns alongside so the /user endpoint can mint 32+16 secrets.
		`ALTER TABLE "IdentityTokens" ADD COLUMN IF NOT EXISTS "name"       varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityTokens" ADD COLUMN IF NOT EXISTS "desc"       text         DEFAULT NULL`,
		`ALTER TABLE "IdentityTokens" ADD COLUMN IF NOT EXISTS "value"      varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityTokens" ADD COLUMN IF NOT EXISTS "check"      varchar(255) DEFAULT NULL`,
		`ALTER TABLE "IdentityTokens" ADD COLUMN IF NOT EXISTS "expired_at" timestamp    DEFAULT NULL`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_tokens_value" ON "IdentityTokens" ("value")`,
	} {
		if _, err := q.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
