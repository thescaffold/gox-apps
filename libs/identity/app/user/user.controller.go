package user

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// UserController mirrors ntx-apps/libs/identity/src/api/user/user.controller.ts.
type UserController struct {
	crud.CrudResource[User, CreateUserDto, UpdateUserDto]

	entity *UserEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *UserController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[User, CreateUserDto, UpdateUserDto]{
		// Mirrors TS user.controller.ts name = 'user'.
		Name: "user",
		// Mirrors TS user.controller.ts searchable = ['ref'].
		// gox User entity has email/phone/name columns instead of `ref` (which
		// in TS holds the primary contact handle, typically email). TS list
		// kept verbatim for API surface parity.
		Searchable: []string{"ref"},
		// Mirrors TS user.controller.ts unique = ({ref}) => [{ref}]. Phase H
		// flips the lookup to the new `ref` column (filled by AuthService at
		// registration time); the legacy email-keyed unique is kept as a
		// secondary constraint so back-compat callers don't break.
		Unique: func(d *CreateUserDto) []map[string]any {
			out := []map[string]any{{"email": d.Email}}
			if d.Ref != nil && *d.Ref != "" {
				out = append(out, map[string]any{"ref": *d.Ref})
			}
			return out
		},
	})
	c.SetLang(c.lang)
}
