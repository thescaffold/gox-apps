package app

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/thescaffold/gox-apps/libs/common/app/ip"
	"github.com/thescaffold/gox-apps/libs/common/app/rate"
	"github.com/thescaffold/gox-apps/libs/common/app/ratelog"
	gxhttp "github.com/thescaffold/gox-packages/libs/core/http"
)

var geoClient = &http.Client{Timeout: 5 * time.Second}

type AppService struct {
	rateEntity    *rate.RateEntity       `inject:""`
	rateLogEntity *ratelog.RateLogEntity `inject:""`
	ipEntity      *ip.IpEntity           `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

func (s *AppService) GetCurrency(currency string) (*rate.Rate, error) {
	// TS Rate.currency has an upper-casing column transformer applied to the
	// query value, so lookups are case-insensitive against the upper-cased
	// stored values.
	return s.rateEntity.First(`"currency" = ?`, strings.ToUpper(currency))
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

// WeeklyHeartbeat mirrors TS app.controller.ts 'apps.cron.heartbeat.weekly'
// subscription: pulls latest rates from the configured provider (EUR-based,
// re-based to USD), upserts a Rate row per currency, and appends a RateLog
// entry. External call failures are swallowed (the cron just no-ops on the
// next run) — TS behaves the same way via its `quickHttpService` envelope.
func (s *AppService) WeeklyHeartbeat() error {
	rates, err := fetchUSDRates()
	if err != nil || len(rates) == 0 {
		return err
	}
	for currency, value := range rates {
		existing, _ := s.rateEntity.First(`"currency" = ?`, currency)
		delta := value
		if existing != nil && existing.Value != nil {
			delta = value - *existing.Value
		}

		v := value
		d := delta
		if existing == nil {
			r := &rate.Rate{Currency: currency, Value: &v, Delta: &d}
			if err := s.rateEntity.Insert(r); err != nil {
				continue
			}
			existing = r
		} else {
			existing.Value = &v
			existing.Delta = &d
			if _, err := s.rateEntity.Update(existing, "id = ?", existing.Id); err != nil {
				continue
			}
		}

		now := time.Now().UTC()
		_ = s.rateLogEntity.Insert(&ratelog.RateLog{
			RateId: existing.Id,
			Value:  &v,
			Delta:  &d,
			Date:   &now,
		})
	}
	return nil
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

// fetchUSDRates mirrors TS pkg/rate/exchangeratesapi/getRates('USD'): the
// upstream plan only supports EUR as the base, so we request EUR and re-base
// every entry against the USD rate. Matches convertRatesToBase() arithmetic.
func fetchUSDRates() (map[string]int, error) {
	base := os.Getenv("EXCHANGERATESAPI_BASE_URL")
	if base == "" {
		return nil, nil
	}
	key := os.Getenv("EXCHANGERATESAPI_ACCESS_KEY")

	client := gxhttp.New("")
	success, _, _, _, body := client.External(
		"GET",
		fmt.Sprintf("%s/v1/latest", base),
		nil,
		map[string]string{"base": "EUR", "access_key": key},
		map[string]string{"content-type": "application/json"},
		0,
	)
	if !success {
		return nil, nil
	}
	// External returns parsed JSON in `body`. Expected shape:
	// { rates: { USD: 1.07, EUR: 1, ... } }.
	resp, ok := body.(map[string]any)
	if !ok {
		return nil, nil
	}
	ratesRaw, ok := resp["rates"].(map[string]any)
	if !ok {
		return nil, nil
	}
	eurToBase, ok := ratesRaw["USD"].(float64)
	if !ok || eurToBase == 0 {
		return nil, nil
	}
	out := make(map[string]int, len(ratesRaw))
	for currency, v := range ratesRaw {
		f, ok := v.(float64)
		if !ok {
			continue
		}
		// TS divides each rate by eurToBaseRate to re-base, then persists into an
		// int column — Postgres rounds to nearest, so round (don't truncate).
		out[currency] = int(math.Round(f / eurToBase))
	}
	return out, nil
}
