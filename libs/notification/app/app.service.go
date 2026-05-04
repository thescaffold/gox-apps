package app

type AppService struct{}

func (s *AppService) GetHello() string { return "Hello World!" }
