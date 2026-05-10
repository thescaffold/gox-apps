package list

import "encoding/json"

// CreateListDto mirrors ntx-apps/libs/statics/src/api/list/dto/create-list.dto.ts.
// code is required (TS @IsString without IsOptional).
type CreateListDto struct {
	ParentId *string         `json:"parentId"               form:"parentId"`
	Key      string          `json:"key"   binding:"required" form:"key"`
	Code     string          `json:"code"  binding:"required" form:"code"`
	Value    string          `json:"value" binding:"required" form:"value"`
	Meta     json.RawMessage `json:"meta,omitempty"           form:"meta"`
	Status   *string         `json:"status,omitempty"         form:"status"`
}

type UpdateListDto struct {
	ParentId *string         `json:"parentId,omitempty"  form:"parentId"`
	Key      *string         `json:"key,omitempty"       form:"key"`
	Code     *string         `json:"code,omitempty"      form:"code"`
	Value    *string         `json:"value,omitempty"     form:"value"`
	Meta     json.RawMessage `json:"meta,omitempty"      form:"meta"`
	Status   *string         `json:"status,omitempty"    form:"status"`
}
