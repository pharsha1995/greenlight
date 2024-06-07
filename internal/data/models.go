package data

import (
	"database/sql"
	"errors"
)

var (
	ErrNoRecord     = errors.New("models: no matching record found")
	ErrEditConflict = errors.New("models: edit conflict when updating record")
)

type Models struct {
	Movies *MovieModel
}

func NewModels(db *sql.DB) *Models {
	return &Models{
		Movies: &MovieModel{DB: db},
	}
}
