package service

type repository interface {
}

type Service struct {
	repo repository
}

func New(repo repository) *Service {

	return &Service{
		repo: repo,
	}
}

func (s *Service) Run() {
}
