package feeTypes

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("fee type not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(f *FeeType) error {
	query := `
		INSERT INTO fee_types (name, description, is_active)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(
		query,
		f.Name,
		f.Description,
		f.IsActive,
	).Scan(
		&f.ID,
		&f.CreatedAt,
		&f.UpdatedAt,
	)
}

func (r *Repository) FindByID(id int64) (*FeeType, error) {
	query := `
		SELECT
			id,
			name,
			description,
			is_active,
			created_at,
			updated_at
		FROM fee_types
		WHERE id = $1
		  AND deleted_at IS NULL`

	var f FeeType

	err := r.db.QueryRow(query, id).Scan(
		&f.ID,
		&f.Name,
		&f.Description,
		&f.IsActive,
		&f.CreatedAt,
		&f.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &f, nil
}

func (r *Repository) FindByName(name string) (*FeeType, error) {
	query := `
		SELECT
			id,
			name,
			description,
			is_active,
			created_at,
			updated_at
		FROM fee_types
		WHERE name = $1
		  AND deleted_at IS NULL`

	var f FeeType

	err := r.db.QueryRow(query, name).Scan(
		&f.ID,
		&f.Name,
		&f.Description,
		&f.IsActive,
		&f.CreatedAt,
		&f.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &f, nil
}

func (r *Repository) List(search string, limit, offset int) ([]FeeType, int, error) {
	var total int

	countQuery := `
		SELECT COUNT(*)
		FROM fee_types
		WHERE deleted_at IS NULL
		AND (
			$1 = ''
			OR name ILIKE '%' || $1 || '%'
			OR description ILIKE '%' || $1 || '%'
		)`

	if err := r.db.QueryRow(countQuery, search).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			id,
			name,
			description,
			is_active,
			created_at,
			updated_at
		FROM fee_types
		WHERE deleted_at IS NULL
		AND (
			$1 = ''
			OR name ILIKE '%' || $1 || '%'
			OR description ILIKE '%' || $1 || '%'
		)
		ORDER BY name
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	feeTypes := []FeeType{}

	for rows.Next() {
		var f FeeType

		if err := rows.Scan(
			&f.ID,
			&f.Name,
			&f.Description,
			&f.IsActive,
			&f.CreatedAt,
			&f.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		feeTypes = append(feeTypes, f)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return feeTypes, total, nil
}

func (r *Repository) Update(f *FeeType) error {
	query := `
		UPDATE fee_types
		SET
			name = $1,
			description = $2,
			is_active = $3,
			updated_at = NOW()
		WHERE id = $4
		  AND deleted_at IS NULL
		RETURNING updated_at`

	err := r.db.QueryRow(
		query,
		f.Name,
		f.Description,
		f.IsActive,
		f.ID,
	).Scan(&f.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	return err
}

func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec(
		`UPDATE fee_types
		 SET deleted_at = NOW()
		 WHERE id = $1
		   AND deleted_at IS NULL`,
		id,
	)

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
