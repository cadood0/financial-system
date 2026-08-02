package payment

import (
	"errors"
	"time"

	"financial-system/modules/charge"
)

var (
	ErrChargeInvalid    = errors.New("monthly charge does not exist")
	ErrPeriodOutOfRange = errors.New("period is outside the charge's active range")
)

type ChargeGetter interface {
	GetByID(id int64) (*charge.MonthlyCharge, error)
}

type Service struct {
	repo    *Repository
	charges ChargeGetter
}

func NewService(repo *Repository, charges ChargeGetter) *Service {
	return &Service{repo: repo, charges: charges}
}

func monthOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func (s *Service) Record(chargeID, amountCents, recordedBy int64, period time.Time, note string) (*Payment, error) {
	ch, err := s.charges.GetByID(chargeID)
	if err != nil {
		if errors.Is(err, charge.ErrNotFound) {
			return nil, ErrChargeInvalid
		}
		return nil, err
	}

	if period.Before(monthOf(ch.StartDate)) {
		return nil, ErrPeriodOutOfRange
	}
	if ch.EndDate != nil && period.After(monthOf(*ch.EndDate)) {
		return nil, ErrPeriodOutOfRange
	}

	p := &Payment{
		MemberID:    ch.MemberID,
		ChargeID:    chargeID,
		RecordedBy:  recordedBy,
		Period:      period,
		AmountCents: amountCents,
		Note:        note,
	}
	if err := s.repo.RecordPayment(p, ch.AmountCents); err != nil {
		return nil, err
	}
	return s.repo.FindByID(p.ID)
}

func (s *Service) GetByID(id int64) (*Payment, error) {
	return s.repo.FindByID(id)
}

func (s *Service) List(memberID, chargeID int64, page, limit int) ([]Payment, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(memberID, chargeID, limit, offset)
}

func (s *Service) Summary(chargeID int64, period time.Time) (*PeriodSummary, error) {
	ch, err := s.charges.GetByID(chargeID)
	if err != nil {
		if errors.Is(err, charge.ErrNotFound) {
			return nil, ErrChargeInvalid
		}
		return nil, err
	}

	paid, err := s.repo.SumForPeriod(chargeID, period)
	if err != nil {
		return nil, err
	}

	status := "unpaid"
	switch {
	case paid >= ch.AmountCents:
		status = "paid"
	case paid > 0:
		status = "partial"
	}

	return &PeriodSummary{
		ChargeID:          chargeID,
		Period:            period,
		ChargeAmountCents: ch.AmountCents,
		PaidCents:         paid,
		RemainingCents:    ch.AmountCents - paid,
		Status:            status,
	}, nil
}
