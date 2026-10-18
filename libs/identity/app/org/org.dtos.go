package org

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

type Ctx struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

type TeamReq struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	Name        string            `json:"name"`
	Kind        string            `json:"kind"`
	Description string            `json:"description"`
	ParentID    string            `json:"parentId"`
}

type TeamUpdateReq struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	Id          string            `param:"id"`
	Name        *string           `json:"name"`
	Description *string           `json:"description"`
	ParentID    *string           `json:"parentId"`
}

type IDReq struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
	Id  string            `param:"id"`
}

type MemberReq struct {
	Ctx    ntxctx.NTXContext `context:"ntx"`
	Id     string            `param:"id"`
	UserID string            `json:"userId"`
	Role   string            `json:"role"`
}

type MemberRemoveReq struct {
	Ctx    ntxctx.NTXContext `context:"ntx"`
	Id     string            `param:"id"`
	UserID string            `param:"userId"`
}

type RoleReq struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Permissions []string          `json:"permissions"`
}

type RoleUpdateReq struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	Id          string            `param:"id"`
	Name        *string           `json:"name"`
	Description *string           `json:"description"`
	Permissions *[]string         `json:"permissions"`
}

type AssignReq struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	Id          string            `param:"id"`
	SubjectType string            `json:"subjectType"`
	SubjectID   string            `json:"subjectId"`
}

type UnassignReq struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	Id          string            `param:"id"`
	SubjectType string            `param:"subjectType"`
	SubjectID   string            `param:"subjectId"`
}
