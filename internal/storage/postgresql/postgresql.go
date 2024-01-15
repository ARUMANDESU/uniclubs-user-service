package postgresql

import "database/sql"

type Storage struct {
	DB *sql.DB
}
