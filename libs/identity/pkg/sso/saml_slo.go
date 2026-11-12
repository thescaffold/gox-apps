package sso

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
)

// SAML single logout (V4-03), HTTP-Redirect binding both ways, every message
// signed over its query string (saml-bindings-2.0 §3.4.4.1), as the profile
// requires:
//
//	SP-initiated: SAMLLogoutRequest sends the browser to the IdP's SLO service;
//	the IdP ends its session and redirects back with a LogoutResponse, which
//	SAMLLogoutResponseVerify checks against the request it answers.
//
//	IdP-initiated: the IdP sends the browser to this side's SLO URL with a
//	LogoutRequest; SAMLLogoutRequestVerify checks it (signed by the IdP, for
//	this SLO URL, recent, never seen before) and says whose session to end;
//	SAMLLogoutResponse answers it.
//
// IdP-initiated sign-in stays refused (D-27): only logout may start at the IdP.

var (
	// ErrSAMLNoLogout: the IdP offers no single logout, or this side holds no key to sign with.
	ErrSAMLNoLogout = errors.New("sso: single logout is not available for this connection")
	// ErrSAMLLogout: a logout message that is not valid (unsigned, another key, another SP, old, replayed).
	ErrSAMLLogout = errors.New("sso: the SAML logout message is not valid")
)

// SAMLLogout is a verified IdP-initiated LogoutRequest.
type SAMLLogout struct {
	ID             string
	NameID         string
	SessionIndexes []string
	RelayState     string
}

func (c *Client) canLogout(sp SAMLServiceProvider, idp *SAMLIdP) error {
	if idp == nil || idp.SLOURL == "" || sp.Key == nil || sp.Cert == nil || sp.SLOURL == "" {
		return ErrSAMLNoLogout
	}
	return nil
}

// SAMLLogoutRequest builds the signed redirect that asks the IdP to end the
// person's session (nameID, and sessionIndex when known), and returns the
// request's id: the IdP's LogoutResponse must answer it.
func (c *Client) SAMLLogoutRequest(sp SAMLServiceProvider, idp *SAMLIdP, nameID, sessionIndex, relayState string) (redirect, requestID string, err error) {
	if err := c.canLogout(sp, idp); err != nil {
		return "", "", err
	}
	s, err := c.serviceProvider(sp, idp, nil, true)
	if err != nil {
		return "", "", err
	}
	req, err := s.MakeLogoutRequest(idp.SLOURL, nameID, saml.WithSessionIndexes(sessionIndex))
	if err != nil {
		return "", "", err
	}
	u, err := s.RedirectLogoutRequestURL(req, relayState)
	if err != nil {
		return "", "", err
	}
	return u.String(), req.ID, nil
}

func redirectRequest(rawQuery string) *http.Request {
	return &http.Request{Method: http.MethodGet, URL: &url.URL{RawQuery: rawQuery}, Header: http.Header{}}
}

// SAMLLogoutResponseVerify checks the IdP's LogoutResponse (the raw query
// string it was redirected with): signed by the IdP, for this side's SLO URL,
// answering one of requestIDs, with a Success status.
func (c *Client) SAMLLogoutResponseVerify(sp SAMLServiceProvider, idp *SAMLIdP, rawQuery string, requestIDs []string) error {
	if err := c.canLogout(sp, idp); err != nil {
		return err
	}
	s, err := c.serviceProvider(sp, idp, nil)
	if err != nil {
		return err
	}
	s.RejectSHA1Signatures = true
	if _, err := s.ParseLogoutResponse(redirectRequest(rawQuery), requestIDs); err != nil {
		return fmt.Errorf("%w: %v", ErrSAMLLogout, reason(err))
	}
	return nil
}

// SAMLLogoutRequestVerify checks an IdP-initiated LogoutRequest (the raw
// query string): signed by the IdP's key, addressed to this side's SLO URL,
// issued by the IdP within the allowed delay, and seen for the first time
// (replay records its id; nil turns that off, never in production).
func (c *Client) SAMLLogoutRequestVerify(sp SAMLServiceProvider, idp *SAMLIdP, rawQuery string, replay saml.AssertionReplayCache) (*SAMLLogout, error) {
	if err := c.canLogout(sp, idp); err != nil {
		return nil, err
	}
	s, err := c.serviceProvider(sp, idp, replay)
	if err != nil {
		return nil, err
	}
	r := redirectRequest(rawQuery)
	if r.URL.Query().Get("Signature") == "" {
		// The binding also allows an enveloped XML signature inside the message;
		// this side asks for the query string signature only.
		return nil, fmt.Errorf("%w: not signed", ErrSAMLLogout)
	}
	lr, err := s.ValidateLogoutRequest(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSAMLLogout, reason(err))
	}
	out := &SAMLLogout{ID: lr.ID, RelayState: r.URL.Query().Get("RelayState")}
	if lr.NameID != nil {
		out.NameID = lr.NameID.Value
	}
	for _, si := range lr.AllSessionIndexes() {
		if si.Value != "" {
			out.SessionIndexes = append(out.SessionIndexes, si.Value)
		}
	}
	if out.NameID == "" {
		return nil, fmt.Errorf("%w: no NameID", ErrSAMLLogout)
	}
	return out, nil
}

// SAMLLogoutResponse answers the IdP's LogoutRequest requestID: the signed
// redirect to the IdP's SLO service, with Success when this side's sessions
// ended, Responder otherwise.
func (c *Client) SAMLLogoutResponse(sp SAMLServiceProvider, idp *SAMLIdP, requestID, relayState string, ended bool) (string, error) {
	if err := c.canLogout(sp, idp); err != nil {
		return "", err
	}
	s, err := c.serviceProvider(sp, idp, nil, true)
	if err != nil {
		return "", err
	}
	status := saml.Status{StatusCode: saml.StatusCode{Value: saml.StatusSuccess}}
	if !ended {
		status = saml.Status{StatusCode: saml.StatusCode{Value: saml.StatusResponder}}
	}
	u, err := s.MakeRedirectLogoutResponseWithStatus(requestID, relayState, status)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// reason is the library's private reason for refusing a message.
func reason(err error) error {
	var ire *saml.InvalidResponseError
	if errors.As(err, &ire) && ire.PrivateErr != nil {
		return ire.PrivateErr
	}
	return err
}
