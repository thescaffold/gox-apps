package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	capitalpayment "github.com/thescaffold/gox-apps/libs/capital/app/payment"
	capitalprovider "github.com/thescaffold/gox-apps/libs/capital/app/provider"
)

// Stripe mirrors ntx-apps/libs/capital/src/api/provider/providers/stripe/stripe.service.ts.
// Uses Stripe's REST API directly (no SDK dependency) — endpoints accept
// form-encoded bodies and respond with JSON.
type Stripe struct {
	apiKey    string
	groupName string
	appName   string
	client    *http.Client
}

// NewStripe reads STR_SECRET_KEY / GROUP_NAME / APP_NAME from env.
func NewStripe() *Stripe {
	return &Stripe{
		apiKey:    os.Getenv("STR_SECRET_KEY"),
		groupName: os.Getenv("GROUP_NAME"),
		appName:   os.Getenv("APP_NAME"),
		client:    &http.Client{},
	}
}

func (s *Stripe) Name() string { return "stripe" }

// Init mirrors TS — finds-or-creates the Stripe customer, then opens a
// checkout session with off-session card capture for future charges.
func (s *Stripe) Init(firstName, lastName, email, userID, clientID, workspaceID string, amount float64, currency string) (map[string]any, error) {
	if amount < 50 {
		amount = 50 // Stripe minimum.
	}
	customerID, err := s.findOrCreateCustomer(email, firstName, lastName, userID, clientID, workspaceID)
	if err != nil {
		return nil, err
	}
	productName := strings.ToLower(fmt.Sprintf("%s %s subscription", s.groupName, s.appName))
	form := url.Values{}
	form.Set("customer", customerID)
	form.Set("mode", "payment")
	form.Set("ui_mode", "custom")
	form.Set("payment_method_types[0]", "card")
	form.Set("payment_intent_data[setup_future_usage]", "off_session")
	form.Set("adaptive_pricing[enabled]", "true")
	form.Set("line_items[0][price_data][currency]", currency)
	form.Set("line_items[0][price_data][product_data][name]", productName)
	form.Set("line_items[0][price_data][unit_amount]", fmt.Sprintf("%d", int(amount)))
	form.Set("line_items[0][quantity]", "1")
	form.Set("metadata[userId]", userID)
	form.Set("metadata[clientId]", clientID)
	form.Set("metadata[workspaceId]", workspaceID)
	resp, err := s.doForm("POST", "/v1/checkout/sessions", form)
	if err != nil {
		return nil, err
	}
	id, _ := resp["id"].(string)
	secret, _ := resp["client_secret"].(string)
	return map[string]any{"reference": id, "secret": secret}, nil
}

// Verify mirrors TS — retrieves the checkout session, validates status +
// payment_status + amount, then lists the customer's payment methods so
// each card surfaces as a Provider row.
func (s *Stripe) Verify(userID, clientID, workspaceID, reference string, amount float64, currency string) (*capitalpayment.ProviderResponse, error) {
	session, err := s.doForm("GET", "/v1/checkout/sessions/"+reference, nil)
	if err != nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Charge verification request failed"}, nil
	}
	status, _ := session["status"].(string)
	switch status {
	case "expired":
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Charge verification request failed. Payment session expired"}, nil
	case "open":
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Charge verification incomplete, Complete payment or try again"}, nil
	}
	id, _ := session["id"].(string)
	if id != reference {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Charge verification failed. Invalid transaction reference"}, nil
	}
	paymentStatus, _ := session["payment_status"].(string)
	if status != "complete" || paymentStatus != "paid" {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Payment verification was not successful"}, nil
	}
	amountTotal, _ := session["amount_total"].(float64)
	if amountTotal < amount {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "The amount paid is less than the expected amount"}, nil
	}
	customerDetails, _ := session["customer_details"].(map[string]any)
	email, _ := customerDetails["email"].(string)
	custResp, err := s.doForm("GET", "/v1/customers?email="+url.QueryEscape(email)+"&limit=1", nil)
	if err != nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Something went wrong, please try again or contact support"}, nil
	}
	custData, _ := custResp["data"].([]any)
	if len(custData) != 1 {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Something went wrong, please try again or contact support"}, nil
	}
	cust, _ := custData[0].(map[string]any)
	customerID, _ := cust["id"].(string)
	pmResp, err := s.doForm("GET", "/v1/customers/"+customerID+"/payment_methods?type=card&limit=10", nil)
	if err != nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Something went wrong listing cards"}, nil
	}
	pmData, _ := pmResp["data"].([]any)
	sessionCurrency, _ := session["currency"].(string)
	cards := make([]capitalpayment.CardLinkData, 0, len(pmData))
	for _, pm := range pmData {
		m, _ := pm.(map[string]any)
		card, _ := m["card"].(map[string]any)
		if card == nil {
			continue
		}
		fingerprint, _ := card["fingerprint"].(string)
		pmID, _ := m["id"].(string)
		country, _ := card["country"].(string)
		cards = append(cards, capitalpayment.CardLinkData{
			Type:      "card",
			Signature: fingerprint,
			Email:     email,
			Token:     pmID,
			Country:   country,
			Meta: map[string]any{
				"first6":     card["brand"],
				"last4":      card["last4"],
				"country":    country,
				"type":       card["networks"],
				"expiry":     joinExpiry(card["exp_month"], card["exp_year"]),
				"issuer":     card["issuer"],
				"customerId": customerID,
				"currency":   sessionCurrency,
			},
		})
	}
	return &capitalpayment.ProviderResponse{
		Status: true, Title: "Payment", Message: "Payment successfully verified",
		Data: cards,
	}, nil
}

// Charge mirrors TS — creates a confirmed off-session PaymentIntent for the
// stored customer + payment method.
func (s *Stripe) Charge(provider *capitalprovider.Provider, userID, clientID, workspaceID, reference string, amount float64, currency string) (*capitalpayment.ProviderResponse, error) {
	customerID := ""
	if provider.Meta != nil {
		var meta map[string]any
		_ = json.Unmarshal(provider.Meta, &meta)
		customerID, _ = meta["customerId"].(string)
	}
	form := url.Values{}
	form.Set("amount", fmt.Sprintf("%d", int(amount)))
	form.Set("currency", provider.Currency)
	form.Set("customer", customerID)
	form.Set("receipt_email", deref(provider.Email))
	form.Set("payment_method", deref(provider.Token))
	form.Set("off_session", "true")
	form.Set("confirm", "true")
	form.Set("metadata[userId]", userID)
	form.Set("metadata[clientId]", clientID)
	form.Set("metadata[workspaceId]", workspaceID)
	form.Set("metadata[reference]", reference)
	resp, err := s.doForm("POST", "/v1/payment_intents", form)
	if err != nil {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Provider charge request failed"}, nil
	}
	status, _ := resp["status"].(string)
	if status != "succeeded" {
		return &capitalpayment.ProviderResponse{Status: false, Title: "Payment", Message: "Provider charge request failed"}, nil
	}
	return &capitalpayment.ProviderResponse{
		Status: true, Title: "Payment", Message: "Provider charge successful",
		Data: map[string]any{"reference": reference, "amount": amount},
	}, nil
}

func (s *Stripe) findOrCreateCustomer(email, firstName, lastName, userID, clientID, workspaceID string) (string, error) {
	resp, err := s.doForm("GET", "/v1/customers?email="+url.QueryEscape(email)+"&limit=1", nil)
	if err != nil {
		return "", err
	}
	if data, _ := resp["data"].([]any); len(data) == 1 {
		c, _ := data[0].(map[string]any)
		id, _ := c["id"].(string)
		return id, nil
	}
	form := url.Values{}
	form.Set("name", strings.TrimSpace(firstName+" "+lastName))
	form.Set("email", email)
	form.Set("metadata[userId]", userID)
	form.Set("metadata[clientId]", clientID)
	form.Set("metadata[workspaceId]", workspaceID)
	created, err := s.doForm("POST", "/v1/customers", form)
	if err != nil {
		return "", err
	}
	id, _ := created["id"].(string)
	return id, nil
}

func (s *Stripe) doForm(method, path string, form url.Values) (map[string]any, error) {
	var body io.Reader
	if form != nil && method != "GET" {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(method, "https://api.stripe.com"+path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.apiKey, "")
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, errors.New("stripe: " + resp.Status + " " + string(raw))
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
