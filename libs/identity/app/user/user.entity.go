package user

import "github.com/awesome-goose/goose/modules/sql"

type User struct {
	sql.BaseEntity
	Name   string  `gorm:"column:name;type:varchar(255);not null"  json:"name"`
	Email  string  `gorm:"column:email;type:varchar(255);not null"  json:"email"`
	Phone  *string `gorm:"column:phone;type:varchar(255)"           json:"phone,omitempty"`
	Secret *string `gorm:"column:secret;type:text"                  json:"-"`
	Status *string `gorm:"column:status;type:varchar(255)"          json:"status,omitempty"`
}
func (User) TableName() string { return "IdentityUsers" }
type UserEntity struct{ *sql.Entity[User] `inject:""` }
func (e *UserEntity) OnRegister() {
	e.Hydrate("IdentityUsers", []string{"email", "phone"}, nil, nil, nil, nil, nil, "created_at desc")
}
