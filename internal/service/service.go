package service

import "infotecs/internal/repository"

type DatabaseService struct {
	Repository repository.Repository
}

func (s *DatabaseService) InjectRepository(repository repository.Repository) {
	s.Repository = repository
}

var Service = &DatabaseService{}
