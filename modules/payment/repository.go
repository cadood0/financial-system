package payment

import (
	"database/sql"
	"errors"
)

var (
	ErrNotFound    = errors.New("payment not found")
	ErrOverpayment = errors.New("payment exceeds remaining balance for this period")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const selectColumns = `
	p.id, p.member_id, m.full_name, p.charge_id, ft.name,
	p.period, p.amount_cents, p.note, p.recorded_by, u.name, p.created_at`

const selectJoins = `
	FROM payments p
	JOIN members m ON m.id = p.member_id
	JOIN monthly_charges mc ON mc.id = p.charge_id
	JOIN fee_types ft ON ft.id = mc.fee_type_id
	JOIN users u ON u.id = p.recorded_by`

func (r *Repository) RecordPayment(p *Payment, chargeAmountCents int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lockedID int64
	if err := tx.QueryRow(
		`SELECT id FROM monthly_charges WHERE id = $1 FOR UPDATE`, p.ChargeID,
	).Scan(&lockedID); err != nil {
		return err
	}

	var alreadyPaid int64
	if err := tx.QueryRow(
		`SELECT COALESCE(SUM(amount_cents), 0) FROM payments WHERE charge_id = $1 AND period = $2`,
		p.ChargeID, p.Period,
	).Scan(&alreadyPaid); err != nil {
		return err
	}

	if alreadyPaid+p.AmountCents > chargeAmountCents {
		return ErrOverpayment
	}

	if err := tx.QueryRow(
		`INSERT INTO payments (member_id, charge_id, recorded_by, period, amount_cents, note)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		p.MemberID, p.ChargeID, p.RecordedBy, p.Period, p.AmountCents, p.Note,
	).Scan(&p.ID, &p.CreatedAt); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) FindByID(id int64) (*Payment, error) {
	query := `SELECT` + selectColumns + selectJoins + ` WHERE p.id = $1`

	var p Payment
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.MemberID, &p.MemberName, &p.ChargeID, &p.FeeTypeName,
		&p.Period, &p.AmountCents, &p.Note, &p.RecordedBy, &p.RecordedByName, &p.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) List(memberID, chargeID int64, limit, offset int) ([]Payment, int, error) {
	var total int
	countQuery := `
		SELECT COUNT(*) FROM payments p
		WHERE ($1 = 0 OR p.member_id = $1)
		  AND ($2 = 0 OR p.charge_id = $2)`
	if err := r.db.QueryRow(countQuery, memberID, chargeID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT` + selectColumns + selectJoins + `
		WHERE ($1 = 0 OR p.member_id = $1)
		  AND ($2 = 0 OR p.charge_id = $2)
		ORDER BY p.id DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(query, memberID, chargeID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	payments := []Payment{}
	for rows.Next() {
		var p Payment
		if err := rows.Scan(
			&p.ID, &p.MemberID, &p.MemberName, &p.ChargeID, &p.FeeTypeName,
			&p.Period, &p.AmountCents, &p.Note, &p.RecordedBy, &p.RecordedByName, &p.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		payments = append(payments, p)
	}
	return payments, total, rows.Err()
}

func (r *Repository) SumForPeriod(chargeID int64, period interface{}) (int64, error) {
	var paid int64
	err := r.db.QueryRow(
		`SELECT COALESCE(SUM(amount_cents), 0) FROM payments WHERE charge_id = $1 AND period = $2`,
		chargeID, period,
	).Scan(&paid)
	return paid, err
}
