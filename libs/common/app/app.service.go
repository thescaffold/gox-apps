package app

import (
	"github.com/thescaffold/gox-apps-common/app/ip"
	"github.com/thescaffold/gox-apps-common/app/rate"
)

type AppService struct {
	rateEntity *rate.RateEntity `inject:""`
	ipEntity   *ip.IpEntity     `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

func (s *AppService) GetCurrency(currency string) (*rate.Rate, error) {
	return s.rateEntity.First(`"currency" = ?`, currency)
}

func (s *AppService) GetLocation(ipValue string) (*ip.Ip, error) {
	existing, err := s.ipEntity.First(`"value" = ?`, ipValue)
	if err == nil && existing != nil {
		return existing, nil
	}
	// stub: store with default country when external lookup is not configured
	entry := &ip.Ip{Value: ipValue, Country: "US"}
	if err := s.ipEntity.Insert(entry); err != nil {
		return nil, err
	}
	return entry, nil
}
