package charge

import (
	"errors"
	"time"
)

var (
	ErrMemberInvalid   = errors.New("member does not exist")
	ErrFeeTypeInvalid  = errors.New("fee type does not exist or is inactive")
	ErrDuplicateCharge = errors.New("member already has an active charge for this fee type")
	ErrBadDates        = errors.New("end date must be after start date")
)

type MemberChecker interface {
	Exists(id int64) (bool, error)
}

type FeeTypeChecker interface {
	ExistsActive(id int64) (bool, error)
}

type Service struct {
	repo     *Repository
	members  MemberChecker
	feeTypes FeeTypeChecker
}

func NewService(repo *Repository, members MemberChecker, feeTypes FeeTypeChecker) *Service {
	return &Service{repo: repo, members: members, feeTypes: feeTypes}
}

func (s *Service) Create(memberID, feeTypeID, amountCents int64, startDate time.Time, endDate *time.Time) (*MonthlyCharge, error) {
	ok, err := s.members.Exists(memberID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrMemberInvalid
	}

	ok, err = s.feeTypes.ExistsActive(feeTypeID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFeeTypeInvalid
	}

	if endDate != nil && !endDate.After(startDate) {
		return nil, ErrBadDates
	}

	_, err = s.repo.FindActive(memberID, feeTypeID)
	if err == nil {
		return nil, ErrDuplicateCharge
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	mc := &MonthlyCharge{
		MemberID:    memberID,
		FeeTypeID:   feeTypeID,
		AmountCents: amountCents,
		StartDate:   startDate,
		EndDate:     endDate,
	}
	if err := s.repo.Create(mc); err != nil {
		return nil, err
	}
	return s.repo.FindByID(mc.ID)
}

func (s *Service) GetByID(id int64) (*MonthlyCharge, error) {
	return s.repo.FindByID(id)
}

func (s *Service) List(memberID, feeTypeID int64, page, limit int) ([]MonthlyCharge, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(memberID, feeTypeID, limit, offset)
}

func (s *Service) Update(id, amountCents int64, endDate *time.Time) (*MonthlyCharge, error) {
	mc, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if endDate != nil && !endDate.After(mc.StartDate) {
		return nil, ErrBadDates
	}

	mc.AmountCents = amountCents
	mc.EndDate = endDate
	if err := s.repo.Update(mc); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
