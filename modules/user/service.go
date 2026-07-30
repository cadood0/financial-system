package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmailTaken = errors.New("email already registered")

var ErrSelfDelete = errors.New("you cannot delete your own account")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(name, email, password string) (*User, error) {
	_, err := s.repo.FindByEmail(email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{Name: name, Email: email, PasswordHash: string(hash)}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (s *Service) Login(email, password string) (*User, error) {
	u, err := s.repo.FindByEmail(email)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

func (s *Service) GetByID(id int64) (*User, error) {
	return s.repo.FindByID(id)
}

func (s *Service) List(search string, page, limit int) ([]User, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(search, limit, offset)
}

func (s *Service) Update(id int64, name, email string) (*User, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if email != u.Email {
		existing, err := s.repo.FindByEmail(email)
		if err == nil && existing.ID != id {
			return nil, ErrEmailTaken
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	u.Name = name
	u.Email = email
	if err := s.repo.Update(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Delete(requesterID, targetID int64) error {
	if requesterID == targetID {
		return ErrSelfDelete
	}
	return s.repo.Delete(targetID)
}
