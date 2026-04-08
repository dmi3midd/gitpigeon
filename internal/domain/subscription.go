package models

type Subscription struct {
	ID           int    `json:"id" db:"id"`
	RepositoryID int    `json:"repository_id" db:"repository_id"`
	Email        string `json:"email" db:"email"`
	CreatedAt    string `json:"created_at" db:"created_at"`
}
