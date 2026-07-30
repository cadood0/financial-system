package member

import "time"

type Member struct {
	ID        int64      `json:"id"`
	FullName  string     `json:"full_name"`
	Phone     string     `json:"phone"`
	CityID    int64      `json:"city_id"`
	CityName  string     `json:"city_name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}
