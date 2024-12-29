package store

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type GoogleUser struct {
	ID            uuid.UUID `json:"id" db:"id"`
	GoogleId      string    `json:"google_id" db:"google_id"`
	Email         string    `json:"email" db:"email"`
	VerifiedEmail bool      `json:"verified_email" db:"verified_email"`
	Name          string    `json:"name" db:"name"`
	Picture       string    `json:"picture" db:"picture"`
	Locale        string    `json:"locale" db:"locale"`
}

type GoogleUserStore interface {
	GetGoogleUserWhereGoogleId(id string) (*GoogleUser, error)
	CreateGoogleUser(googleUser *GoogleUser) error
	UpdateGoogleUser(googleUser *GoogleUser) error
}

type GoogleUserPostgresStore struct {
	db *sqlx.DB
}

func NewGoogleUserPostgresStore(db *sqlx.DB) *GoogleUserPostgresStore {
	return &GoogleUserPostgresStore{db: db}
}

func (s *GoogleUserPostgresStore) GetGoogleUserWhereGoogleId(id string) (*GoogleUser, error) {
	query := `SELECT id, google_id, email, verified_email, name, picture, locale FROM google_users WHERE google_id = $1`
	var googleUser GoogleUser
	err := s.db.Get(&googleUser, query, id)
	if err != nil {
		return nil, err
	}
	return &googleUser, nil
}

func (s *GoogleUserPostgresStore) CreateGoogleUser(googleUser *GoogleUser) error {
	query := `INSERT INTO google_users (google_id, email, verified_email, name, picture, locale) 
	          VALUES (:google_id, :email, :verified_email, :name, :picture, :locale)`
	_, err := s.db.NamedExec(query, googleUser)
	if err != nil {
		return err
	}
	return nil
}

func (s *GoogleUserPostgresStore) UpdateGoogleUser(googleUser *GoogleUser) error {
	query := `UPDATE google_users 
	          SET email = :email, verified_email = :verified_email, name = :name, picture = :picture, locale = :locale 
	          WHERE id = :id`
	_, err := s.db.NamedExec(query, googleUser)
	if err != nil {
		return err
	}
	return nil
}
