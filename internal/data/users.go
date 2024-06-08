package data

import (
	"errors"
	"time"

	"github.com/pharsha1995/greenlight/internal/data/validator"
	"golang.org/x/crypto/bcrypt"
)

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
