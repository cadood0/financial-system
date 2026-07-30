package city

import "time"

type City struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Region    string     `json:"region"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}
