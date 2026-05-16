// Package providers ports ntx-apps/libs/capital/src/api/provider/providers/.
// Each file here implements payment.Provider against a real third-party
// payment rail. Constructors read credentials from env and tolerate empty
// values — missing keys cause HTTP calls to fail predictably rather than
// panic at boot.
package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	capitalpayment "github.com/thescaffold/gox-apps/libs/capital/app/payment"
	capitalprovider "github.com/thescaffold/gox-apps/libs/capital/app/provider"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// Paystack mirrors ntx-apps/libs/capital/src/api/provider/providers/paystack/paystack.service.ts.
type Paystack struct {
	baseURL   string
	secretKey string
	client    *http.Client
}

// NewPaystack reads PSK_BASE_URL / PSK_SECRET_KEY from env. Defaults match
// the TS service.
func NewPaystack() *Paystack {
	base := os.Getenv("PSK_BASE_URL")
	if base == "" {
		base = "https://api.paystack.co"
	}
	return &Paystack{
		baseURL:   base,
		secretKey: os.Getenv("PSK_SECRET_KEY"),
		client:    &http.Client{},
	}
}

func (p *Paystack) Name() string { return "paystack" }

// Init mirrors TS — returns a Paystack-prefixed reference for the
// client-side checkout (no server-side initialise call required).
func (p *Paystack) Init(firstName, lastName, email, userID, clientID, workspaceID string, amount float64, currency string) (map[string]any, error) {
	return map[string]any{"reference": utils.Reference("PSK", 36)}, nil
}

// Verify mirrors TS — calls GET /transaction/verify/:reference, validates the
// authorization is reusable, and returns the card-link payload on success.
func (p *Paystack) Verify(userID, clientID, workspaceID, reference string, amount float64, currency string) (*capitalpayment.ProviderResponse, error) {
	resp, err := p.do("GET", "/transaction/verify/"+reference, nil)
	if err != nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Charge verification request failed"}, nil
	}
	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Unable to verify your payment"}, nil
	}
	gotRef, _ := data["reference"].(string)
	status, _ := data["status"].(string)
	gotAmount, _ := data["amount"].(float64)
	gotCurrency, _ := data["currency"].(string)
	if gotRef == "" || status == "" || gotAmount == 0 || gotCurrency == "" {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Unable to verify your payment"}, nil
	}
	if gotRef != reference {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Request and response has different transaction reference"}, nil
	}
	if status != "success" {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Payment verification was not successful"}, nil
	}
	if gotAmount < amount {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "The amount paid is less than the expected amount"}, nil
	}
	authz, _ := data["authorization"].(map[string]any)
	reusable, _ := authz["reusable"].(bool)
	if authz == nil || !reusable {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "The provided provider is not reusable for later charge"}, nil
	}
	customer, _ := data["customer"].(map[string]any)
	signature, _ := authz["signature"].(string)
	authCode, _ := authz["authorization_code"].(string)
	cardCountry, _ := authz["country_code"].(string)
	custEmail, _ := customer["email"].(string)

	return &capitalpayment.ProviderResponse{
		Status: true, Title: "Payment", Message: "Payment successfully verified",
		Data: capitalpayment.CardLinkData{
			Type:      "card",
			Signature: signature,
			Email:     custEmail,
			Token:     authCode,
			Country:   cardCountry,
			Meta: map[string]any{
				"first6":  authz["bin"],
				"last4":   authz["last4"],
				"country": cardCountry,
				"type":    authz["provider_type"],
				"expiry":  joinExpiry(authz["exp_month"], authz["exp_year"]),
				"issuer":  authz["bank"],
			},
		},
	}, nil
}

// Charge mirrors TS — POST /transaction/charge_authorization with the saved
// authorization_code, returns paid status.
func (p *Paystack) Charge(provider *capitalprovider.Provider, userID, clientID, workspaceID, reference string, amount float64, currency string) (*capitalpayment.ProviderResponse, error) {
	body := map[string]any{
		"authorization_code": deref(provider.Token),
		"email":              deref(provider.Email),
		"amount":             amount,
		"currency":           provider.Currency,
		"reference":          reference,
	}
	resp, err := p.do("POST", "/transaction/charge_authorization", body)
	if err != nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Provider charge request failed"}, nil
	}
	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Unable to charge provider"}, nil
	}
	status, _ := data["status"].(string)
	if status != "success" {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Provider charge failed"}, nil
	}
	return &capitalpayment.ProviderResponse{
		Status: true, Title: "Payment", Message: "Provider charge successful",
		Data: map[string]any{"reference": reference, "amount": amount},
	}, nil
}

func (p *Paystack) do(method, path string, body map[string]any) (map[string]any, error) {
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, p.baseURL+path, rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("paystack: non-2xx response")
	}
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
