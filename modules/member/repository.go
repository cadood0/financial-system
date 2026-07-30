package member

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("member not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(m *Member) error {
	query := `
		INSERT INTO members (full_name, phone, city_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(query, m.FullName, m.Phone, m.CityID).
		Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

func (r *Repository) FindByID(id int64) (*Member, error) {
	query := `
		SELECT m.id, m.full_name, m.phone, m.city_id, c.name, m.created_at, m.updated_at
		FROM members m
		JOIN cities c ON c.id = m.city_id
		WHERE m.id = $1 AND m.deleted_at IS NULL`

	var m Member
	err := r.db.QueryRow(query, id).
		Scan(&m.ID, &m.FullName, &m.Phone, &m.CityID, &m.CityName, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) FindByPhone(phone string) (*Member, error) {
	query := `
		SELECT m.id, m.full_name, m.phone, m.city_id, c.name, m.created_at, m.updated_at
		FROM members m
		JOIN cities c ON c.id = m.city_id
		WHERE m.phone = $1 AND m.deleted_at IS NULL`

	var m Member
	err := r.db.QueryRow(query, phone).
		Scan(&m.ID, &m.FullName, &m.Phone, &m.CityID, &m.CityName, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) List(search string, cityID int64, limit, offset int) ([]Member, int, error) {
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM members m
		WHERE m.deleted_at IS NULL
		  AND ($1 = '' OR m.full_name ILIKE '%' || $1 || '%' OR m.phone ILIKE '%' || $1 || '%')
		  AND ($2 = 0 OR m.city_id = $2)`
	if err := r.db.QueryRow(countQuery, search, cityID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT m.id, m.full_name, m.phone, m.city_id, c.name, m.created_at, m.updated_at
		FROM members m
		JOIN cities c ON c.id = m.city_id
		WHERE m.deleted_at IS NULL
		  AND ($1 = '' OR m.full_name ILIKE '%' || $1 || '%' OR m.phone ILIKE '%' || $1 || '%')
		  AND ($2 = 0 OR m.city_id = $2)
		ORDER BY m.id
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(query, search, cityID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	members := []Member{}
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.FullName, &m.Phone, &m.CityID, &m.CityName, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, err
		}
		members = append(members, m)
	}
	return members, total, rows.Err()
}

func (r *Repository) Update(m *Member) error {
	query := `
		UPDATE members
		SET full_name = $1, phone = $2, city_id = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING updated_at`
	err := r.db.QueryRow(query, m.FullName, m.Phone, m.CityID, m.ID).Scan(&m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec(
		`UPDATE members SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
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

func (r *Repository) CountActiveByCity(cityID int64) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM members WHERE city_id = $1 AND deleted_at IS NULL`, cityID).
		Scan(&count)
	return count, err
}
