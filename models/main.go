package models

import (
	"database/sql"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" //...

	"github.com/dalais/sdku_backend/config"
)

var db *sql.DB

func init() {
	// Загрузка значений из .env файла в систему
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	InitDB()
}

func main() {
}

// InitDB ...
func InitDB() {
	conf := config.New()
	dbEngine := conf.DbConnection
	dbURL := conf.DbConnection + "://" + conf.DbUsername + ":" + conf.DbPassword + "@" + conf.DbHost + "/" + conf.DbDatabase
	var err error
	db, err = sql.Open(dbEngine, dbURL)
	if err != nil {
		log.Panic(err)
	}

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}
}
