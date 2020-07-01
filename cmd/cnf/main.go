package cnf

import (
	"database/sql"

	"github.com/dalais/sdku_backend/config"
)

// APIKey ... Глобальный секретный ключ
var APIKey []byte

// Db ...
var Db *sql.DB

// ROOT ...
var ROOT string

// Conf ...
var Conf config.LocalConfig
