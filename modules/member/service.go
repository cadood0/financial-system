package member

import "errors"

var (
	ErrPhoneTaken  = errors.New("phone already registered")
	ErrCityInvalid = errors.New("city does not exist")
)

type CityChecker interface {
	Exists(id int64) (bool, error)
}

type Service struct {
	repo   *Repository
	cities CityChecker
}

func NewService(repo *Repository, cities CityChecker) *Service {
	return &Service{repo: repo, cities: cities}
}

func (s *Service) Create(fullName, phone string, cityID int64) (*Member, error) {
	ok, err := s.cities.Exists(cityID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrCityInvalid
	}

	_, err = s.repo.FindByPhone(phone)
	if err == nil {
		return nil, ErrPhoneTaken
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	m := &Member{FullName: fullName, Phone: phone, CityID: cityID}
	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return s.repo.FindByID(m.ID)
}

func (s *Service) GetByID(id int64) (*Member, error) {
	return s.repo.FindByID(id)
}

func (s *Service) List(search string, cityID int64, page, limit int) ([]Member, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(search, cityID, limit, offset)
}

func (s *Service) Update(id int64, fullName, phone string, cityID int64) (*Member, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	ok, err := s.cities.Exists(cityID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrCityInvalid
	}

	if phone != m.Phone {
		existing, err := s.repo.FindByPhone(phone)
		if err == nil && existing.ID != id {
			return nil, ErrPhoneTaken
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	m.FullName = fullName
	m.Phone = phone
	m.CityID = cityID
	if err := s.repo.Update(m); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
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

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
