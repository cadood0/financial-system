package feetype

import "errors"

var (
	ErrNameTaken = errors.New("fee type name already exists")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(name, description string, isActive bool) (*FeeType, error) {
	_, err := s.repo.FindByName(name)
	if err == nil {
		return nil, ErrNameTaken
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	f := &FeeType{
		Name:        name,
		Description: description,
		IsActive:    isActive,
	}

	if err := s.repo.Create(f); err != nil {
		return nil, err
	}

	return s.repo.FindByID(f.ID)
}

func (s *Service) GetByID(id int64) (*FeeType, error) {
	return s.repo.FindByID(id)
}

func (s *Service) List(search string, page, limit int) ([]FeeType, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(search, limit, offset)
}

func (s *Service) Update(id int64, name, description string, isActive bool) (*FeeType, error) {
	f, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if f.Name != name {
		existing, err := s.repo.FindByName(name)
		if err == nil && existing.ID != id {
			return nil, ErrNameTaken
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	f.Name = name
	f.Description = description
	f.IsActive = isActive

	if err := s.repo.Update(f); err != nil {
		return nil, err
	}

	return s.repo.FindByID(id)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
