// ./database/database.go
package hotworxData

import (
	"context"
	"database/sql"
)

type Database struct {
	SqlDb *sql.DB
}

var dbContext = context.Background()
