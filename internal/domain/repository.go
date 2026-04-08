package models

type Repository struct {
	ID          int    `json:"id" db:"id"`
	Owner       string `json:"owner" db:"owner"`
	Name        string `json:"name" db:"name"`
	LastSeenTag string `json:"last_seen_tag" db:"last_seen_tag"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}
