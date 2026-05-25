package provider

import (
	"time"

	notificationlog "github.com/thescaffold/gox-apps/libs/notification/app/log"
)

// enrichRenderContext layers identity context + a `now` block onto the data
// map passed to the mustache renderer. Mirrors TS provider.service.ts lines
// 146-171 which spread `user`, `address`, `preference`, `organization`, and a
// formatted `now` object into the template scope.
//
// data is mutated in place. Missing/nil dependencies are tolerated: with no
// IdentityFetcher wired, only `now` is added.
func enrichRenderContext(data map[string]any, fetcher IdentityFetcher, log *notificationlog.Log, to string) {
	if data == nil {
		return
	}

	// Always add a `now` block — TS adds this regardless of identity presence.
	// Keys/formats mirror TS provider.service.ts exactly: dateTime/time use the
	// 12-hour FULL_DATE_TIME/FULL_TIME formats, and day/month/year are
	// zero-padded strings (dayjs DD/MM/YYYY). NOTE: no iso/pretty in TS.
	now := time.Now().UTC()
	data["now"] = map[string]any{
		"dateTime": now.Format("2006-01-02 03:04:05"),
		"time":     now.Format("03:04:05"),
		"date":     now.Format("2006-01-02"),
		"day":      now.Format("02"),
		"month":    now.Format("01"),
		"year":     now.Format("2006"),
	}

	if fetcher == nil {
		return
	}

	// Prefer userId lookup; fall back to owner string.
	var ctx *UserContext
	if log != nil && log.UserId != nil && *log.UserId != "" {
		ctx, _ = fetcher.FetchByUserId(*log.UserId)
	}
	if ctx == nil && to != "" {
		ctx, _ = fetcher.FetchByOwner(to)
	}
	if ctx == nil {
		return
	}

	if ctx.User != nil {
		data["user"] = ctx.User
	}
	if ctx.Address != nil {
		data["address"] = ctx.Address
	}
	if ctx.Preference != nil {
		data["preference"] = ctx.Preference
	}
	if ctx.Organization != nil {
		data["organization"] = ctx.Organization
	}
}
