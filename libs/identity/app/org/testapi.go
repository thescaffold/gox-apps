package org

import orgsvc "github.com/thescaffold/gox-apps/libs/identity/pkg/org"

// NewControllerForTest builds a controller over an organisation service without the DI container.
func NewControllerForTest(svc *orgsvc.Service) *OrgController {
	s := &OrgService{}
	s.Use(svc)
	return &OrgController{svc: s}
}
