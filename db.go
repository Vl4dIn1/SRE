package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/001_init.sql
var initSQL string

func InitDB(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации драйвера БД: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	maxRetries := 15
	for i := 1; i <= maxRetries; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = db.PingContext(pingCtx)
		cancel()

		if err == nil {
			slog.Info("Успешное подключение к PostgreSQL", "attempt", i)
			break
		}

		if i == maxRetries {
			break
		}

		slog.Warn("База данных ещё не готова, ожидание...", "attempt", i, "max_retries", maxRetries, "err", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных после %d попыток: %w", maxRetries, err)
	}

	if _, err := db.ExecContext(ctx, initSQL); err != nil {
		return nil, fmt.Errorf("ошибка применения схемы БД: %w", err)
	}

	return db, nil
}

func GetAllServers(ctx context.Context, db *sql.DB) ([]Server, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name, ip_address, environment, status, created_at FROM servers ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := make([]Server, 0)
	for rows.Next() {
		var s Server
		if err := rows.Scan(&s.ID, &s.Name, &s.IPAddress, &s.Environment, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}

	return servers, rows.Err()
}

func GetServerByID(ctx context.Context, db *sql.DB, id int) (*Server, error) {
	var s Server
	err := db.QueryRowContext(ctx, "SELECT id, name, ip_address, environment, status, created_at FROM servers WHERE id = $1", id).
		Scan(&s.ID, &s.Name, &s.IPAddress, &s.Environment, &s.Status, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func CreateServer(ctx context.Context, db *sql.DB, req ServerInput) (*Server, error) {
	query := `
		INSERT INTO servers (name, ip_address, environment, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, ip_address, environment, status, created_at
	`
	var s Server
	err := db.QueryRowContext(ctx, query, req.Name, req.IPAddress, req.Environment, req.Status).Scan(
		&s.ID, &s.Name, &s.IPAddress, &s.Environment, &s.Status, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func UpdateServer(ctx context.Context, db *sql.DB, id int, req ServerInput) (*Server, error) {
	query := `
		UPDATE servers
		SET name = $1, ip_address = $2, environment = $3, status = $4
		WHERE id = $5
		RETURNING id, name, ip_address, environment, status, created_at
	`
	var s Server
	err := db.QueryRowContext(ctx, query, req.Name, req.IPAddress, req.Environment, req.Status, id).Scan(
		&s.ID, &s.Name, &s.IPAddress, &s.Environment, &s.Status, &s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func DeleteServer(ctx context.Context, db *sql.DB, id int) (bool, error) {
	res, err := db.ExecContext(ctx, "DELETE FROM servers WHERE id = $1", id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
