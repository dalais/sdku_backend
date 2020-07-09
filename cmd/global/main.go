package gl

import (
	"database/sql"
	"log"
	"os"

	"github.com/dalais/sdku_backend/config"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/sessions"
)

// Db ...
var Db *sql.DB

// RPool ...
var RPool *redis.Pool

// ROOT Project root path
var ROOT string

// Conf ...
var Conf config.ENV

// StoreSession ...
var StoreSession *sessions.CookieStore

// InitRPool init redis connections pool
func InitRPool() {
	RPool = &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", Conf.Redis.Host+":"+Conf.Redis.Port)
			if err != nil {
				log.Printf("ERROR: fail init redis pool: %s", err.Error())
				os.Exit(1)
			}
			return conn, err
		},
	}
}

// Rping redis connection ping
func Rping(conn redis.Conn) bool {
	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		log.Printf("ERROR: fail ping redis conn: %s", err.Error())
		os.Exit(1)
		return false
	}
	return true
}
