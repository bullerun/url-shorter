package postgreSQL

import (
	"REST-API-Service/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Storage struct {
	db *pgx.Conn
}

func initPrepares(db *pgx.Conn) error {
	const op = "storage.postgreSQL.initPrepares"
	_, err := db.Prepare(context.Background(), "insertAlias", "INSERT INTO url(alias, url) VALUES($1, $2) RETURNING id;")
	if err != nil {
		db.Close(context.Background())
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Prepare(context.Background(), "getUrl", "SELECT url FROM url WHERE alias = $1")
	if err != nil {
		db.Close(context.Background())
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
func New(storagePath string) (*Storage, error) {
	const op = "storage.postgreSql.New"
	db, err := pgx.Connect(context.Background(), storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS url(
			id SERIAL PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL);
			CREATE INDEX IF NOT EXISTS url_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err = initPrepares(db); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}
func (s *Storage) SaveUrl(urlToSave, alias string) (int64, error) {
	const op = "storage.postgreSql.SaveUrl"

	var id int64

	err := s.db.QueryRow(context.Background(), "insertAlias", alias, urlToSave).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
func (s *Storage) GetUrl(alias string) (string, error) {
	const op = "storage.postgreSql.GetUrl"
	var fullUrl string
	err := s.db.QueryRow(context.Background(), "getUrl", alias).Scan(&fullUrl)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return fullUrl, nil
}
func (s *Storage) Close() {
	s.db.Close(context.Background())
}

//TODO DeleteURL()
