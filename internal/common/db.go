package common

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Init() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		"mimir_app",
		"mimir_password",
		"127.0.0.1",
		"3306",
		"mimir",
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	DB.SetMaxIdleConns(10)
	DB.SetMaxOpenConns(100)
	DB.SetConnMaxLifetime(time.Hour)

	if err = DB.Ping(); err != nil {
		panic(err)
	}
}

func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			fmt.Printf("Error closing database: %v\n", err)
		}
	}
}
