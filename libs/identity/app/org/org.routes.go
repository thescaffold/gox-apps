package org

import (
	"github.com/awesome-goose/goose/modules/router"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
)

// Every route needs a verified user token, and the workspace comes from its
// claims, never from the request. Reading teams and people is open to any member
// of the workspace; changing them, and all of roles, needs an owner or admin.
//
//	GET    /org/members                              people with their teams and roles
//	GET    /org/catalog                              apps and actions a role can grant (admin)
//	GET    /org/teams                                teams, groups and departments
//	POST   /org/teams                                create (admin)
//	PUT    /org/teams/:id                            rename, re-describe, move (admin)
//	DELETE /org/teams/:id                            delete (admin)
//	GET    /org/teams/:id/members                    who is in it
//	POST   /org/teams/:id/members                    add or change a person (admin)
//	DELETE /org/teams/:id/members/:userId            remove (admin)
//	GET    /org/roles                                custom roles (admin)
//	POST   /org/roles                                define (admin)
//	PUT    /org/roles/:id                            change (admin)
//	DELETE /org/roles/:id                            delete (admin)
//	GET    /org/roles/:id/assignments                who holds it (admin)
//	POST   /org/roles/:id/assignments                give to a person or team (admin)
//	DELETE /org/roles/:id/assignments/:subjectType/:subjectId   take it away (admin)
var ROUTES = router.ForRoutes(
	router.Get("/org/members", []any{OrgController{}, "Members"}, auth.RequireUser()...),
	router.Get("/org/catalog", []any{OrgController{}, "Catalog"}, auth.RequireUser()...),
	router.Get("/org/teams", []any{OrgController{}, "ListTeams"}, auth.RequireUser()...),
	router.Post("/org/teams", []any{OrgController{}, "CreateTeam"}, auth.RequireUser()...),
	router.Put("/org/teams/:id", []any{OrgController{}, "UpdateTeam"}, auth.RequireUser()...),
	router.Delete("/org/teams/:id", []any{OrgController{}, "DeleteTeam"}, auth.RequireUser()...),
	router.Get("/org/teams/:id/members", []any{OrgController{}, "TeamMembers"}, auth.RequireUser()...),
	router.Post("/org/teams/:id/members", []any{OrgController{}, "AddTeamMember"}, auth.RequireUser()...),
	router.Delete("/org/teams/:id/members/:userId", []any{OrgController{}, "RemoveTeamMember"}, auth.RequireUser()...),
	router.Get("/org/roles", []any{OrgController{}, "ListRoles"}, auth.RequireUser()...),
	router.Post("/org/roles", []any{OrgController{}, "CreateRole"}, auth.RequireUser()...),
	router.Put("/org/roles/:id", []any{OrgController{}, "UpdateRole"}, auth.RequireUser()...),
	router.Delete("/org/roles/:id", []any{OrgController{}, "DeleteRole"}, auth.RequireUser()...),
	router.Get("/org/roles/:id/assignments", []any{OrgController{}, "Assignments"}, auth.RequireUser()...),
	router.Post("/org/roles/:id/assignments", []any{OrgController{}, "Assign"}, auth.RequireUser()...),
	router.Delete("/org/roles/:id/assignments/:subjectType/:subjectId", []any{OrgController{}, "Unassign"}, auth.RequireUser()...),
)
