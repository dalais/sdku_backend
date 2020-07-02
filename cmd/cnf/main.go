package cnf

import (
	"database/sql"

	"github.com/dalais/sdku_backend/config"
)

// APIKey ... Glopal app key
var APIKey []byte

// Db ...
var Db *sql.DB

// ROOT Project root path
var ROOT string

// Conf ...
var Conf config.LocalConfig
