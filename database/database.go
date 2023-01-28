// ./database/database.go
package database

import (
	"context"
	"database/sql"
)

type Database struct {
	SqlDb *sql.DB
}

var dbContext = context.Background()
