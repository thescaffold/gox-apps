package apikey

import "github.com/thescaffold/gox-apps/libs/identity/pkg/apikey"

// NewControllerForTest builds a controller over a key service without the DI container.
func NewControllerForTest(svc *apikey.Service) *ApiKeyController {
	s := &ApiKeyService{}
	s.Use(svc)
	return &ApiKeyController{svc: s}
}
