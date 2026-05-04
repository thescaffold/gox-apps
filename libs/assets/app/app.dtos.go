package app

type HealthDto struct {
	Type string `query:"type"`
}

type DynamicFileDto struct {
	Id      *string `query:"id"`
	Variant *string `query:"variant"`
	Name    *string `query:"name"`
	Colors  *string `query:"colors"`
	Square  bool    `query:"square"`
	Size    int     `query:"size"`
	Raw     bool    `query:"raw"`
	Store   bool    `query:"store"`
}
