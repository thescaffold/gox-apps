package provider

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/cbroglie/mustache"
	notificationlog "github.com/thescaffold/gox-apps/libs/notification/app/log"
	notificationmessage "github.com/thescaffold/gox-apps/libs/notification/app/message"
	notificationrule "github.com/thescaffold/gox-apps/libs/notification/app/rule"
	notificationtemplate "github.com/thescaffold/gox-apps/libs/notification/app/template"
)

type ProviderService struct {
	templateEntity *notificationtemplate.TemplateEntity `inject:""`
	ruleEntity     *notificationrule.RuleEntity         `inject:""`
	logEntity      *notificationlog.LogEntity           `inject:""`
	messageEntity  *notificationmessage.MessageEntity   `inject:""`
	mailgun        *MailgunProvider                     `inject:""`
	zoho           *ZohoProvider                        `inject:""`
	cloudflare     *CloudflareProvider                  `inject:""`
	termii         *TermiiProvider                      `inject:""`
	africastalking *AfricastalkingProvider              `inject:""`

	// Identity is an optional context enricher. When set, every template render
	// receives {user, address, preference, organization, now} alongside the raw
	// event data — mirroring ntx-apps/libs/notification/src/pkg/provider/provider.service.ts
	// lines 75-187. nil ⇒ render with raw data only (back-compat).
	Identity IdentityFetcher

	// queuePusher, when non-nil, is preferred over synchronous Send by Dispatch.
	queuePusher QueuePusher
}

// SetIdentity wires (or replaces) the IdentityFetcher used for context
// enrichment. Safe to call before or after DI; not goroutine-safe.
func (s *ProviderService) SetIdentity(f IdentityFetcher) { s.Identity = f }

// Send processes a notification log: looks up the template, renders it, and
// dispatches to each requested channel, then updates the log with results.
func (s *ProviderService) Send(log *notificationlog.Log) {
	if log.UserId == nil || *log.UserId == "" {
		return
	}

	groupName := os.Getenv("GROUP_NAME")
	tmpl, err := s.templateEntity.First(
		`"key" = ? AND "group" = ?`, log.Key, groupName,
	)
	if err != nil || tmpl == nil {
		return
	}

	var channels []string
	if err := json.Unmarshal(log.Channels, &channels); err != nil || len(channels) == 0 {
		return
	}

	var data map[string]any
	if log.Data != nil {
		_ = json.Unmarshal(log.Data, &data)
	}
	if data == nil {
		data = map[string]any{}
	}

	to := ""
	if v, ok := data["owner"].(string); ok && v != "" {
		to = v
	}
	if to == "" && s.Identity != nil && log.UserId != nil {
		// TS: to = data?.owner ?? user?.ref — fall back to the resolved user's ref.
		if uc, err := s.Identity.FetchByUserId(*log.UserId); err == nil && uc != nil {
			if ref, ok := uc.User["ref"].(string); ok {
				to = ref
			}
		}
	}
	if to == "" {
		return
	}

	var notification map[string]any
	rule, err := s.ruleEntity.First(
		`"group" = ? AND "key" = ?`, "subscription", to,
	)
	if err == nil && rule != nil {
		_ = json.Unmarshal(rule.Rules, &notification)
	}

	// Enrich the render context with identity-resolved user/address/preference/organization,
	// plus a `now` block of formatted timestamps. Falls back to raw data when no
	// IdentityFetcher is wired or the lookup returns nil.
	enrichRenderContext(data, s.Identity, log, to)

	var results []map[string]any
	for _, channel := range channels {
		if channel == "" {
			continue
		}

		// TS always restricts to typeChannels: the subscription rule's value when
		// the log type is present in it, else defaultMessageTypeChannels(type)
		// (= web,email). A channel outside that set is skipped — so e.g. sms is
		// dropped when there is no rule, matching TS.
		logType := ""
		if log.Type != nil {
			logType = *log.Type
		}
		var typeChannels []string
		if v, ok := notification[logType]; ok {
			typeChannels = channelList(v)
		} else {
			typeChannels = defaultMessageTypeChannels()
		}
		if !containsStr(typeChannels, channel) {
			continue
		}

		templateValue := s.templateField(tmpl, channel)
		if templateValue == "" {
			continue
		}

		provider := resolveProvider(channel)

		rendered, err := mustache.Render(templateValue, data)
		if err != nil {
			rendered = templateValue
		}

		msg := EmailMessage{
			To:      to,
			Subject: log.Subject,
			HTML:    rendered,
		}

		var ok bool
		switch channel {
		case "email":
			ok = s.emailProvider(provider, msg)
		case "sms":
			ok = s.smsProvider(provider, SMSMessage{
				To:      to,
				Subject: log.Subject,
				Text:    rendered,
			})
		case "web":
			ok = s.saveWebMessage(log, to)
		case "mobile":
			ok = false
		}

		results = append(results, map[string]any{"channel": channel, "status": ok})
	}

	resultJSON, _ := json.Marshal(results)
	allOk := len(results) > 0
	for _, r := range results {
		if !r["status"].(bool) {
			allOk = false
			break
		}
	}
	status := "failed"
	if allOk {
		status = "success"
	}
	updated := &notificationlog.Log{Status: &status, Message: resultJSON}
	if tmpl != nil {
		updated.TemplateId = &tmpl.Id
	}
	_, _ = s.logEntity.Update(updated, `"id" = ?`, log.Id)
}

// resolveProvider picks the provider for a channel: a channel-specific override
// (NOTIFICATION_EMAIL_DEFAULT_PROVIDER / NOTIFICATION_SMS_DEFAULT_PROVIDER) wins
// over the shared NOTIFICATION_DEFAULT_PROVIDER. Falls back to "mailgun" when
// nothing is configured (matches prior gox behaviour).
func resolveProvider(channel string) string {
	switch channel {
	case "email":
		if v := os.Getenv("NOTIFICATION_EMAIL_DEFAULT_PROVIDER"); v != "" {
			return v
		}
	case "sms":
		if v := os.Getenv("NOTIFICATION_SMS_DEFAULT_PROVIDER"); v != "" {
			return v
		}
	}
	if v := os.Getenv("NOTIFICATION_DEFAULT_PROVIDER"); v != "" {
		return v
	}
	return "mailgun"
}

func (s *ProviderService) emailProvider(providerName string, msg EmailMessage) bool {
	switch providerName {
	case "zoho":
		return s.zoho.Email(msg)
	case "cloudflare":
		return s.cloudflare.Email(msg)
	default:
		return s.mailgun.Email(msg)
	}
}

// smsProvider routes the sms channel to the configured default provider,
// mirroring TS this.use(provider).sms(message). Only termii and africastalking
// implement sms, so any other NOTIFICATION_DEFAULT_PROVIDER yields false —
// matching mailgun/zoho whose .sms() return false in TS.
func (s *ProviderService) smsProvider(providerName string, msg SMSMessage) bool {
	switch providerName {
	case "termii":
		return s.termii.SMS(msg)
	case "africastalking":
		return s.africastalking.SMS(msg)
	default:
		return false
	}
}

func (s *ProviderService) saveWebMessage(log *notificationlog.Log, to string) bool {
	channel := "web"
	m := &notificationmessage.Message{
		UserId:      log.UserId,
		ClientId:    log.ClientId,
		WorkspaceId: log.WorkspaceId,
		Key:         &log.Key,
		Subject:     &log.Subject,
		Channel:     &channel,
		Type:        log.Type,
		Priority:    log.Priority,
		Theme:       log.Theme,
		Scope:       log.Scope,
		Position:    log.Position,
		Status:      log.Status,
		ReadAt:      log.ReadAt,
		PublishAt:   log.PublishAt,
		ExpireAt:    log.ExpireAt,
	}
	// TS web() persists the Message with reference = log.reference (not a freshly
	// minted to-key-timestamp) and data = log.data.
	m.Reference = &log.Reference
	m.Data = log.Data
	return s.messageEntity.Insert(m) == nil
}

func (s *ProviderService) templateField(tmpl *notificationtemplate.Template, channel string) string {
	switch channel {
	case "email":
		if tmpl.Email != nil {
			return *tmpl.Email
		}
	case "sms":
		if tmpl.Sms != nil {
			return *tmpl.Sms
		}
	case "web":
		if tmpl.Web != nil {
			return *tmpl.Web
		}
	case "mobile":
		if tmpl.Mobile != nil {
			return *tmpl.Mobile
		}
	}
	return ""
}

// defaultMessageTypeChannels mirrors TS app.config.ts: every message type maps
// to "web,email".
func defaultMessageTypeChannels() []string { return []string{"web", "email"} }

// channelList normalises a subscription rule value (a comma-joined string or a
// JSON array) into a channel slice, matching TS `typeChannels.includes(channel)`
// over either shape.
func channelList(v any) []string {
	switch t := v.(type) {
	case string:
		parts := strings.Split(t, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		return out
	case []any:
		return toStrSlice(t)
	}
	return nil
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func toStrSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
