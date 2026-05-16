package rule

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// RuleController mirrors ntx-apps/libs/notification/src/api/rule/rule.controller.ts.
type RuleController struct {
	crud.CrudResource[Rule, CreateRuleDto, UpdateRuleDto]

	entity *RuleEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *RuleController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Rule, CreateRuleDto, UpdateRuleDto]{
		// Mirrors TS rule.controller.ts:18 name = 'rule'.
		Name: "rule",
		// Mirrors TS rule.controller.ts:19 searchable = ['group','key'].
		Searchable: []string{"group", "key"},
		// Mirrors TS rule.controller.ts:21 unique = ({group,key}) => [{group,key}].
		Unique: func(d *CreateRuleDto) []map[string]any {
			return []map[string]any{{"group": d.Group, "key": d.Key}}
		},
	})
	c.SetLang(c.lang)
}
