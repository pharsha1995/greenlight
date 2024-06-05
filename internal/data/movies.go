package data

import (
	"time"

	"github.com/pharsha1995/greenlight/internal/data/validator"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   int32     `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, m *Movie) {
	v.Check(validator.ValidString(m.Title, 500), "title", "must not be empty and less than 500 bytes")
	v.Check(validator.WithinRange(m.Year, 1888, int32(time.Now().Year())), "year", "must be between 1888 and current year")
	v.Check(m.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(m.Genres != nil, "genres", "must be provided")
	v.Check(validator.WithinRange(len(m.Genres), 1, 5), "genres", "must contain between 1 and 5 genres")
	v.Check(validator.Unique(m.Genres), "genres", "must not contain duplicate and empty values")
}
