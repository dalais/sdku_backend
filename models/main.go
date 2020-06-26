package models

import (
	"database/sql"
	"log"
	"time"

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

	// Настройка количества подключений к базе данных
	// для большей информации можно почитать нижеприведенную статью
	// https://www.alexedwards.net/blog/configuring-sqldb
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
}
