package app

import (
	"errors"
	"fmt"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/thescaffold/gox-apps/libs/notification/app/log"
	"github.com/thescaffold/gox-apps/libs/notification/app/message"
	"github.com/thescaffold/gox-apps/libs/notification/app/rule"
	"github.com/thescaffold/gox-apps/libs/notification/pkg/provider"
	"gorm.io/gorm"
)

// Service returns the running AppService, or nil before the module has been
// declared. A host process that composes its own services uses it to deliver a
// notification directly (see Deliver) instead of publishing an event.
func Service() *AppService { return notificationAppSvc }

// NewAppService builds an AppService over db without the DI container, for a
// process that composes its own services (and for tests).
func NewAppService(db *gorm.DB, p *provider.ProviderService) *AppService {
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	logs := &log.LogEntity{Entity: (&sql.Entity[log.Log]{}).With(q)}
	msgs := &message.MessageEntity{Entity: (&sql.Entity[message.Message]{}).With(q)}
	rules := &rule.RuleEntity{Entity: (&sql.Entity[rule.Rule]{}).With(q)}
	logs.OnRegister()
	msgs.OnRegister()
	rules.OnRegister()
	return &AppService{logEntity: logs, messageEntity: msgs, ruleEntity: rules, provider: p}
}

// Deliver sends one notification now and says whether it went out. It takes
// the same payload as the apps.notification.message.new event (reference, key,
// userId, subject, channels, data, ...), but unlike the event it reports the
// outcome: an error means the message was not delivered and may be delivered
// again.
//
// The reference makes it safe to call again. A reference that was delivered is
// not sent twice; one that failed is sent again on the same Log row, and its
// in-app message is not duplicated.
func (s *AppService) Deliver(p map[string]any) error {
	ref, _ := p["reference"].(string)
	key, _ := p["key"].(string)
	user, _ := p["userId"].(string)
	if ref == "" || key == "" || user == "" {
		return errors.New("notification: reference, key and userId are required")
	}
	if s.provider == nil {
		return errors.New("notification: no provider configured")
	}
	l, err := s.logEntity.First(`"reference" = ?`, ref)
	if err != nil || l == nil {
		if l, err = s.persist(p); err != nil {
			return err
		}
	}
	if l.Status != nil && *l.Status == statusSuccess {
		return nil
	}
	s.provider.Send(l)
	after, err := s.logEntity.First(`"id" = ?`, l.Id)
	if err != nil || after == nil {
		return fmt.Errorf("notification: could not confirm delivery of %s", ref)
	}
	if after.Status == nil || *after.Status != statusSuccess {
		got := "not sent"
		if after.Status != nil {
			got = *after.Status
		}
		return fmt.Errorf("notification: %s was %s", ref, got)
	}
	return nil
}

// statusSuccess is the Log status ProviderService.Send records when every
// channel delivered.
const statusSuccess = "success"
