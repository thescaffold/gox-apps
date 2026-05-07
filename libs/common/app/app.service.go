package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/thescaffold/gox-apps-common/app/ip"
	"github.com/thescaffold/gox-apps-common/app/rate"
)

var geoClient = &http.Client{Timeout: 5 * time.Second}

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

	entry := &ip.Ip{Value: ipValue, Country: "US"}

	if result, lookupErr := geoLookup(ipValue); lookupErr == nil {
		entry.Country = result.Country
		if result.City != "" {
			entry.City = &result.City
		}
	}

	if err := s.ipEntity.Insert(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

type geoResult struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

func geoLookup(ipValue string) (*geoResult, error) {
	base := os.Getenv("GEO_API_URL")
	if base == "" {
		base = "http://ip-api.com/json"
	}
	resp, err := geoClient.Get(fmt.Sprintf("%s/%s?fields=country,city", base, ipValue))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r geoResult
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}
