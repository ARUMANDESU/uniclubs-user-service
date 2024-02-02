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

	err = db.Ping()
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

	defer stmt.Close()

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

func (s *Storage) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	const op = "storage.postgresql.GetUserByEmail"

	stmt, err := s.DB.Prepare(`
		SELECT u.id, u.email, u.pass_hash, u.first_name, u.last_name, u.created_at, u.barcode, u.major, u.group_name, u.year, r.name as role
		FROM users u LEFT JOIN roles r
		ON  u.role_id = r.id
		WHERE u.id = $1;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	result := stmt.QueryRowContext(ctx, userID)
	user := models.User{}

	err = result.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.Barcode, &user.Major, &user.GroupName, &user.Year, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotExists)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "storage.postgresql.GetUserByEmail"

	stmt, err := s.DB.Prepare(`
		SELECT u.id, u.email, u.pass_hash, u.first_name, u.last_name, u.created_at, u.barcode, u.major, u.group_name, u.year, r.name as role
		FROM users u LEFT JOIN roles r
		ON  u.role_id = r.id
		WHERE u.email = $1;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	result := stmt.QueryRowContext(ctx, email)
	user := models.User{}

	err = result.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.Barcode, &user.Major, &user.GroupName, &user.Year, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotExists)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) GetUserRoleByID(ctx context.Context, userID int64) (role string, err error) {
	const op = "storage.postgresql.GetUserRoleByID"

	stmt, err := s.DB.Prepare(`
		SELECT r.name
		FROM users u left join roles r 
		ON u.role_id = r.id
		where u.id = $1;
	`)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrUserNotExists)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return role, nil
}

func (s *Storage) UpdateUser(ctx context.Context, user models.User) error {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) DeleteUserByID(ctx context.Context, userID int64) error {
	//TODO implement me
	panic("implement me")
}
