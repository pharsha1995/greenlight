package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
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
	v.Check(validator.ValidString(m.Title, 1, 500), "title", "must not be empty and less than 500 bytes")
	v.Check(validator.WithinRange(m.Year, 1888, int32(time.Now().Year())), "year", "must be between 1888 and current year")
	v.Check(m.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(m.Genres != nil, "genres", "must be provided")
	v.Check(validator.WithinRange(len(m.Genres), 1, 5), "genres", "must contain between 1 and 5 genres")
	v.Check(validator.Unique(m.Genres), "genres", "must not contain duplicate and empty values")
}

type MovieModel struct {
	DB *sql.DB
}

func (m *MovieModel) Insert(movie *Movie) error {
	stmt := `INSERT INTO movies (title, year, runtime, genres)
	         VALUES ($1, $2, $3, $4)
					 RETURNING id, created_at, version`

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m *MovieModel) Get(id int64) (*Movie, error) {
	stmt := `SELECT id, created_at, title, year, runtime, genres, version
	         FROM movies
					 WHERE id = $1`

	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return &movie, nil
}

func (m *MovieModel) Update(movie *Movie) error {
	stmt := `UPDATE movies
	         SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
					 WHERE id = $5 AND version = $6
					 RETURNING version`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		}
		return err
	}

	return nil
}

func (m *MovieModel) Delete(id int64) error {
	stmt := `DELETE FROM movies
	         WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected == 0 {
		return ErrNoRecord
	}

	return nil
}

func (m *MovieModel) GetAll(title string, genres []string, filters *Filters) ([]*Movie, *Metadata, error) {
	stmt := fmt.Sprintf(`
	        SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
	        FROM movies
					WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
					AND (genres @> $2 OR $2 = '{}')
					ORDER BY %s %s, id ASC
					LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	args := []any{title, pq.Array(genres), filters.limit(), filters.offset()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	totalRecords := 0
	movies := []*Movie{}
	for rows.Next() {
		movie := Movie{}
		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, nil, err
		}

		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	return movies, calculateMetadata(totalRecords, filters.Page, filters.PageSize), nil
}
