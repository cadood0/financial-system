package dashboard

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Summary() (*Summary, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM members WHERE deleted_at IS NULL),
			(SELECT COUNT(*) FROM cities WHERE deleted_at IS NULL),
			(SELECT COALESCE(SUM(amount_cents), 0) FROM payments),
			(SELECT COALESCE(SUM(mc.amount_cents), 0)
			   FROM monthly_charges mc
			   CROSS JOIN LATERAL generate_series(
			       date_trunc('month', mc.start_date::timestamp),
			       LEAST(
			           date_trunc('month', NOW()::timestamp),
			           COALESCE(date_trunc('month', mc.end_date::timestamp),
			                    date_trunc('month', NOW()::timestamp))
			       ),
			       interval '1 month'
			   ) AS gs(month)
			   WHERE mc.deleted_at IS NULL),
			(SELECT COALESCE(SUM(amount_cents), 0)
			   FROM payments
			   WHERE period <= date_trunc('month', NOW())::date)`

	var s Summary
	var paidToDate int64
	err := r.db.QueryRow(query).Scan(
		&s.TotalMembers,
		&s.TotalCities,
		&s.TotalCollectedCents,
		&s.ExpectedToDateCents,
		&paidToDate,
	)
	if err != nil {
		return nil, err
	}
	s.OutstandingCents = s.ExpectedToDateCents - paidToDate
	if s.OutstandingCents < 0 {
		s.OutstandingCents = 0
	}
	return &s, nil
}

func (r *Repository) MonthlyRevenue(months int) ([]MonthlyRevenue, error) {
	query := `
		SELECT date_trunc('month', created_at) AS month,
		       SUM(amount_cents)
		FROM payments
		WHERE created_at >= date_trunc('month', NOW()) - make_interval(months => $1 - 1)
		GROUP BY month
		ORDER BY month`

	rows, err := r.db.Query(query, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []MonthlyRevenue{}
	for rows.Next() {
		var m MonthlyRevenue
		if err := rows.Scan(&m.Month, &m.CollectedCents); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *Repository) RevenueByCity() ([]CityRevenue, error) {
	query := `
		SELECT c.id, c.name, COALESCE(SUM(p.amount_cents), 0)
		FROM cities c
		LEFT JOIN members m ON m.city_id = c.id AND m.deleted_at IS NULL
		LEFT JOIN payments p ON p.member_id = m.id
		WHERE c.deleted_at IS NULL
		GROUP BY c.id, c.name
		ORDER BY 3 DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []CityRevenue{}
	for rows.Next() {
		var cr CityRevenue
		if err := rows.Scan(&cr.CityID, &cr.CityName, &cr.CollectedCents); err != nil {
			return nil, err
		}
		result = append(result, cr)
	}
	return result, rows.Err()
}

func (r *Repository) RevenueByFeeType() ([]FeeTypeRevenue, error) {
	query := `
		SELECT ft.id, ft.name, COALESCE(SUM(p.amount_cents), 0)
		FROM fee_types ft
		LEFT JOIN monthly_charges mc ON mc.fee_type_id = ft.id
		LEFT JOIN payments p ON p.charge_id = mc.id
		WHERE ft.deleted_at IS NULL
		GROUP BY ft.id, ft.name
		ORDER BY 3 DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []FeeTypeRevenue{}
	for rows.Next() {
		var fr FeeTypeRevenue
		if err := rows.Scan(&fr.FeeTypeID, &fr.FeeTypeName, &fr.CollectedCents); err != nil {
			return nil, err
		}
		result = append(result, fr)
	}
	return result, rows.Err()
}
