package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/pharsha1995/greenlight/internal/data/validator"
	"golang.org/x/crypto/bcrypt"
)

var ErrDuplicateEmail = errors.New("users: duplicate email")
var emailUniquePQErrMsg = `pq: duplicate key value violates unique constraint "users_email_key"`

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

type User struct {
	ID        int64     `json:"id"`
	CreateAt  time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(validator.ValidString(user.Name, 1, 500), "name", "must not be empty and less than 500 bytes")
	v.Check(validator.ValidString(user.Email, 1, 500), "email", "must not be empty and less than 500 bytes")
	v.Check(validator.Matches(user.Email, validator.EmailRX), "email", "must be a valid email address")

	if user.Password.plaintext != nil {
		v.Check(validator.ValidString(*user.Password.plaintext, 8, 72), "password", "must not be empty and between 8 and 72 bytes")
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(user *User) error {
	stmt := `INSERT INTO users (name, email, password_hash, activated)
					 VALUES ($1, $2, $3, $4)
					 RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.CreateAt, &user.Version)
	if err != nil {
		if err.Error() == emailUniquePQErrMsg {
			return ErrDuplicateEmail
		}
		return err
	}

	return nil
}

func (m *UserModel) GetByEmail(email string) (*User, error) {
	stmt := `SELECT id, created_at, name, email, password_hash, activated, version
					 FROM users
					 WHERE email = $1`

	user := User{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, email).Scan(
		&user.ID,
		&user.CreateAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return &user, nil
}

func (m *UserModel) Update(user *User) error {
	stmt := `UPDATE users
	         SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
					 WHERE id = $5 AND version = $6
					 RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.Version)
	if err != nil {
		if err.Error() == emailUniquePQErrMsg {
			return ErrDuplicateEmail
		} else if errors.Is(err, sql.ErrNoRows) {
			return ErrNoRecord
		}
		return err
	}

	return nil
}
