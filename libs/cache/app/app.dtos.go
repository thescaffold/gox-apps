package app

type HealthDto struct{}

type GetDto struct {
	Key   string `query:"key"   binding:"required"`
	Group string `query:"group"`
}

type SetDto struct {
	Key   string `json:"key"   binding:"required"`
	Value any    `json:"value"`
	Group string `json:"group"`
	TTL   int64  `json:"ttl"` // seconds; 0 = no expiry
}

type DelDto struct {
	Key   string `query:"key"   binding:"required"`
	Group string `query:"group"`
}

type FlushDto struct {
	Group string `query:"group"`
}
