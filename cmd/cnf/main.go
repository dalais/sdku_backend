package cnf

import (
	"database/sql"

	"github.com/dalais/sdku_backend/config"
	"github.com/gorilla/sessions"
)

// Db ...
var Db *sql.DB

// ROOT Project root path
var ROOT string

// Conf ...
var Conf config.LocalConfig

// StoreSession ...
var StoreSession *sessions.CookieStore
