package models

import (
	"database/sql"

	_ "github.com/lib/pq" //...
)

// Db connect
var Db *sql.DB
