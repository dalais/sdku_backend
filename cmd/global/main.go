package gl

import (
	"database/sql"

	"github.com/dalais/sdku_backend/config"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/sessions"
)

// Db ...
var Db *sql.DB

// Rdb ...
var Rdb redis.Conn

// ROOT Project root path
var ROOT string

// Conf ...
var Conf config.ENV

// StoreSession ...
var StoreSession *sessions.CookieStore
