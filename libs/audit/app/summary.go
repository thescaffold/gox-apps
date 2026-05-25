package app

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	auditlog "github.com/thescaffold/gox-apps/libs/audit/app/log"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// prettyLayout mirrors core FORMAT.PRETTY ('ddd, DD MMM YYYY, hh:ss A').
const prettyLayout = "Mon, 02 Jan 2006, 03:05 PM"

var (
	entityPrefixRe  = regexp.MustCompile(`(?i)^(Origine|Identity|Capital|Common|Notification|Flag|Health|Cron|Queue)`)
	camelBoundaryRe = regexp.MustCompile(`([A-Z])`)
)

// RunSummary ports ntx-apps/libs/audit SummaryBatch + its app.controller heartbeat
// wiring: for each user with audit logs in the previous <period>, aggregate a
// comprehensive metrics summary and emit an `apps.notification.message.new`
// tracker message. period is "weekly" | "monthly" | "yearly".
func (s *AppService) RunSummary(period string) error {
	unit := convertPeriodType(period)
	label := convertPeriodLabel(period)
	from, to := periodRange(unit)

	logs, err := s.logService.LogsInRange(from, to)
	if err != nil {
		return err
	}
	if len(logs) == 0 || s.tracker == nil {
		return nil
	}

	// Group logs by user, preserving created_at DESC order (entity default sort),
	// so each user's slice is newest-first like the TS per-user query.
	order := make([]string, 0)
	byUser := map[string][]auditlog.Log{}
	for _, l := range logs {
		if _, seen := byUser[l.UserId]; !seen {
			order = append(order, l.UserId)
		}
		byUser[l.UserId] = append(byUser[l.UserId], l)
	}

	now := time.Now().UTC()
	expireAt := addPeriod(now, unit)
	for _, userID := range order {
		userLogs := byUser[userID]
		ref := utils.Reference("SMR", 36)
		data := aggregateMetrics(userLogs, ref, label, from, to)
		s.tracker.Message("apps.notification.message.new", map[string]any{
			"reference":   ref,
			"userId":      userID,
			"clientId":    userLogs[0].ClientId,
			"workspaceId": userLogs[0].WorkspaceId,
			"key":         "summary",
			"channels":    []string{"email"},
			"data":        data,
			"subject":     "Your " + label + " Summary",
			"type":        "system",
			"priority":    "medium",
			"theme":       "info",
			"scope":       "user",
			"position":    "bottom right",
			"publishAt":   now,
			"expireAt":    expireAt,
		})
	}
	return nil
}

// ── period helpers (mirror app.config.ts convertPeriod* + dayjs subtract/startOf/endOf) ──

func convertPeriodType(period string) string {
	switch period {
	case "weekly":
		return "week"
	case "monthly":
		return "month"
	case "yearly":
		return "year"
	default:
		return "day"
	}
}

func convertPeriodLabel(period string) string {
	switch period {
	case "weekly":
		return "Weekly"
	case "monthly":
		return "Monthly"
	case "yearly":
		return "Yearly"
	case "daily":
		return "Daily"
	default:
		return ""
	}
}

// periodRange returns the previous period's [start, end] in UTC — the TS
// `now.subtract(1, unit).startOf(unit) .. .endOf(unit)`.
func periodRange(unit string) (time.Time, time.Time) {
	now := time.Now().UTC()
	endNs := 999999999
	switch unit {
	case "year":
		y := now.Year() - 1
		return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(y, 12, 31, 23, 59, 59, endNs, time.UTC)
	case "month":
		firstThis := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		prevEnd := firstThis.Add(-time.Nanosecond) // last instant of previous month
		from := time.Date(prevEnd.Year(), prevEnd.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(prevEnd.Year(), prevEnd.Month(), prevEnd.Day(), 23, 59, 59, endNs, time.UTC)
		return from, to
	case "week":
		// dayjs default week starts Sunday.
		midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		thisSunday := midnight.AddDate(0, 0, -int(now.Weekday()))
		from := thisSunday.AddDate(0, 0, -7)
		to := thisSunday.Add(-time.Nanosecond)
		return from, to
	default: // day
		y := now.AddDate(0, 0, -1)
		return time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, time.UTC),
			time.Date(y.Year(), y.Month(), y.Day(), 23, 59, 59, endNs, time.UTC)
	}
}

// addPeriod mirrors dayjs now.add(1, unit) for the notification expireAt.
func addPeriod(t time.Time, unit string) time.Time {
	switch unit {
	case "year":
		return t.AddDate(1, 0, 0)
	case "month":
		return t.AddDate(0, 1, 0)
	case "week":
		return t.AddDate(0, 0, 7)
	default:
		return t.AddDate(0, 0, 1)
	}
}

// ── aggregation (mirror SummaryBatch.aggregateMetrics) ──

func aggregateMetrics(logs []auditlog.Log, ref, periodLabel string, from, to time.Time) map[string]any {
	var actionCreate, actionUpdate, actionDelete, actionRead int

	type entityMetric struct {
		name                             string
		created, updated, deleted, total int
	}
	entityMap := map[string]*entityMetric{}
	var entityOrder []string

	type serviceMetric struct {
		name     string
		actions  int
		entities []string
		set      map[string]bool
	}
	serviceMap := map[string]*serviceMetric{}
	var serviceOrder []string

	dailyMap := map[string]int{}

	var instancesCreated, instancesTerminated, instancesUpdated int
	var deploymentsTotal, deploymentsSuccessful, deploymentsFailed, deploymentsRollbacks, deploymentEvents int
	var zonesCreated, zonesTotal, recordsCreated int
	var domainsCreated, domainsTotal int
	var sourceNamespaces, sourceResources, sourceTags int
	var snapshotsTotal, snapshotRestores int
	var varTypes, vars int
	var loginCount, sessionCount, deviceCount, inviteCount int
	var paymentsCount, transactionsCount, usagesCount int
	activeWorkspaces := map[string]bool{}
	var workspacesCreated int

	for _, l := range logs {
		action := strings.ToLower(l.Action)
		if action == "" {
			action = "read"
		}
		switch action {
		case "create":
			actionCreate++
		case "update":
			actionUpdate++
		case "delete":
			actionDelete++
		case "read":
			actionRead++
		}

		entityName := l.EntityName
		if entityName == "" {
			entityName = "Unknown"
		}
		em, ok := entityMap[entityName]
		if !ok {
			em = &entityMetric{name: formatEntityName(entityName)}
			entityMap[entityName] = em
			entityOrder = append(entityOrder, entityName)
		}
		em.total++
		switch action {
		case "create":
			em.created++
		case "update":
			em.updated++
		case "delete":
			em.deleted++
		}

		serviceName := l.Service
		if serviceName == "" {
			serviceName = "default"
		}
		sm, ok := serviceMap[serviceName]
		if !ok {
			sm = &serviceMetric{name: utils.TitleCase(serviceName), set: map[string]bool{}}
			serviceMap[serviceName] = sm
			serviceOrder = append(serviceOrder, serviceName)
		}
		sm.actions++
		if !sm.set[entityName] {
			sm.set[entityName] = true
			sm.entities = append(sm.entities, entityName)
		}

		dailyMap[logTime(l).Format("2006-01-02")]++

		if l.WorkspaceId != "" {
			activeWorkspaces[l.WorkspaceId] = true
		}

		entityLower := strings.ToLower(entityName)
		meta := parseMeta(l.Meta)

		switch {
		case entityLower == "origineinstances" || entityLower == "instances":
			switch action {
			case "create":
				instancesCreated++
			case "delete":
				instancesTerminated++
			case "update":
				instancesUpdated++
			}
		}

		if entityLower == "originedeployments" || entityLower == "deployments" {
			deploymentsTotal++
			status := strings.ToLower(strPtr(l.Status))
			if status == "" {
				status = strings.ToLower(metaString(meta, "status"))
			}
			switch status {
			case "success", "completed", "running":
				deploymentsSuccessful++
			case "failed", "error":
				deploymentsFailed++
			}
		}
		if strings.Contains(entityLower, "deploymentevent") || strings.Contains(entityLower, "deploymentlog") {
			deploymentEvents++
		}
		if entityLower == "originezones" || entityLower == "zones" {
			zonesTotal++
			if action == "create" {
				zonesCreated++
			}
		}
		if entityLower == "originerecords" || entityLower == "records" {
			if action == "create" {
				recordsCreated++
			}
		}
		if entityLower == "originedomains" || entityLower == "domains" {
			domainsTotal++
			if action == "create" {
				domainsCreated++
			}
		}
		if strings.Contains(entityLower, "sourcenamespace") {
			sourceNamespaces++
		}
		if strings.Contains(entityLower, "sourceresource") {
			sourceResources++
		}
		if strings.Contains(entityLower, "sourcetag") {
			sourceTags++
		}
		if entityLower == "originesnapshots" || entityLower == "snapshots" {
			snapshotsTotal++
		}
		if strings.Contains(entityLower, "snapshotrestore") {
			snapshotRestores++
		}
		if entityLower == "originevartypes" || entityLower == "vartypes" {
			varTypes++
		}
		if entityLower == "originevars" || entityLower == "vars" {
			vars++
		}

		if entityLower == "identityproviderlogs" || entityLower == "providerlogs" {
			loginCount++
		}
		if strings.Contains(entityLower, "devicesession") || strings.Contains(entityLower, "session") {
			sessionCount++
		}
		if entityLower == "identitydevices" || entityLower == "devices" {
			deviceCount++
		}
		if entityLower == "identityinvites" || entityLower == "invites" {
			inviteCount++
		}
		if entityLower == "identityworkspaces" || entityLower == "workspaces" {
			if action == "create" {
				workspacesCreated++
			}
		}
		if entityLower == "capitalpayments" || entityLower == "payments" {
			paymentsCount++
		}
		if entityLower == "capitaltransactions" || entityLower == "transactions" {
			transactionsCount++
		}
		if entityLower == "capitalusages" || entityLower == "usages" {
			usagesCount++
		}
	}

	// entityMetrics sorted by total desc.
	entityMetrics := make([]map[string]any, 0, len(entityOrder))
	for _, k := range entityOrder {
		em := entityMap[k]
		entityMetrics = append(entityMetrics, map[string]any{
			"name": em.name, "created": em.created, "updated": em.updated,
			"deleted": em.deleted, "total": em.total,
		})
	}
	sort.SliceStable(entityMetrics, func(i, j int) bool {
		return entityMetrics[i]["total"].(int) > entityMetrics[j]["total"].(int)
	})
	topEntities := entityMetrics
	if len(topEntities) > 5 {
		topEntities = topEntities[:5]
	}

	// serviceMetrics sorted by actions desc.
	serviceMetrics := make([]map[string]any, 0, len(serviceOrder))
	for _, k := range serviceOrder {
		sm := serviceMap[k]
		serviceMetrics = append(serviceMetrics, map[string]any{
			"name": sm.name, "actions": sm.actions, "entities": sm.entities,
		})
	}
	sort.SliceStable(serviceMetrics, func(i, j int) bool {
		return serviceMetrics[i]["actions"].(int) > serviceMetrics[j]["actions"].(int)
	})

	// dailyActivity sorted by date asc.
	dates := make([]string, 0, len(dailyMap))
	for d := range dailyMap {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	dailyActivity := make([]map[string]any, 0, len(dates))
	for _, d := range dates {
		dailyActivity = append(dailyActivity, map[string]any{"date": d, "count": dailyMap[d]})
	}

	// recentActivities — first 10 (newest-first).
	recentActivities := make([]map[string]any, 0, 10)
	for i, l := range logs {
		if i >= 10 {
			break
		}
		entityName := l.EntityName
		if entityName == "" {
			entityName = "Item"
		}
		actionLabel := l.Action
		if actionLabel == "" {
			actionLabel = "Activity"
		}
		desc := strPtr(l.Desc)
		if desc == "" {
			a := l.Action
			if a == "" {
				a = "Action"
			}
			desc = utils.TitleCase(a) + " on " + formatEntityName(entityName)
		}
		recentActivities = append(recentActivities, map[string]any{
			"date":        logTime(l).Format(prettyLayout),
			"action":      utils.TitleCase(actionLabel),
			"entityName":  formatEntityName(entityName),
			"description": desc,
		})
	}

	successRate := 0
	if deploymentsTotal > 0 {
		successRate = int(math.Round(float64(deploymentsSuccessful) / float64(deploymentsTotal) * 100))
	}

	highlights := generateHighlights(len(logs), instancesCreated, instancesTerminated,
		deploymentsTotal, deploymentsSuccessful, deploymentsFailed, zonesCreated,
		domainsCreated, snapshotsTotal, loginCount, paymentsCount, topEntities)

	var workspace any
	if len(logs) > 0 {
		workspace = nestedMeta(parseMeta(logs[0].Meta), "context", "workspace")
	}

	return map[string]any{
		"totalActivities": len(logs),
		"periodType":      periodLabel,
		"from":            from.Format(prettyLayout),
		"to":              to.Format(prettyLayout),

		"totalCreated": actionCreate,
		"totalUpdated": actionUpdate,
		"totalDeleted": actionDelete,
		"totalRead":    actionRead,

		"entityMetrics":  entityMetrics,
		"topEntities":    topEntities,
		"serviceMetrics": serviceMetrics,

		"instances": map[string]any{
			"total": instancesCreated, "created": instancesCreated,
			"terminated": instancesTerminated,
			"active":     instancesCreated - instancesTerminated,
			"updated":    instancesUpdated,
		},
		"deployments": map[string]any{
			"total": deploymentsTotal, "successful": deploymentsSuccessful,
			"failed": deploymentsFailed, "rollbacks": deploymentsRollbacks,
			"successRate": successRate, "events": deploymentEvents,
		},
		"zones":      map[string]any{"total": zonesTotal, "created": zonesCreated, "records": recordsCreated},
		"domains":    map[string]any{"total": domainsTotal, "created": domainsCreated},
		"sources":    map[string]any{"namespaces": sourceNamespaces, "resources": sourceResources, "tags": sourceTags},
		"snapshots":  map[string]any{"total": snapshotsTotal, "restores": snapshotRestores},
		"variables":  map[string]any{"types": varTypes, "vars": vars},
		"security":   map[string]any{"logins": loginCount, "sessions": sessionCount, "devices": deviceCount, "invites": inviteCount},
		"billing":    map[string]any{"payments": paymentsCount, "transactions": transactionsCount, "usages": usagesCount},
		"workspaces": map[string]any{"active": len(activeWorkspaces), "created": workspacesCreated},

		"recentActivities": recentActivities,
		"highlights":       highlights,
		"dailyActivity":    dailyActivity,

		"hasActivity":         len(logs) > 0,
		"hasHighlights":       len(highlights) > 0,
		"hasTopEntities":      len(topEntities) > 0,
		"hasRecentActivities": len(recentActivities) > 0,

		"reference": ref,
		"workspace": workspace,
	}
}

func generateHighlights(totalActivities, instancesCreated, instancesTerminated,
	deploymentsTotal, deploymentsSuccessful, deploymentsFailed, zonesCreated,
	domainsCreated, snapshotsTotal, loginCount, paymentsCount int,
	topEntities []map[string]any) []string {

	if totalActivities == 0 {
		return []string{"No activity recorded this period"}
	}
	h := []string{fmt.Sprintf("You performed %d actions this period", totalActivities)}
	if instancesCreated > 0 {
		h = append(h, fmt.Sprintf("%d new instance%s created", instancesCreated, plural(instancesCreated)))
	}
	if instancesTerminated > 0 {
		h = append(h, fmt.Sprintf("%d instance%s terminated", instancesTerminated, plural(instancesTerminated)))
	}
	if deploymentsTotal > 0 {
		rate := int(math.Round(float64(deploymentsSuccessful) / float64(deploymentsTotal) * 100))
		h = append(h, fmt.Sprintf("%d deployment%s with %d%% success rate", deploymentsTotal, plural(deploymentsTotal), rate))
		if deploymentsFailed > 0 {
			h = append(h, fmt.Sprintf("%d deployment%s failed - review recommended", deploymentsFailed, plural(deploymentsFailed)))
		}
	}
	if zonesCreated > 0 {
		h = append(h, fmt.Sprintf("%d new DNS zone%s configured", zonesCreated, plural(zonesCreated)))
	}
	if domainsCreated > 0 {
		h = append(h, fmt.Sprintf("%d custom domain%s added", domainsCreated, plural(domainsCreated)))
	}
	if snapshotsTotal > 0 {
		h = append(h, fmt.Sprintf("%d snapshot%s created", snapshotsTotal, plural(snapshotsTotal)))
	}
	if loginCount > 0 {
		h = append(h, fmt.Sprintf("%d successful login%s recorded", loginCount, plural(loginCount)))
	}
	if paymentsCount > 0 {
		h = append(h, fmt.Sprintf("%d payment%s processed", paymentsCount, plural(paymentsCount)))
	}
	if len(topEntities) > 0 {
		top := topEntities[0]
		h = append(h, fmt.Sprintf("Most activity: %v (%v actions)", top["name"], top["total"]))
	}
	if len(h) > 6 {
		h = h[:6]
	}
	return h
}

// formatEntityName mirrors TS formatEntityName: strip a leading known prefix,
// insert a space before each uppercase letter, trim, then titleCase.
func formatEntityName(name string) string {
	s := entityPrefixRe.ReplaceAllString(name, "")
	s = camelBoundaryRe.ReplaceAllString(s, " $1")
	return utils.TitleCase(strings.TrimSpace(s))
}

func plural(n int) string {
	if n > 1 {
		return "s"
	}
	return ""
}

func logTime(l auditlog.Log) time.Time {
	if l.CreatedAt != nil {
		return l.CreatedAt.UTC()
	}
	return time.Time{}
}

func strPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func parseMeta(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}

func metaString(meta map[string]any, key string) string {
	if meta == nil {
		return ""
	}
	s, _ := meta[key].(string)
	return s
}

func nestedMeta(meta map[string]any, keys ...string) any {
	var cur any = meta
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[k]
	}
	return cur
}
