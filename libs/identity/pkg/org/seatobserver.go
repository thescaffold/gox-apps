package org

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

// How a seat change came about (SeatChange.Via).
const (
	// ViaMembers: an owner or admin changed or removed the seat (SetSeat, RemoveSeat).
	ViaMembers = "members"
	// ViaInvitation: the person accepted an invitation (invitation.Accept, AcceptAs).
	ViaInvitation = "invitation"
)

// SeatChange is one change identity makes to a person's seat in a workspace.
type SeatChange struct {
	WorkspaceID string
	// ActorID is who made the change: the owner or admin for ViaMembers, whoever sent the
	// invitation for ViaInvitation (the person who accepted, when the sender is unknown).
	ActorID string
	// UserID is the person whose seat changed.
	UserID string
	// From and To are the seats before and after; "" is no seat (a join has no From, a
	// removal no To).
	From, To string
	Via      string
	// Teams is how many teams a removal took the person out of.
	Teams int64
}

// SeatObserver is the host's record of a seat change (its audit log, say). It runs inside the
// transaction that makes the change, so what it writes exists if and only if the change does;
// its error undoes the change.
type SeatObserver func(ctx context.Context, tx *gorm.DB, c SeatChange) error

var seatObserver struct {
	mu sync.RWMutex
	fn SeatObserver
}

// SetSeatObserver installs the host's observer of seat changes; nil removes it.
func SetSeatObserver(fn SeatObserver) {
	seatObserver.mu.Lock()
	seatObserver.fn = fn
	seatObserver.mu.Unlock()
}

// SeatObserverHook is the installed observer (nil when none), for hosts that change seats
// outside identity (SCIM provisioning, say) and want them recorded the same way.
func SeatObserverHook() SeatObserver {
	seatObserver.mu.RLock()
	defer seatObserver.mu.RUnlock()
	return seatObserver.fn
}

// ObserveSeat hands c to the installed observer, inside tx; with none it does nothing.
func ObserveSeat(ctx context.Context, tx *gorm.DB, c SeatChange) error {
	if fn := SeatObserverHook(); fn != nil {
		return fn(ctx, tx, c)
	}
	return nil
}
