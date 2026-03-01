package storage

import (
	"GoFlow/internal/model"
	"context"
	"github.com/ClickHouse/clickhouse-go/v2"
	"time"
)

type ClickHouseStorage struct {
	conn clickhouse.Conn
}

func NewClickHouseStorage(addr, user, password, db string) (*ClickHouseStorage, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: db,
			Username: user,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &ClickHouseStorage{conn: conn}, nil
}

func (s *ClickHouseStorage) InsertEvents(ctx context.Context, events []model.Event) error {
	batch, err := s.conn.PrepareBatch(ctx, "INSERT INTO events")
	if err != nil {
		return err
	}

	for _, e := range events {
		err := batch.Append(
			e.EventID,
			e.SiteID,
			e.Type,
			e.UserID,
			e.Path,
			e.Timestamp,
			e.ReceivedAt,
			e.Properties,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

func (s *ClickHouseStorage) Close() error {
	return s.conn.Close()
}
