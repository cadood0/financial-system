package charge

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("monthly charge not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const selectColumns = `
	mc.id, mc.member_id, m.full_name, mc.fee_type_id, ft.name,
	mc.amount_cents, mc.start_date, mc.end_date, mc.created_at, mc.updated_at`

const selectJoins = `
	FROM monthly_charges mc
	JOIN members m ON m.id = mc.member_id
	JOIN fee_types ft ON ft.id = mc.fee_type_id`

func (r *Repository) scanRow(row *sql.Row) (*MonthlyCharge, error) {
	var mc MonthlyCharge
	err := row.Scan(
		&mc.ID, &mc.MemberID, &mc.MemberName, &mc.FeeTypeID, &mc.FeeTypeName,
		&mc.AmountCents, &mc.StartDate, &mc.EndDate, &mc.CreatedAt, &mc.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &mc, nil
}

func (r *Repository) Create(mc *MonthlyCharge) error {
	query := `
		INSERT INTO monthly_charges (member_id, fee_type_id, amount_cents, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(query,
		mc.MemberID, mc.FeeTypeID, mc.AmountCents, mc.StartDate, mc.EndDate,
	).Scan(&mc.ID, &mc.CreatedAt, &mc.UpdatedAt)
}

func (r *Repository) FindByID(id int64) (*MonthlyCharge, error) {
	query := `SELECT` + selectColumns + selectJoins + `
		WHERE mc.id = $1 AND mc.deleted_at IS NULL`
	return r.scanRow(r.db.QueryRow(query, id))
}

func (r *Repository) FindActive(memberID, feeTypeID int64) (*MonthlyCharge, error) {
	query := `SELECT` + selectColumns + selectJoins + `
		WHERE mc.member_id = $1 AND mc.fee_type_id = $2 AND mc.deleted_at IS NULL`
	return r.scanRow(r.db.QueryRow(query, memberID, feeTypeID))
}

func (r *Repository) List(memberID, feeTypeID int64, limit, offset int) ([]MonthlyCharge, int, error) {
	var total int
	countQuery := `
		SELECT COUNT(*) FROM monthly_charges mc
		WHERE mc.deleted_at IS NULL
		  AND ($1 = 0 OR mc.member_id = $1)
		  AND ($2 = 0 OR mc.fee_type_id = $2)`
	if err := r.db.QueryRow(countQuery, memberID, feeTypeID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT` + selectColumns + selectJoins + `
		WHERE mc.deleted_at IS NULL
		  AND ($1 = 0 OR mc.member_id = $1)
		  AND ($2 = 0 OR mc.fee_type_id = $2)
		ORDER BY mc.id
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(query, memberID, feeTypeID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	charges := []MonthlyCharge{}
	for rows.Next() {
		var mc MonthlyCharge
		if err := rows.Scan(
			&mc.ID, &mc.MemberID, &mc.MemberName, &mc.FeeTypeID, &mc.FeeTypeName,
			&mc.AmountCents, &mc.StartDate, &mc.EndDate, &mc.CreatedAt, &mc.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		charges = append(charges, mc)
	}
	return charges, total, rows.Err()
}

func (r *Repository) Update(mc *MonthlyCharge) error {
	query := `
		UPDATE monthly_charges
		SET amount_cents = $1, end_date = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING updated_at`
	err := r.db.QueryRow(query, mc.AmountCents, mc.EndDate, mc.ID).Scan(&mc.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec(
		`UPDATE monthly_charges SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
