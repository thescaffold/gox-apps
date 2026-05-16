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

// Flutterwave mirrors ntx-apps/libs/capital/src/api/provider/providers/flutterwave/flutterwave.service.ts.
type Flutterwave struct {
	baseURL   string
	secretKey string
	client    *http.Client
}

// NewFlutterwave reads FLW_BASE_URL / FLW_SECRET_KEY from env.
func NewFlutterwave() *Flutterwave {
	base := os.Getenv("FLW_BASE_URL")
	if base == "" {
		base = "https://api.flutterwave.com"
	}
	return &Flutterwave{
		baseURL:   base,
		secretKey: os.Getenv("FLW_SECRET_KEY"),
		client:    &http.Client{},
	}
}

func (f *Flutterwave) Name() string { return "flutterwave" }

// Init mirrors TS — returns a Flutterwave-prefixed tx_ref for client-side use.
func (f *Flutterwave) Init(firstName, lastName, email, userID, clientID, workspaceID string, amount float64, currency string) (map[string]any, error) {
	return map[string]any{"reference": utils.Reference("FLW", 36)}, nil
}

// Verify mirrors TS — GET /v3/transactions/verify_by_reference?tx_ref=…
// validates the transaction matches amount/reference and surfaces a
// card-link payload from card.token.
func (f *Flutterwave) Verify(userID, clientID, workspaceID, reference string, amount float64, currency string) (*capitalpayment.ProviderResponse, error) {
	// TS uses toMajor(amount) which divides by 100 for Naira and similar.
	amountMajor := amount / 100
	resp, err := f.do("GET", "/v3/transactions/verify_by_reference?tx_ref="+reference, nil)
	if err != nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Charge verification request failed"}, nil
	}
	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Unable to verify your payment"}, nil
	}
	txID := numericString(data["id"])
	status, _ := data["status"].(string)
	gotAmount, _ := data["amount"].(float64)
	gotCurrency, _ := data["currency"].(string)
	if txID == "" || status == "" || gotAmount == 0 || gotCurrency == "" {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Unable to verify your payment"}, nil
	}
	if txID != reference {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Request and response has different transaction ID"}, nil
	}
	if status != "successful" {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Payment verification was not successful"}, nil
	}
	if gotAmount < amountMajor {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "The amount paid is less than the expected amount"}, nil
	}
	card, _ := data["card"].(map[string]any)
	customer, _ := data["customer"].(map[string]any)
	token, _ := card["token"].(string)
	cardCountry, _ := card["country"].(string)
	custEmail, _ := customer["email"].(string)

	return &capitalpayment.ProviderResponse{
		Status: true, Title: "Payment", Message: "Payment successfully verified",
		Data: capitalpayment.CardLinkData{
			Type:      "card",
			Signature: token,
			Email:     custEmail,
			Token:     token,
			Country:   cardCountry,
			Meta: map[string]any{
				"first6":  card["first_6digits"],
				"last4":   card["last_4digits"],
				"country": cardCountry,
				"type":    card["type"],
				"expiry":  card["expiry"],
				"issuer":  card["issuer"],
			},
		},
	}, nil
}

// Charge mirrors TS — POST /v3/tokenized-charges with the saved token.
func (f *Flutterwave) Charge(provider *capitalprovider.Provider, userID, clientID, workspaceID, reference string, amount float64, currency string) (*capitalpayment.ProviderResponse, error) {
	amountMajor := amount / 100
	body := map[string]any{
		"token":    deref(provider.Token),
		"email":    deref(provider.Email),
		"amount":   amountMajor,
		"currency": provider.Currency,
		"country":  provider.Country,
		"tx_ref":   reference,
	}
	resp, err := f.do("POST", "/v3/tokenized-charges", body)
	if err != nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Provider charge request failed"}, nil
	}
	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Unable to charge provider"}, nil
	}
	status, _ := data["status"].(string)
	if status != "successful" {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Provider charge failed"}, nil
	}
	return &capitalpayment.ProviderResponse{
		Status: true, Title: "Payment", Message: "Provider charge successful",
		Data: map[string]any{"reference": reference, "amount": amount},
	}, nil
}

func (f *Flutterwave) do(method, path string, body map[string]any) (map[string]any, error) {
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, f.baseURL+path, rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+f.secretKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("flutterwave: non-2xx response")
	}
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
