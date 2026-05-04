package permissiontype

type CreatePermissionTypeDto struct {
	Name     string  `json:"name"     binding:"required"`
	Key      string  `json:"key"      binding:"required"`
	Desc     *string `json:"desc,omitempty"`
	Resource string  `json:"resource" binding:"required"`
	Action   string  `json:"action"   binding:"required"`
	Scope    string  `json:"scope"    binding:"required"`
	Status   *string `json:"status,omitempty"`
}

type UpdatePermissionTypeDto struct {
	Name     *string `json:"name,omitempty"`
	Key      *string `json:"key,omitempty"`
	Desc     *string `json:"desc,omitempty"`
	Resource *string `json:"resource,omitempty"`
	Action   *string `json:"action,omitempty"`
	Scope    *string `json:"scope,omitempty"`
	Status   *string `json:"status,omitempty"`
}
