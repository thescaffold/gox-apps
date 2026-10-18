// Package org is a workspace's organisation: teams (a team, a group or a
// department, optionally nested), who is in them, company-defined roles holding
// permission strings, and who or which team holds each role.
//
// Everything is scoped to one workspace: every method takes the workspace id and
// refuses ids that belong to another, so a caller holding a team id from another
// workspace changes nothing and learns nothing. A person can only be added to a
// team (or given a role) if they hold a seat in the workspace.
//
// A custom role grants permission strings (see ValidatePermission). A person's
// effective custom-role permissions are those of roles given to them directly,
// to any team they are in, or to an ancestor of such a team (a role on a
// department reaches the teams beneath it).
package org

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Team kinds.
const (
	KindTeam       = "team"
	KindGroup      = "group"
	KindDepartment = "department"
)

// Team member roles.
const (
	TeamLead   = "lead"
	TeamMember = "member"
)

// Who holds a custom role.
const (
	SubjectUser = "user"
	SubjectTeam = "team"
)

// Errors a caller can match.
var (
	ErrNotFound   = errors.New("not found")
	ErrInvalid    = errors.New("invalid")
	ErrNotMember  = errors.New("that person is not in this workspace")
	ErrDuplicate  = errors.New("that name is already used in this workspace")
	ErrHasChild   = errors.New("this team has teams beneath it; move or delete them first")
	ErrCycle      = errors.New("a team cannot sit beneath itself")
	ErrPermission = errors.New("invalid permission")
)

type actorKey struct{}

// WithActor records who is making the changes in ctx, so they are written to the
// audit trail. Calls without an actor change things and leave no event.
func WithActor(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, actorKey{}, userID)
}

func actorOf(ctx context.Context) string { s, _ := ctx.Value(actorKey{}).(string); return s }

// Notifier tells a person something happened to their place in the organisation
// (added to a team, given a role). The host connects it to its notification centre.
type Notifier func(ctx context.Context, workspaceID, userID, title, body, link string)

var notifier struct {
	mu sync.RWMutex
	fn Notifier
}

// SetNotifier installs the host's notifier; with none, nobody is notified.
func SetNotifier(fn Notifier) {
	notifier.mu.Lock()
	notifier.fn = fn
	notifier.mu.Unlock()
}

func (s *Service) tell(ctx context.Context, ws, user, title, body string) {
	notifier.mu.RLock()
	fn := notifier.fn
	notifier.mu.RUnlock()
	if fn != nil && user != "" && user != actorOf(ctx) { // you are not told about what you did yourself
		fn(context.WithoutCancel(ctx), ws, user, title, body, "/personal")
	}
}

// Event is one entry of the audit trail.
type Event struct {
	ID          string          `json:"id"`
	ActorID     string          `json:"actorId" gorm:"column:actor_id"`
	ActorName   string          `json:"actorName" gorm:"column:actor_name"`
	Action      string          `json:"action"`
	SubjectType string          `json:"subjectType" gorm:"column:subject_type"`
	SubjectID   string          `json:"subjectId" gorm:"column:subject_id"`
	Detail      json.RawMessage `json:"detail"`
	CreatedAt   time.Time       `json:"createdAt" gorm:"column:created_at"`
}

// log appends to the audit trail when an actor is known. It never fails the change.
func (s *Service) log(ctx context.Context, ws, action, subjectType, subjectID string, detail map[string]any) {
	actor := actorOf(ctx)
	if actor == "" {
		return
	}
	raw, _ := json.Marshal(detail)
	_ = s.DB.WithContext(ctx).Exec(`INSERT INTO "IdentityOrgEvents" (id, workspace_id, actor_id, action, subject_type, subject_id, detail) VALUES (?, ?, ?, ?, ?, ?, ?::jsonb)`,
		uuid.NewString(), ws, actor, action, subjectType, subjectID, string(raw)).Error
}

// Activity returns the workspace's most recent organisation changes, newest first.
func (s *Service) Activity(ctx context.Context, ws string, limit int) ([]Event, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	out := []Event{}
	err := s.DB.WithContext(ctx).Raw(`SELECT e.id, e.actor_id, COALESCE(u.name, '') AS actor_name, e.action, e.subject_type, e.subject_id, e.detail, e.created_at
		FROM "IdentityOrgEvents" e LEFT JOIN "IdentityUsers" u ON u.id = e.actor_id
		WHERE e.workspace_id = ? ORDER BY e.created_at DESC, e.id DESC LIMIT ?`, ws, limit).Scan(&out).Error
	return out, err
}

// Service is the organisation store.
type Service struct{ DB *gorm.DB }

// New returns a Service over db.
func New(db *gorm.DB) *Service { return &Service{DB: db} }

// --- teams ----------------------------------------------------------------

// Team is a team, group or department.
type Team struct {
	ID          string    `json:"id" gorm:"column:id;primaryKey"`
	WorkspaceID string    `json:"workspaceId" gorm:"column:workspace_id"`
	ParentID    *string   `json:"parentId" gorm:"column:parent_id"`
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	// Members is filled by ListTeams.
	Members int `json:"members" gorm:"->;column:members"`
}

func (Team) TableName() string { return "IdentityTeams" }

func cleanName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if n := len([]rune(s)); n < 1 || n > 80 {
		return "", fmt.Errorf("%w: a name is 1-80 characters", ErrInvalid)
	}
	return s, nil
}

func cleanDesc(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len([]rune(s)) > 500 {
		return "", fmt.Errorf("%w: a description is at most 500 characters", ErrInvalid)
	}
	return s, nil
}

func isDuplicate(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "23505") || strings.Contains(strings.ToLower(err.Error()), "duplicate key"))
}

func (s *Service) team(ctx context.Context, ws, id string) (*Team, error) {
	var t Team
	if err := s.DB.WithContext(ctx).Where("id = ? AND workspace_id = ?", id, ws).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

// CreateTeam adds a team under parentID (empty for top level).
func (s *Service) CreateTeam(ctx context.Context, ws, name, kind, description, parentID string) (*Team, error) {
	name, err := cleanName(name)
	if err != nil {
		return nil, err
	}
	if description, err = cleanDesc(description); err != nil {
		return nil, err
	}
	if kind == "" {
		kind = KindTeam
	}
	if kind != KindTeam && kind != KindGroup && kind != KindDepartment {
		return nil, fmt.Errorf("%w: kind must be team, group or department", ErrInvalid)
	}
	t := &Team{ID: uuid.NewString(), WorkspaceID: ws, Name: name, Kind: kind, Description: description, CreatedAt: time.Now().UTC()}
	if parentID != "" {
		if _, err := s.team(ctx, ws, parentID); err != nil {
			return nil, fmt.Errorf("%w: the parent is not in this workspace", ErrInvalid)
		}
		t.ParentID = &parentID
	}
	if err := s.DB.WithContext(ctx).Create(t).Error; err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	s.log(ctx, ws, "team.created", "team", t.ID, map[string]any{"name": t.Name, "kind": t.Kind})
	return t, nil
}

// ListTeams returns the workspace's teams with member counts, by name.
func (s *Service) ListTeams(ctx context.Context, ws string) ([]Team, error) {
	var out []Team
	err := s.DB.WithContext(ctx).Raw(`SELECT t.*, (SELECT count(*) FROM "IdentityTeamMembers" m WHERE m.team_id = t.id) AS members
		FROM "IdentityTeams" t WHERE t.workspace_id = ? ORDER BY lower(t.name), t.id`, ws).Scan(&out).Error
	if out == nil {
		out = []Team{}
	}
	return out, err
}

// UpdateTeam renames, re-describes or moves a team. parentID nil leaves the
// parent alone; a pointer to "" makes it top level.
func (s *Service) UpdateTeam(ctx context.Context, ws, id string, name, description *string, parentID *string) (*Team, error) {
	t, err := s.team(ctx, ws, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		if t.Name, err = cleanName(*name); err != nil {
			return nil, err
		}
	}
	if description != nil {
		if t.Description, err = cleanDesc(*description); err != nil {
			return nil, err
		}
	}
	if parentID != nil {
		if *parentID == "" {
			t.ParentID = nil
		} else {
			if _, err := s.team(ctx, ws, *parentID); err != nil {
				return nil, fmt.Errorf("%w: the parent is not in this workspace", ErrInvalid)
			}
			ancestors, err := s.ancestors(ctx, ws, *parentID)
			if err != nil {
				return nil, err
			}
			if *parentID == id || ancestors[id] {
				return nil, ErrCycle
			}
			t.ParentID = parentID
		}
	}
	if err := s.DB.WithContext(ctx).Model(&Team{}).Where("id = ? AND workspace_id = ?", id, ws).
		Updates(map[string]any{"name": t.Name, "description": t.Description, "parent_id": t.ParentID}).Error; err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	s.log(ctx, ws, "team.updated", "team", id, map[string]any{"name": t.Name})
	return t, nil
}

// DeleteTeam removes a team, its memberships and the roles it held. A team that
// still has teams beneath it is refused.
func (s *Service) DeleteTeam(ctx context.Context, ws, id string) error {
	team, err := s.team(ctx, ws, id)
	if err != nil {
		return err
	}
	var kids int64
	if err := s.DB.WithContext(ctx).Model(&Team{}).Where("workspace_id = ? AND parent_id = ?", ws, id).Count(&kids).Error; err != nil {
		return err
	}
	if kids > 0 {
		return ErrHasChild
	}
	if err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM "IdentityTeamMembers" WHERE team_id = ? AND workspace_id = ?`, id, ws).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM "IdentityRoleAssignments" WHERE workspace_id = ? AND subject_type = 'team' AND subject_id = ?`, ws, id).Error; err != nil {
			return err
		}
		return tx.Exec(`DELETE FROM "IdentityTeams" WHERE id = ? AND workspace_id = ?`, id, ws).Error
	}); err != nil {
		return err
	}
	s.log(ctx, ws, "team.deleted", "team", id, map[string]any{"name": team.Name})
	return nil
}

// ancestors returns the ids of id and every team above it.
func (s *Service) ancestors(ctx context.Context, ws, id string) (map[string]bool, error) {
	var all []Team
	if err := s.DB.WithContext(ctx).Where("workspace_id = ?", ws).Find(&all).Error; err != nil {
		return nil, err
	}
	parent := map[string]string{}
	for _, t := range all {
		if t.ParentID != nil {
			parent[t.ID] = *t.ParentID
		}
	}
	out := map[string]bool{}
	for cur := id; cur != "" && !out[cur]; cur = parent[cur] {
		out[cur] = true
	}
	return out, nil
}

// --- members --------------------------------------------------------------

// Person is a workspace member as the organisation shows them.
type Person struct {
	UserID string `json:"userId" gorm:"column:user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	// Role is the person's seat in the workspace (owner, admin, member…); in a
	// team listing it is their role in that team (lead or member).
	Role string `json:"role"`
}

// IsMember reports whether userID holds a seat in the workspace.
func (s *Service) IsMember(ctx context.Context, ws, userID string) (bool, error) {
	var n int64
	err := s.DB.WithContext(ctx).Table(`"IdentityUserClientWorkspaces"`).
		Where("workspace_id = ? AND user_id = ? AND deleted_at IS NULL", ws, userID).Count(&n).Error
	return n > 0, err
}

// IsAdmin reports whether userID is an owner or admin of the workspace.
func (s *Service) IsAdmin(ctx context.Context, ws, userID string) (bool, error) {
	var n int64
	err := s.DB.WithContext(ctx).Table(`"IdentityUserClientWorkspaces"`).
		Where("workspace_id = ? AND user_id = ? AND deleted_at IS NULL AND lower(status) IN ('owner','admin')", ws, userID).Count(&n).Error
	return n > 0, err
}

// AddMember puts a workspace member in a team (or changes their role in it).
func (s *Service) AddMember(ctx context.Context, ws, teamID, userID, role string) error {
	if role == "" {
		role = TeamMember
	}
	if role != TeamLead && role != TeamMember {
		return fmt.Errorf("%w: role must be lead or member", ErrInvalid)
	}
	if _, err := s.team(ctx, ws, teamID); err != nil {
		return err
	}
	if ok, err := s.IsMember(ctx, ws, userID); err != nil {
		return err
	} else if !ok {
		return ErrNotMember
	}
	var had int64
	s.DB.WithContext(ctx).Table(`"IdentityTeamMembers"`).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&had)
	if err := s.DB.WithContext(ctx).Exec(`INSERT INTO "IdentityTeamMembers" (team_id, user_id, workspace_id, role) VALUES (?, ?, ?, ?)
		ON CONFLICT (team_id, user_id) DO UPDATE SET role = EXCLUDED.role`, teamID, userID, ws, role).Error; err != nil {
		return err
	}
	team, _ := s.team(ctx, ws, teamID)
	name := ""
	if team != nil {
		name = team.Name
	}
	s.log(ctx, ws, "team.member.added", "team", teamID, map[string]any{"userId": userID, "role": role, "team": name})
	if had == 0 {
		s.tell(ctx, ws, userID, "You were added to "+name, "You are now part of the "+name+" "+map[string]string{TeamLead: "team as its lead", TeamMember: "team"}[role]+".")
	}
	return nil
}

// RemoveMember takes a person out of a team; not being in it is not an error.
func (s *Service) RemoveMember(ctx context.Context, ws, teamID, userID string) error {
	if _, err := s.team(ctx, ws, teamID); err != nil {
		return err
	}
	if err := s.DB.WithContext(ctx).Exec(`DELETE FROM "IdentityTeamMembers" WHERE team_id = ? AND user_id = ? AND workspace_id = ?`, teamID, userID, ws).Error; err != nil {
		return err
	}
	s.log(ctx, ws, "team.member.removed", "team", teamID, map[string]any{"userId": userID})
	return nil
}

// TeamMembers lists a team's people, leads first.
func (s *Service) TeamMembers(ctx context.Context, ws, teamID string) ([]Person, error) {
	if _, err := s.team(ctx, ws, teamID); err != nil {
		return nil, err
	}
	out := []Person{}
	err := s.DB.WithContext(ctx).Raw(`SELECT u.id AS user_id, u.name AS name, u.email AS email, m.role AS role
		FROM "IdentityTeamMembers" m JOIN "IdentityUsers" u ON u.id = m.user_id
		WHERE m.team_id = ? AND m.workspace_id = ? AND u.deleted_at IS NULL
		ORDER BY (m.role = 'lead') DESC, lower(u.name), u.id`, teamID, ws).Scan(&out).Error
	return out, err
}

// Member is a workspace member with where they sit in the organisation.
type Member struct {
	Person
	Teams []string `json:"teams"` // team ids
	Roles []string `json:"roles"` // custom role ids held directly
}

// Members lists the workspace's people with their teams and directly-held roles.
func (s *Service) Members(ctx context.Context, ws string) ([]Member, error) {
	var people []Person
	if err := s.DB.WithContext(ctx).Raw(`SELECT u.id AS user_id, u.name AS name, u.email AS email, m.status AS role
		FROM "IdentityUserClientWorkspaces" m JOIN "IdentityUsers" u ON u.id = m.user_id
		WHERE m.workspace_id = ? AND m.deleted_at IS NULL AND u.deleted_at IS NULL
		ORDER BY lower(u.name), u.id`, ws).Scan(&people).Error; err != nil {
		return nil, err
	}
	teams := map[string][]string{}
	var tm []struct{ UserID, TeamID string }
	if err := s.DB.WithContext(ctx).Raw(`SELECT user_id AS user_id, team_id AS team_id FROM "IdentityTeamMembers" WHERE workspace_id = ? ORDER BY team_id`, ws).Scan(&tm).Error; err != nil {
		return nil, err
	}
	for _, r := range tm {
		teams[r.UserID] = append(teams[r.UserID], r.TeamID)
	}
	roles := map[string][]string{}
	var ra []struct{ SubjectID, RoleID string }
	if err := s.DB.WithContext(ctx).Raw(`SELECT subject_id AS subject_id, role_id AS role_id FROM "IdentityRoleAssignments" WHERE workspace_id = ? AND subject_type = 'user' ORDER BY role_id`, ws).Scan(&ra).Error; err != nil {
		return nil, err
	}
	for _, r := range ra {
		roles[r.SubjectID] = append(roles[r.SubjectID], r.RoleID)
	}
	out := make([]Member, len(people))
	for i, p := range people {
		out[i] = Member{Person: p, Teams: orEmpty(teams[p.UserID]), Roles: orEmpty(roles[p.UserID])}
	}
	return out, nil
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// TeamIDsOf returns the teams a person is in within the workspace.
func (s *Service) TeamIDsOf(ctx context.Context, ws, userID string) ([]string, error) {
	var ids []string
	err := s.DB.WithContext(ctx).Raw(`SELECT team_id FROM "IdentityTeamMembers" WHERE workspace_id = ? AND user_id = ? ORDER BY team_id`, ws, userID).Scan(&ids).Error
	return orEmpty(ids), err
}

// --- custom roles ---------------------------------------------------------

// Role is a company-defined role.
type Role struct {
	ID          string    `json:"id" gorm:"column:id;primaryKey"`
	WorkspaceID string    `json:"workspaceId" gorm:"column:workspace_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions" gorm:"-"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	// Holders is filled by ListRoles: people and teams holding the role.
	Holders int `json:"holders" gorm:"->;column:holders"`
}

func (Role) TableName() string { return "IdentityCustomRoles" }

// roleRow is the stored shape (permissions as JSON).
type roleRow struct {
	ID          string    `gorm:"column:id;primaryKey"`
	WorkspaceID string    `gorm:"column:workspace_id"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	Permissions []byte    `gorm:"column:permissions"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	Holders     int       `gorm:"->;column:holders"`
}

func (r roleRow) role() Role {
	var perms []string
	_ = json.Unmarshal(r.Permissions, &perms)
	if perms == nil {
		perms = []string{}
	}
	return Role{ID: r.ID, WorkspaceID: r.WorkspaceID, Name: r.Name, Description: r.Description, Permissions: perms, CreatedAt: r.CreatedAt, Holders: r.Holders}
}

var (
	segRE   = regexp.MustCompile(`^[a-z0-9*][a-z0-9*._-]*$`)
	actions = regexp.MustCompile(`^(\*|[a-z][a-z-]*)$`)
)

// ValidatePermission checks and normalises one permission string for a custom
// role. It takes "action:apps:<app>:resource:instance:scope" (six parts) or the
// seven-part form with a leading role segment, which is replaced by "custom".
// Only the workspace and self scopes are allowed — never "global", which is not
// a workspace's to grant — and the app must be one of knownApps when that is
// non-empty.
func ValidatePermission(p string, knownApps map[string]bool) (string, error) {
	parts := strings.Split(strings.TrimSpace(p), ":")
	if len(parts) == 7 {
		parts = parts[1:]
	}
	if len(parts) != 6 {
		return "", fmt.Errorf("%w: %q must look like action:apps:<app>:resource:instance:scope", ErrPermission, p)
	}
	action, product, app, resource, instance, scope := parts[0], parts[1], parts[2], parts[3], parts[4], parts[5]
	switch {
	case !actions.MatchString(action):
		return "", fmt.Errorf("%w: %q is not a valid action", ErrPermission, action)
	case product != "apps":
		return "", fmt.Errorf("%w: only apps permissions can be granted", ErrPermission)
	case !segRE.MatchString(app) || app == "*":
		return "", fmt.Errorf("%w: name the app", ErrPermission)
	case len(knownApps) > 0 && !knownApps[app]:
		return "", fmt.Errorf("%w: %q is not an app in this deployment", ErrPermission, app)
	case !segRE.MatchString(resource) || !segRE.MatchString(instance):
		return "", fmt.Errorf("%w: bad resource or instance", ErrPermission)
	case scope != "workspace" && scope != "self":
		return "", fmt.Errorf("%w: the scope must be workspace or self", ErrPermission)
	}
	return "custom:" + strings.Join(parts, ":"), nil
}

// KnownApps returns the apps the deployment has registered permissions for
// (empty when none have been registered, in which case any app name is accepted).
func (s *Service) KnownApps(ctx context.Context) (map[string]bool, error) {
	var keys []string
	if err := s.DB.WithContext(ctx).Raw(`SELECT DISTINCT "key" FROM "IdentityPermissions" WHERE user_id IS NULL AND role_type_id IS NOT NULL AND deleted_at IS NULL`).Scan(&keys).Error; err != nil {
		return nil, err
	}
	apps := map[string]bool{}
	for _, k := range keys {
		if p := strings.Split(k, ":"); len(p) == 7 && p[2] == "apps" && p[6] != "global" {
			apps[p[3]] = true
		}
	}
	return apps, nil
}

// CatalogEntry is one app and the actions its registered permissions use.
type CatalogEntry struct {
	App     string   `json:"app"`
	Actions []string `json:"actions"`
}

// Catalog lists the apps a custom role can grant, with their actions, for an editor.
func (s *Service) Catalog(ctx context.Context) ([]CatalogEntry, error) {
	var keys []string
	if err := s.DB.WithContext(ctx).Raw(`SELECT DISTINCT "key" FROM "IdentityPermissions" WHERE user_id IS NULL AND role_type_id IS NOT NULL AND deleted_at IS NULL`).Scan(&keys).Error; err != nil {
		return nil, err
	}
	by := map[string]map[string]bool{}
	for _, k := range keys {
		p := strings.Split(k, ":")
		if len(p) != 7 || p[2] != "apps" || p[6] == "global" {
			continue
		}
		if by[p[3]] == nil {
			by[p[3]] = map[string]bool{}
		}
		if p[1] != "*" {
			by[p[3]][p[1]] = true
		}
	}
	out := []CatalogEntry{}
	for app, set := range by {
		acts := []string{}
		for a := range set {
			acts = append(acts, a)
		}
		sort.Strings(acts)
		out = append(out, CatalogEntry{App: app, Actions: acts})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].App < out[j].App })
	return out, nil
}

func (s *Service) cleanPerms(ctx context.Context, in []string) ([]string, error) {
	known, err := s.KnownApps(ctx)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := []string{}
	for _, p := range in {
		n, err := ValidatePermission(p, known)
		if err != nil {
			return nil, err
		}
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	if len(out) > 200 {
		return nil, fmt.Errorf("%w: a role holds at most 200 permissions", ErrInvalid)
	}
	sort.Strings(out)
	return out, nil
}

// CreateRole defines a role.
func (s *Service) CreateRole(ctx context.Context, ws, name, description string, permissions []string) (*Role, error) {
	name, err := cleanName(name)
	if err != nil {
		return nil, err
	}
	if description, err = cleanDesc(description); err != nil {
		return nil, err
	}
	perms, err := s.cleanPerms(ctx, permissions)
	if err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(perms)
	row := roleRow{ID: uuid.NewString(), WorkspaceID: ws, Name: name, Description: description, Permissions: raw, CreatedAt: time.Now().UTC()}
	if err := s.DB.WithContext(ctx).Exec(`INSERT INTO "IdentityCustomRoles" (id, workspace_id, name, description, permissions, created_at) VALUES (?, ?, ?, ?, ?::jsonb, ?)`,
		row.ID, ws, name, description, string(raw), row.CreatedAt).Error; err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	r := row.role()
	s.log(ctx, ws, "role.created", "role", r.ID, map[string]any{"name": r.Name, "permissions": r.Permissions})
	return &r, nil
}

func (s *Service) role(ctx context.Context, ws, id string) (*Role, error) {
	var row roleRow
	if err := s.DB.WithContext(ctx).Table(`"IdentityCustomRoles"`).Where("id = ? AND workspace_id = ?", id, ws).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	r := row.role()
	return &r, nil
}

// ListRoles returns the workspace's custom roles with how many hold each.
func (s *Service) ListRoles(ctx context.Context, ws string) ([]Role, error) {
	var rows []roleRow
	if err := s.DB.WithContext(ctx).Raw(`SELECT r.*, (SELECT count(*) FROM "IdentityRoleAssignments" a WHERE a.role_id = r.id) AS holders
		FROM "IdentityCustomRoles" r WHERE r.workspace_id = ? ORDER BY lower(r.name), r.id`, ws).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]Role, len(rows))
	for i, r := range rows {
		out[i] = r.role()
	}
	return out, nil
}

// UpdateRole changes a role's name, description or permissions (nil leaves one alone).
func (s *Service) UpdateRole(ctx context.Context, ws, id string, name, description *string, permissions *[]string) (*Role, error) {
	r, err := s.role(ctx, ws, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		if r.Name, err = cleanName(*name); err != nil {
			return nil, err
		}
	}
	if description != nil {
		if r.Description, err = cleanDesc(*description); err != nil {
			return nil, err
		}
	}
	if permissions != nil {
		if r.Permissions, err = s.cleanPerms(ctx, *permissions); err != nil {
			return nil, err
		}
	}
	raw, _ := json.Marshal(r.Permissions)
	if err := s.DB.WithContext(ctx).Exec(`UPDATE "IdentityCustomRoles" SET name = ?, description = ?, permissions = ?::jsonb WHERE id = ? AND workspace_id = ?`,
		r.Name, r.Description, string(raw), id, ws).Error; err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	s.log(ctx, ws, "role.updated", "role", id, map[string]any{"name": r.Name, "permissions": r.Permissions})
	return r, nil
}

// DeleteRole removes a role and takes it from everyone who held it.
func (s *Service) DeleteRole(ctx context.Context, ws, id string) error {
	r, err := s.role(ctx, ws, id)
	if err != nil {
		return err
	}
	if err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM "IdentityRoleAssignments" WHERE role_id = ? AND workspace_id = ?`, id, ws).Error; err != nil {
			return err
		}
		return tx.Exec(`DELETE FROM "IdentityCustomRoles" WHERE id = ? AND workspace_id = ?`, id, ws).Error
	}); err != nil {
		return err
	}
	s.log(ctx, ws, "role.deleted", "role", id, map[string]any{"name": r.Name})
	return nil
}

// Assignment is who holds a role.
type Assignment struct {
	SubjectType string `json:"subjectType" gorm:"column:subject_type"`
	SubjectID   string `json:"subjectId" gorm:"column:subject_id"`
	Name        string `json:"name"`
}

// Assign gives a role to a person or a team of the workspace.
func (s *Service) Assign(ctx context.Context, ws, roleID, subjectType, subjectID string) error {
	role, err := s.role(ctx, ws, roleID)
	if err != nil {
		return err
	}
	switch subjectType {
	case SubjectUser:
		if ok, err := s.IsMember(ctx, ws, subjectID); err != nil {
			return err
		} else if !ok {
			return ErrNotMember
		}
	case SubjectTeam:
		if _, err := s.team(ctx, ws, subjectID); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: subjectType must be user or team", ErrInvalid)
	}
	res := s.DB.WithContext(ctx).Exec(`INSERT INTO "IdentityRoleAssignments" (role_id, subject_type, subject_id, workspace_id) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING`,
		roleID, subjectType, subjectID, ws)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // already held
	}
	s.log(ctx, ws, "role.assigned", "role", roleID, map[string]any{"name": role.Name, "subjectType": subjectType, "subjectId": subjectID})
	// Tell everyone who now holds the role through this: the person, or the team's people and those of the teams beneath it.
	for _, u := range s.recipients(ctx, ws, subjectType, subjectID) {
		s.tell(ctx, ws, u, "You were given the role "+role.Name, "A workspace admin gave you the "+role.Name+" role. What it allows is now part of your access.")
	}
	return nil
}

// recipients returns the people a role given to subject reaches.
func (s *Service) recipients(ctx context.Context, ws, subjectType, subjectID string) []string {
	if subjectType == SubjectUser {
		return []string{subjectID}
	}
	var all []Team
	if err := s.DB.WithContext(ctx).Where("workspace_id = ?", ws).Find(&all).Error; err != nil {
		return nil
	}
	kids := map[string][]string{}
	for _, t := range all {
		if t.ParentID != nil {
			kids[*t.ParentID] = append(kids[*t.ParentID], t.ID)
		}
	}
	ids, queue, seen := []string{}, []string{subjectID}, map[string]bool{}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if seen[cur] {
			continue
		}
		seen[cur] = true
		ids = append(ids, cur)
		queue = append(queue, kids[cur]...)
	}
	var users []string
	s.DB.WithContext(ctx).Raw(`SELECT DISTINCT user_id FROM "IdentityTeamMembers" WHERE workspace_id = ? AND team_id IN ?`, ws, ids).Scan(&users)
	return users
}

// Unassign takes a role from a person or team; not holding it is not an error.
func (s *Service) Unassign(ctx context.Context, ws, roleID, subjectType, subjectID string) error {
	role, err := s.role(ctx, ws, roleID)
	if err != nil {
		return err
	}
	res := s.DB.WithContext(ctx).Exec(`DELETE FROM "IdentityRoleAssignments" WHERE role_id = ? AND subject_type = ? AND subject_id = ? AND workspace_id = ?`,
		roleID, subjectType, subjectID, ws)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		s.log(ctx, ws, "role.unassigned", "role", roleID, map[string]any{"name": role.Name, "subjectType": subjectType, "subjectId": subjectID})
	}
	return nil
}

// Assignments lists who holds a role, with names.
func (s *Service) Assignments(ctx context.Context, ws, roleID string) ([]Assignment, error) {
	if _, err := s.role(ctx, ws, roleID); err != nil {
		return nil, err
	}
	out := []Assignment{}
	err := s.DB.WithContext(ctx).Raw(`
		SELECT a.subject_type, a.subject_id, COALESCE(u.name, t.name, '') AS name
		FROM "IdentityRoleAssignments" a
		LEFT JOIN "IdentityUsers" u ON a.subject_type = 'user' AND u.id = a.subject_id
		LEFT JOIN "IdentityTeams" t ON a.subject_type = 'team' AND t.id = a.subject_id
		WHERE a.role_id = ? AND a.workspace_id = ? ORDER BY a.subject_type, lower(COALESCE(u.name, t.name, '')), a.subject_id`, roleID, ws).Scan(&out).Error
	return out, err
}

// --- resolution -----------------------------------------------------------

// Granted is what a person's organisation position gives them in a workspace.
type Granted struct {
	Permissions []string // custom-role permission strings
	Roles       []string // custom role names
	RoleIDs     []string // custom role ids
	Teams       []string // team ids the person is directly in
	// OrgUnits is Teams plus every team above them, so a rule written for a
	// department (an app switched off for it) applies to the teams beneath it.
	OrgUnits []string
}

// For resolves a person's custom roles and teams in the workspace. A person who
// holds no seat in it gets nothing, whatever rows exist.
func (s *Service) For(ctx context.Context, ws, userID string) (Granted, error) {
	g := Granted{Permissions: []string{}, Roles: []string{}, RoleIDs: []string{}, Teams: []string{}, OrgUnits: []string{}}
	if ws == "" || userID == "" {
		return g, nil
	}
	if ok, err := s.IsMember(ctx, ws, userID); err != nil || !ok {
		return g, err
	}
	teams, err := s.TeamIDsOf(ctx, ws, userID)
	if err != nil {
		return g, err
	}
	g.Teams = teams
	// A role held by a team reaches its members and the teams beneath it, so
	// collect each of the person's teams and every team above them.
	subject := map[string]bool{}
	for _, t := range teams {
		anc, err := s.ancestors(ctx, ws, t)
		if err != nil {
			return g, err
		}
		for a := range anc {
			subject[a] = true
		}
	}
	ids := make([]string, 0, len(subject))
	for id := range subject {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	g.OrgUnits = ids
	var rows []roleRow
	q := s.DB.WithContext(ctx).Table(`"IdentityCustomRoles" r`).Select("r.*").
		Joins(`JOIN "IdentityRoleAssignments" a ON a.role_id = r.id AND a.workspace_id = r.workspace_id`).
		Where("r.workspace_id = ?", ws)
	if len(ids) > 0 {
		q = q.Where("(a.subject_type = 'user' AND a.subject_id = ?) OR (a.subject_type = 'team' AND a.subject_id IN ?)", userID, ids)
	} else {
		q = q.Where("a.subject_type = 'user' AND a.subject_id = ?", userID)
	}
	if err := q.Distinct().Scan(&rows).Error; err != nil {
		return g, err
	}
	seen := map[string]bool{}
	for _, r := range rows {
		role := r.role()
		g.Roles = append(g.Roles, role.Name)
		g.RoleIDs = append(g.RoleIDs, role.ID)
		for _, p := range role.Permissions {
			if !seen[p] {
				seen[p] = true
				g.Permissions = append(g.Permissions, p)
			}
		}
	}
	sort.Strings(g.Permissions)
	sort.Strings(g.Roles)
	sort.Strings(g.RoleIDs)
	return g, nil
}
