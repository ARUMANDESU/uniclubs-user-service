package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	DB *sql.DB
}

func New(databaseDSN string) (*Storage, error) {
	const op = "storage.postgresql.new"

	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: set options for db

	return &Storage{DB: db}, nil
}

func (s *Storage) Close() error {
	return s.DB.Close()
}

func (s *Storage) SaveUser(ctx context.Context, user *domain.User) error {
	const op = "storage.postgresql.saveUser"

	query := `
		INSERT INTO users(email, pass_hash, first_name,last_name, barcode, major, group_name, year, role_id)
		values($1, $2, $3, $4, $5, $6, $7, $8, DEFAULT)
		returning id;
	`

	args := []any{
		user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.Barcode, user.Major, user.GroupName, user.Year,
	}

	result := s.DB.QueryRowContext(ctx, query, args...)

	err := result.Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("%s: %w", op, domain.ErrUserExists)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	const op = "storage.postgresql.getUserByID"

	query := `
		SELECT u.id, u.email, u.pass_hash, u.first_name,u.last_name, u.avatar_url, u.created_at,
		       u.barcode, u.major, u.group_name, u.year, r.name as role
		FROM users u LEFT JOIN roles r
		ON  u.role_id = r.id
		WHERE u.id = $1;
	`

	result := s.DB.QueryRowContext(ctx, query, userID)
	user := domain.User{}

	err := result.Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.AvatarURL,
		&user.CreatedAt, &user.Barcode, &user.Major, &user.GroupName, &user.Year, &user.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const op = "storage.postgresql.getUserByEmail"

	query := `
		SELECT u.id, u.email, u.pass_hash, u.first_name,u.last_name, u.avatar_url,
		       u.created_at, u.barcode, u.major, u.group_name, u.year, r.name as role
		FROM users u LEFT JOIN roles r
		ON  u.role_id = r.id
		WHERE u.email = $1 and u.activated;
	`

	result := s.DB.QueryRowContext(ctx, query, email)
	user := domain.User{}

	err := result.Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.AvatarURL,
		&user.CreatedAt, &user.Barcode, &user.Major, &user.GroupName, &user.Year, &user.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) GetUserRoleByID(ctx context.Context, userID int64) (string, error) {
	const op = "storage.postgresql.getUserRoleByID"

	query := `
		SELECT r.name
		FROM users u left join roles r 
		ON u.role_id = r.id
		where u.id = $1 and u.activated;
	`
	var role string

	err := s.DB.QueryRowContext(ctx, query, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, domain.ErrUserNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return role, nil
}

func (s *Storage) UpdateUser(ctx context.Context, user *domain.User) error {
	const op = "storage.postgresql.updateUser"

	query := `
		UPDATE users
		SET email = $2, first_name = $3, last_name = $4,
		    phone_number = $5, barcode = $6, major = $7,
		    group_name = $8, year = $9, avatar_url = $10
		WHERE id = $1 and activated;
	`

	args := []any{
		user.ID, user.Email, user.FirstName, user.LastName, user.PhoneNumber,
		user.Barcode, user.Major, user.GroupName, user.Year, user.AvatarURL,
	}

	result, err := s.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrUserNotFound)
	}

	return nil
}

func (s *Storage) DeleteUserByID(ctx context.Context, userID int64) error {
	const op = "storage.postgresql.deleteUserByID"

	result, err := s.DB.ExecContext(ctx, `DELETE FROM users WHERE id = $1 and activated`, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrUserNotFound)
	}

	return nil
}

func (s *Storage) ActivateUser(ctx context.Context, userID int64) error {
	const op = "storage.postgresql.activateUser"

	result, err := s.DB.ExecContext(ctx, `UPDATE users SET activated = true  WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrUserNotFound)
	}

	return nil
}

func (s *Storage) GetAll(ctx context.Context, query string, filters domain.Filters) ([]*domain.User, domain.Metadata, error) {
	const op = "storage.postgresql.getAll"

	q := `
		SELECT count(*) OVER(), u.id,u.email, u.first_name, u.last_name,u.avatar_url,
		       u.created_at, u.barcode, u.major,u.group_name, u.year, r.name as role
		FROM users u LEFT JOIN roles r
		ON  u.role_id = r.id
		WHERE 
			( (STRPOS(LOWER(email), LOWER($1)) > 0 OR $1 = '') OR
			(STRPOS(LOWER(first_name), LOWER($1)) > 0 OR $1 = '') OR
			(STRPOS(LOWER(last_name), LOWER($1)) > 0 OR $1 = '') ) 
		  	AND 
		    u.activated
		ORDER BY id ASC
        LIMIT $2 OFFSET $3;
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	args := []any{query, filters.Limit(), filters.Offset()}

	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, domain.Metadata{}, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	var totalRecords int32
	var users []*domain.User

	for rows.Next() {
		var user domain.User

		err = rows.Scan(
			&totalRecords, &user.ID, &user.Email, &user.FirstName, &user.LastName, &user.AvatarURL,
			&user.CreatedAt, &user.Barcode, &user.Major, &user.GroupName, &user.Year, &user.Role,
		)
		if err != nil {
			return nil, domain.Metadata{}, fmt.Errorf("%s: %w", op, err)
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, domain.Metadata{}, fmt.Errorf("%s: %w", op, err)
	}

	metadata := domain.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return users, metadata, nil
}

func (s *Storage) UpdateUserRole(ctx context.Context, userID int64, role string) error {
	const op = "storage.postgresql.updateUserRole"

	query := `
		UPDATE users
		SET role_id = r.id
		FROM roles r
		WHERE users.id = $1 AND users.activated AND r.name = $2;
	`

	result, err := s.DB.ExecContext(ctx, query, userID, role)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrUserNotFound)
	}

	return nil
}

func (s *Storage) DeleteNonActivatedUsers(ctx context.Context, days int) error {
	const op = "storage.postgresql.deleteNonActivatedUsers"

	query := `
		DELETE FROM users
		WHERE NOT activated AND created_at < NOW() - $1::interval;
	`

	result, err := s.DB.ExecContext(ctx, query, fmt.Sprintf("%d days", days))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
