package org

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// ErrNotAllowed: the caller's own seat does not permit the change.
var ErrNotAllowed = errors.New("not allowed")

// A seat is a person's role in a workspace, held in
// IdentityUserClientWorkspaces.status and carried in the tokens identity
// issues. The default seats are owner, admin, member and guest. A product
// with a richer role table registers its own ordered list with SetSeatRoles.

var (
	seatsMu    sync.RWMutex
	seatRoles  = []string{"owner", "admin", "member", "guest"}
	seatRankOf = rankOf(seatRoles)
)

func rankOf(roles []string) map[string]int {
	m := make(map[string]int, len(roles))
	for i, r := range roles {
		m[r] = len(roles) - i
	}
	return m
}

// SetSeatRoles replaces the seats a workspace may grant, highest first. It
// must keep owner first and admin second, because IsAdmin and the owner rules
// below rely on them.
func SetSeatRoles(roles []string) error {
	clean := make([]string, 0, len(roles))
	seen := map[string]bool{}
	for _, r := range roles {
		r = strings.ToLower(strings.TrimSpace(r))
		if r == "" || seen[r] {
			return fmt.Errorf("%w: seat roles must be unique and non-empty", ErrInvalid)
		}
		seen[r] = true
		clean = append(clean, r)
	}
	if len(clean) < 2 || clean[0] != "owner" || clean[1] != "admin" {
		return fmt.Errorf("%w: seat roles must start with owner, admin", ErrInvalid)
	}
	seatsMu.Lock()
	defer seatsMu.Unlock()
	seatRoles = clean
	seatRankOf = rankOf(clean)
	return nil
}

// SeatRoles returns the seats a workspace may grant, highest first.
func SeatRoles() []string {
	seatsMu.RLock()
	defer seatsMu.RUnlock()
	return append([]string(nil), seatRoles...)
}

func seatRank(role string) int {
	seatsMu.RLock()
	defer seatsMu.RUnlock()
	return seatRankOf[strings.ToLower(role)]
}

// Seat returns userID's seat in the workspace, or "" when they hold none.
func (s *Service) Seat(ctx context.Context, ws, userID string) (string, error) {
	var status []string
	err := s.DB.WithContext(ctx).Table(`"IdentityUserClientWorkspaces"`).
		Where("workspace_id = ? AND user_id = ? AND deleted_at IS NULL", ws, userID).
		Order("created_at").Limit(1).Pluck("status", &status).Error
	if err != nil || len(status) == 0 {
		return "", err
	}
	return strings.ToLower(status[0]), nil
}

func (s *Service) owners(ctx context.Context, ws string) (int64, error) {
	var n int64
	err := s.DB.WithContext(ctx).Table(`"IdentityUserClientWorkspaces"`).
		Where("workspace_id = ? AND deleted_at IS NULL AND lower(status) = 'owner'", ws).Count(&n).Error
	return n, err
}

// checkSeatChange applies the rules every seat change follows: the actor is an
// owner or admin; nobody grants or takes away a seat above their own; only an
// owner grants or removes owner; and the last owner keeps their seat.
func (s *Service) checkSeatChange(ctx context.Context, ws, actorID, userID, to string) (from string, err error) {
	actor, err := s.Seat(ctx, ws, actorID)
	if err != nil {
		return "", err
	}
	ok, err := Authorize(ctx, ws, actorID, ManageMembers, func() (bool, error) { return actor == "owner" || actor == "admin", nil })
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("%w: only an owner or admin can change seats", ErrNotAllowed)
	}
	from, err = s.Seat(ctx, ws, userID)
	if err != nil {
		return "", err
	}
	if from == "" {
		return "", ErrNotMember
	}
	if seatRank(from) > seatRank(actor) || (to != "" && seatRank(to) > seatRank(actor)) {
		return "", fmt.Errorf("%w: you cannot change a seat above your own", ErrNotAllowed)
	}
	if (from == "owner" || to == "owner") && actor != "owner" {
		return "", fmt.Errorf("%w: only an owner can grant or remove the owner seat", ErrNotAllowed)
	}
	if from == "owner" && to != "owner" {
		n, err := s.owners(ctx, ws)
		if err != nil {
			return "", err
		}
		if n <= 1 {
			return "", fmt.Errorf("%w: the workspace must keep at least one owner", ErrInvalid)
		}
	}
	return from, nil
}

// SetSeat changes userID's seat in the workspace to role, acting as actorID. The host's seat
// observer (SetSeatObserver) runs in the same transaction.
func (s *Service) SetSeat(ctx context.Context, ws, actorID, userID, role string) error {
	role = strings.ToLower(strings.TrimSpace(role))
	if seatRank(role) == 0 {
		return fmt.Errorf("%w: role must be one of %s", ErrInvalid, strings.Join(SeatRoles(), ", "))
	}
	from, err := s.checkSeatChange(ctx, ws, actorID, userID, role)
	if err != nil {
		return err
	}
	if from == role {
		return nil
	}
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`UPDATE "IdentityUserClientWorkspaces" SET status = ?, updated_at = now()
			WHERE workspace_id = ? AND user_id = ? AND deleted_at IS NULL`, role, ws, userID).Error; err != nil {
			return err
		}
		return ObserveSeat(ctx, tx, SeatChange{WorkspaceID: ws, ActorID: actorID, UserID: userID, From: from, To: role, Via: ViaMembers})
	})
	if err != nil {
		return err
	}
	s.log(ctx, ws, "member.seat.changed", "user", userID, map[string]any{"from": from, "to": role})
	s.tell(ctx, ws, userID, "Your role changed", "You are now "+role+" in this workspace.")
	return nil
}

// RemoveSeat takes userID out of the workspace (and its teams), acting as actorID. The host's
// clean-up (OnSeatRemoved), then its seat observer (SetSeatObserver), run in the same transaction.
func (s *Service) RemoveSeat(ctx context.Context, ws, actorID, userID string) error {
	from, err := s.checkSeatChange(ctx, ws, actorID, userID, "")
	if err != nil {
		return err
	}
	var teams int64
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`UPDATE "IdentityUserClientWorkspaces" SET deleted_at = now()
			WHERE workspace_id = ? AND user_id = ? AND deleted_at IS NULL`, ws, userID).Error; err != nil {
			return err
		}
		res := tx.Exec(`DELETE FROM "IdentityTeamMembers" WHERE workspace_id = ? AND user_id = ?`, ws, userID)
		if res.Error != nil {
			return res.Error
		}
		teams = res.RowsAffected
		if fn := SeatRemovedHook(); fn != nil {
			if err := fn(ctx, tx, ws, userID, actorID); err != nil {
				return err
			}
		}
		return ObserveSeat(ctx, tx, SeatChange{WorkspaceID: ws, ActorID: actorID, UserID: userID, From: from, Via: ViaMembers, Teams: teams})
	})
	if err != nil {
		return err
	}
	s.log(ctx, ws, "member.seat.removed", "user", userID, map[string]any{"from": from, "teams": teams})
	return nil
}

// CheckGrant applies the rules for granting seat to someone new (an
// invitation): the actor is an owner or admin, the seat exists, nobody grants
// a seat above their own, and only an owner grants owner.
func (s *Service) CheckGrant(ctx context.Context, ws, actorID, seat string) error {
	seat = strings.ToLower(strings.TrimSpace(seat))
	if seatRank(seat) == 0 {
		return fmt.Errorf("%w: role must be one of %s", ErrInvalid, strings.Join(SeatRoles(), ", "))
	}
	actor, err := s.Seat(ctx, ws, actorID)
	if err != nil {
		return err
	}
	ok, err := Authorize(ctx, ws, actorID, ManageMembers, func() (bool, error) { return actor == "owner" || actor == "admin", nil })
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: only an owner or admin can invite people", ErrNotAllowed)
	}
	if seatRank(seat) > seatRank(actor) {
		return fmt.Errorf("%w: you cannot invite someone above your own role", ErrNotAllowed)
	}
	if seat == "owner" && actor != "owner" {
		return fmt.Errorf("%w: only an owner can invite an owner", ErrNotAllowed)
	}
	return nil
}
