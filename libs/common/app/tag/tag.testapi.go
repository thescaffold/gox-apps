package tag

// NewTagServiceForTest is a TagService over entity, without the DI container.
func NewTagServiceForTest(entity *TagEntity) *TagService { return &TagService{entity: entity} }
