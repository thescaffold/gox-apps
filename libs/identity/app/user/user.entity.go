package user

import "github.com/awesome-goose/goose/modules/sql"

// User mirrors ntx-apps/libs/identity/src/api/user/entities/user.entity.ts.
// TS uses `ref` (canonical contact handle, typically email) + `type`
// (UserRefType enum: email/phone/etc). gox carries the original
// name/email/phone fields too — they overlap with TS `ref`+`type` and stay
// for now (shadow-column strategy from Phase 4). New code should write to
// `ref` + `type`; legacy callers keep working off the older columns.
type User struct {
	sql.BaseEntity
	Ref    *string `gorm:"column:ref;type:varchar(255)"    json:"ref,omitempty"`
	Type   *string `gorm:"column:type;type:varchar(255)"   json:"type,omitempty"`
	Name   string  `gorm:"column:name;type:varchar(255);not null"  json:"name"`
	Email  string  `gorm:"column:email;type:varchar(255);not null"  json:"email"`
	Phone  *string `gorm:"column:phone;type:varchar(255)"           json:"phone,omitempty"`
	Secret *string `gorm:"column:secret;type:text"                  json:"-"`
	Status *string `gorm:"column:status;type:varchar(255)"          json:"status,omitempty"`
}

func (User) TableName() string { return "IdentityUsers" }

type UserEntity struct {
	*sql.Entity[User] `inject:""`
}

func (e *UserEntity) OnRegister() {
	// Searchable list now matches TS user.controller.ts:20 — `ref` is the
	// primary handle going forward; email/phone remain queryable for
	// back-compat with pre-shadow-column callers.
	e.Hydrate("IdentityUsers", []string{"ref", "email", "phone"}, nil, nil, nil, nil, nil, "created_at desc")
}
