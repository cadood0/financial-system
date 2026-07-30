package city

import (
	"errors"
)

var ErrNameTaken = errors.New("name already Created")
var ErrCityHasMembers = errors.New("city still has active members")

type Service struct {
	repo    *Repository
	members MemberCounter
}

type MemberCounter interface {
	CountActiveByCity(cityID int64) (int, error)
}

func NewService(repo *Repository, members MemberCounter) *Service {
	return &Service{repo: repo, members: members}
}

func (s *Service) Create(name, region string) (*City, error) {
	_, err := s.repo.FindByName(name)
	if err == nil {
		return nil, ErrNameTaken
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	ct := &City{Name: name, Region: region}
	if err := s.repo.Create(ct); err != nil {
		return nil, err
	}
	return ct, nil
}

func (s *Service) GetByID(id int64) (*City, error) {
	return s.repo.FindByID(id)
}

func (s *Service) List(search string, page, limit int) ([]City, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(search, limit, offset)
}

func (s *Service) Update(id int64, name, region string) (*City, error) {
	ct, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if name != ct.Name {
		existing, err := s.repo.FindByName(name)
		if err == nil && existing.ID != id {
			return nil, ErrNameTaken
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	ct.Name = name
	ct.Region = region
	if err := s.repo.Update(ct); err != nil {
		return nil, err
	}
	return ct, nil
}

func (s *Service) Delete(id int64) (map[string]any, error) {
	count, err := s.members.CountActiveByCity(id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrCityHasMembers
	}

	err = s.repo.Delete(id)
	if err != nil {
		return nil, err
	}

	// Build the JSON response object here
	response := map[string]any{
		"message": "city successfully deleted",
		"id":      id,
	}
	return response, nil
}

func (s *Service) Exists(id int64) (bool, error) {
	_, err := s.repo.FindByID(id)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
