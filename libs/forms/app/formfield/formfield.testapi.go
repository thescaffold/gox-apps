package formfield

import (
	"github.com/awesome-goose/goose/modules/sql"
	formpkg "github.com/thescaffold/gox-apps/libs/forms/app/form"
)

// NewControllerForTest builds a FormFieldController over q without DI.
func NewControllerForTest(q *sql.Query) (*FormFieldController, *formpkg.FormEntity) {
	fe := &formpkg.FormEntity{Entity: (&sql.Entity[formpkg.Form]{}).With(q)}
	fe.OnRegister()
	ffe := &FormFieldEntity{Entity: (&sql.Entity[FormField]{}).With(q)}
	ffe.OnRegister()
	c := &FormFieldController{entity: ffe, forms: fe}
	c.OnRegister()
	return c, fe
}
