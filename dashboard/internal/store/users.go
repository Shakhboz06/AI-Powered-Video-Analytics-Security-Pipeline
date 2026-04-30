package store

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *sql.DB
}

type Users struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  Password `json:"-"`
	CreatedAt string   `json:"created_at"`
}

type Password struct {
	text *string
	hash []byte
}

type TokenVersion struct{
	Version int  `json:"token_version"`
}
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (pass *Password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	pass.text = &text
	pass.hash = hash

	return nil
}

func (pass *Password) Compare(text string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(pass.hash), []byte(text))

	return err == nil
}

func (s *UserStore) Create(ctx context.Context, user *Users) (*Users, error) {

	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3) RETURNING id, username, email
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	
	err := s.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password.hash).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserStore) GetUser(ctx context.Context, userEmail string) (*Users, error) {

	query := `SELECT id, username, email, password, created_at FROM users WHERE users.email = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user Users
	err := s.db.QueryRowContext(ctx, query, userEmail).Scan(&user.ID, &user.Username, &user.Email, &user.Password.hash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (s *UserStore) GetUserByID(ctx context.Context, id int64) (*Users, error) {
	query := `SELECT id, username, email, password, created_at FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user Users
	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password.hash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}


func (s *UserStore) UpdateUserTokenVersion(ctx context.Context, id int64) (*TokenVersion, error){
	quary := `UPDATE users SET token_version = token_version + 1 WHERE id = $1 RETURNING token_version`
	
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var tk_version TokenVersion 
	if err := s.db.QueryRowContext(ctx,quary,id).Scan(&tk_version.Version); err != nil{
		return  nil,err		
	}

	
	return &tk_version, nil
}


func (s *UserStore) GetUserTokenVersion(ctx context.Context, id int64) (*TokenVersion, error){
	quary := `SELECT token_version FROM users WHERE id = $1`
	
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var tk_version TokenVersion 
	if err := s.db.QueryRowContext(ctx,quary,id).Scan(&tk_version.Version); err != nil{
		return  nil,err		
	}

	
	return &tk_version, nil
}