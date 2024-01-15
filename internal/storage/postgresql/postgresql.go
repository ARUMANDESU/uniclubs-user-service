package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/models"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	DB *sql.DB
}

func New(databaseDSN string) (*Storage, error) {
	const op = "storage.postgresql.New"

	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, user *models.User) error {
	const op = "storage.postgresql.SaveUser"

	stmt, err := s.DB.Prepare(`
		INSERT INTO users(email, pass_hash, first_name, last_name, barcode, major, group_name, year, role_id)
		values($1, $2, $3, $4, $5, $6, $7, $8, DEFAULT)
		returning id;
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	args := []any{
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Barcode,
		user.Major,
		user.GroupName,
		user.Year,
	}

	result := stmt.QueryRowContext(ctx, args...)

	err = result.Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("%s: %w", op, storage.ErrUserExists)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUser(ctx context.Context, userID int64) (user models.User, err error) {
	//TODO implement me
	panic("implement me")
}
