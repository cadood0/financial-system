package city

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("city not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ct *City) error {
	query := `
		INSERT INTO cities (name, region)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(query, ct.Name, ct.Region).
		Scan(&ct.ID, &ct.CreatedAt, &ct.UpdatedAt)
}

func (r *Repository) FindByName(name string) (*City, error) {
	query := `
		SELECT id, name, region, created_at, updated_at
		FROM cities
		WHERE name = $1 AND deleted_at IS NULL`

	var ct City
	err := r.db.QueryRow(query, name).
		Scan(&ct.ID, &ct.Name, &ct.Region, &ct.CreatedAt, &ct.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

func (r *Repository) FindByID(id int64) (*City, error) {
	query := `
		SELECT id, name, region, created_at, updated_at
		FROM cities
		WHERE id = $1 AND deleted_at IS NULL`

	var ct City
	err := r.db.QueryRow(query, id).
		Scan(&ct.ID, &ct.Name, &ct.Region, &ct.CreatedAt, &ct.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

func (r *Repository) List(search string, limit, offset int) ([]City, int, error) {
	var total int
	countQuery := `
		SELECT COUNT(*) FROM cities
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR region ILIKE '%' || $1 || '%') 
		AND deleted_at IS NULL`
	if err := r.db.QueryRow(countQuery, search).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, name, region, created_at, updated_at
		FROM cities
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR region ILIKE '%' || $1 || '%') AND deleted_at IS NULL
		ORDER BY name
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	cities := []City{}
	for rows.Next() {
		var ct City
		if err := rows.Scan(&ct.ID, &ct.Name, &ct.Region, &ct.CreatedAt, &ct.UpdatedAt); err != nil {
			return nil, 0, err
		}
		cities = append(cities, ct)
	}
	return cities, total, rows.Err()
}

func (r *Repository) Update(ct *City) error {
	query := `
		UPDATE cities
		SET name = $1, region = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING updated_at`
	err := r.db.QueryRow(query, ct.Name, ct.Region, ct.ID).Scan(&ct.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec(
		`UPDATE cities SET deleted_at = NOW()
	 WHERE id = $1 AND deleted_at IS NULL`, id)
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
