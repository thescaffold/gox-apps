package provider

import (
	"encoding/json"
	"os"
	"time"

	"github.com/cbroglie/mustache"
	notificationlog "github.com/thescaffold/gox-apps-notification/app/log"
	notificationmessage "github.com/thescaffold/gox-apps-notification/app/message"
	notificationrule "github.com/thescaffold/gox-apps-notification/app/rule"
	notificationtemplate "github.com/thescaffold/gox-apps-notification/app/template"
)

type ProviderService struct {
	templateEntity *notificationtemplate.TemplateEntity `inject:""`
	ruleEntity     *notificationrule.RuleEntity         `inject:""`
	logEntity      *notificationlog.LogEntity           `inject:""`
	messageEntity  *notificationmessage.MessageEntity   `inject:""`
	mailgun        *MailgunProvider                     `inject:""`
	zoho           *ZohoProvider                        `inject:""`
}

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

	provider := os.Getenv("NOTIFICATION_DEFAULT_PROVIDER")
	if provider == "" {
		provider = "mailgun"
	}

	var results []map[string]any
	for _, channel := range channels {
		if channel == "" {
			continue
		}

		if notification != nil {
			logType := ""
		if log.Type != nil {
			logType = *log.Type
		}
		if allowed, ok := notification[logType]; ok {
				if !containsStr(toStrSlice(allowed), channel) {
					continue
				}
			}
		}

		templateValue := s.templateField(tmpl, channel)
		if templateValue == "" {
			continue
		}

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
			ok = false
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

func (s *ProviderService) emailProvider(providerName string, msg EmailMessage) bool {
	switch providerName {
	case "zoho":
		return s.zoho.Email(msg)
	default:
		return s.mailgun.Email(msg)
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
	ref := to + "-" + log.Key + "-" + time.Now().Format("20060102150405")
	m.Reference = &ref
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
