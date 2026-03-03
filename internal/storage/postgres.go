package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	Pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, connString string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	s := &PostgresStorage{Pool: pool}
	if err := s.InitSchema(ctx); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return s, nil
}

func (s *PostgresStorage) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS sites (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL,
		api_key TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`
	_, err := s.Pool.Exec(ctx, query)
	return err
}

func (s *PostgresStorage) GetValidSiteIDs(ctx context.Context) (map[string]bool, error) {
	rows, err := s.Pool.Query(ctx, "SELECT id FROM sites")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	return ids, nil
}

func (s *PostgresStorage) Close() {
	s.Pool.Close()
}
