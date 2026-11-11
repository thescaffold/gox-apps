package org

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

// A workspace's handle is its readable, unique address (for example the
// /{org} segment of a URL): lowercase letters, digits and single hyphens,
// 2 to 40 characters, starting and ending with a letter or digit. Hosts
// reserve the words their own URLs use (ReserveHandles). A workspace's region
// is where its data lives; it is chosen once, from the regions the host
// offers (SetRegions).

var handleRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9]|-[a-z0-9]){1,39}$`)

// Errors a caller can match.
var (
	ErrHandleInvalid = errors.New("use 2 to 40 lowercase letters, digits or single hyphens, starting and ending with a letter or digit")
	ErrHandleTaken   = errors.New("that address is taken")
	ErrRegion        = errors.New("that region is not available")
)

var (
	namesMu  sync.RWMutex
	reserved = map[string]bool{
		"admin": true, "api": true, "app": true, "apps": true, "auth": true, "dashboard": true, "docs": true,
		"help": true, "invite": true, "join": true, "login": true, "logout": true, "new": true, "onboarding": true,
		"pricing": true, "security": true, "settings": true, "signup": true, "start": true, "status": true,
		"support": true, "www": true, "guard": true, "error": true, "account": true, "billing": true,
	}
	regions = []string{"eu"}
)

// ReserveHandles adds words no workspace may take.
func ReserveHandles(words ...string) {
	namesMu.Lock()
	defer namesMu.Unlock()
	for _, w := range words {
		reserved[strings.ToLower(strings.TrimSpace(w))] = true
	}
}

// SetRegions replaces the regions a new workspace may choose (the first is the default).
func SetRegions(list ...string) {
	namesMu.Lock()
	defer namesMu.Unlock()
	if len(list) > 0 {
		regions = append([]string(nil), list...)
	}
}

// Regions returns the regions on offer, default first.
func Regions() []string {
	namesMu.RLock()
	defer namesMu.RUnlock()
	return append([]string(nil), regions...)
}

// Region checks r ("" means the default) and returns the region to store.
func Region(r string) (string, error) {
	rs := Regions()
	r = strings.ToLower(strings.TrimSpace(r))
	if r == "" {
		return rs[0], nil
	}
	for _, x := range rs {
		if x == r {
			return r, nil
		}
	}
	return "", ErrRegion
}

// ValidHandle reports whether h may be a handle.
func ValidHandle(h string) error {
	if !handleRe.MatchString(h) {
		return ErrHandleInvalid
	}
	namesMu.RLock()
	defer namesMu.RUnlock()
	if reserved[h] {
		return ErrHandleTaken
	}
	return nil
}

// HandleFor makes a handle from a name: "Acme Corp" → "acme-corp". The result
// may still be taken; FreeHandle finds one that is not.
func HandleFor(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range norm.NFKD.String(strings.ToLower(name)) {
		switch {
		case r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			b.WriteRune(r)
			dash = false
		case unicode.Is(unicode.Mn, r):
			// accents fall away: é → e
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	h := strings.Trim(b.String(), "-")
	if len(h) > 36 {
		h = strings.Trim(h[:36], "-")
	}
	if len(h) < 2 {
		h = "org" + h
	}
	return h
}

// FreeHandle returns base, or base-2, base-3… — the first that is valid and
// not used by another workspace.
func FreeHandle(ctx context.Context, db *gorm.DB, base string) (string, error) {
	return FreeHandleWith(base, func(h string) (bool, error) {
		var n int64
		err := db.WithContext(ctx).Table(`"IdentityWorkspaces"`).Where(`lower("handle") = ?`, h).Count(&n).Error
		return n > 0, err
	})
}

// FreeHandleWith is FreeHandle over the caller's own lookup (taken reports
// whether a workspace already has the handle).
func FreeHandleWith(base string, taken func(h string) (bool, error)) (string, error) {
	if ValidHandle(base) != nil {
		base = HandleFor(base)
	}
	for i := 1; i < 1000; i++ {
		h := base
		if i > 1 {
			h = fmt.Sprintf("%s-%d", base, i)
		}
		if ValidHandle(h) != nil {
			continue
		}
		used, err := taken(h)
		if err != nil {
			return "", err
		}
		if !used {
			return h, nil
		}
	}
	return "", ErrHandleTaken
}

// SetHandle changes a workspace's handle; only its owners and admins may.
func (s *Service) SetHandle(ctx context.Context, ws, actorID, handle string) error {
	handle = strings.ToLower(strings.TrimSpace(handle))
	if err := ValidHandle(handle); err != nil {
		return err
	}
	if ok, err := s.Allowed(ctx, ws, actorID, ManageSettings); err != nil {
		return err
	} else if !ok {
		return ErrNotAllowed
	}
	var n int64
	if err := s.DB.WithContext(ctx).Table(`"IdentityWorkspaces"`).Where(`lower("handle") = ? AND "id" <> ?`, handle, ws).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return ErrHandleTaken
	}
	err := s.DB.WithContext(ctx).Exec(`UPDATE "IdentityWorkspaces" SET "handle" = ?, "updated_at" = now() WHERE "id" = ?`, handle, ws).Error
	if err != nil && strings.Contains(err.Error(), "uq_identity_workspaces_handle") {
		return ErrHandleTaken
	}
	if err == nil {
		s.log(ctx, ws, "workspace.handle", "workspace", ws, map[string]any{"handle": handle, "by": actorID})
	}
	return err
}

// Rename changes a workspace's name; only its owners and admins may.
func (s *Service) Rename(ctx context.Context, ws, actorID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 80 {
		return fmt.Errorf("%w: a name is 1 to 80 characters", ErrInvalid)
	}
	if ok, err := s.Allowed(ctx, ws, actorID, ManageSettings); err != nil {
		return err
	} else if !ok {
		return ErrNotAllowed
	}
	if err := s.DB.WithContext(ctx).Exec(`UPDATE "IdentityWorkspaces" SET "name" = ?, "updated_at" = now() WHERE "id" = ?`, name, ws).Error; err != nil {
		return err
	}
	s.log(ctx, ws, "workspace.rename", "workspace", ws, map[string]any{"name": name, "by": actorID})
	return nil
}
