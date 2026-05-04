package app

type HealthDto struct {
	Type string `query:"type"`
}

type FilterByKeyDto struct {
	Key      string  `path:"key"`
	ParentId *string `query:"parentId"`
}

type FindByCodeDto struct {
	Code     string  `path:"code"`
	ParentId *string `query:"parentId"`
}

type FindByValueDto struct {
	Value    string  `path:"value"`
	ParentId *string `query:"parentId"`
}
