package page

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// PageController mirrors ntx-apps/libs/blobs/src/api/page/page.controller.ts.
type PageController struct {
	crud.CrudResource[Page, CreatePageDto, UpdatePageDto]

	entity *PageEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *PageController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Page, CreatePageDto, UpdatePageDto]{
		// Mirrors TS page.controller.ts:18 name = 'page'.
		Name: "page",
		// A page has no tenant column: it is its file's. Reads are confined to pages of the
		// caller's workspace's files; writes need a platform admin (uploads go through
		// /upload/batch, which checks the file) (U26).
		WorkspaceScopeClause: `file_id IN (SELECT id FROM "BlobsFiles" WHERE workspace_id = ?)`,
		// Mirrors TS page.controller.ts:19 searchable = [].
		Searchable: nil,
		// TS unique = ({}: Page) => [] — never blocks creation.
	})
	c.SetLang(c.lang)
}
