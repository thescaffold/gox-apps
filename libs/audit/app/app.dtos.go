package app

type HealthDto struct {
	Type string `query:"type"`
}

type ActivitiesDto struct {
	Page    int `query:"page"`
	PerPage int `query:"perPage"`
}
