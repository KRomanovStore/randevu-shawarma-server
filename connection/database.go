package connection

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"randevu-shawarma-server/supply"
	"randevu-shawarma-server/users"
)

func initDB() *sql.DB {
	connStr := "user=kostia password=foDfyf-vufvim-muvwy9 dbname=randevu_database host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func СonnectToDatabase() {
	db := initDB()
	defer db.Close()

	users.SetDatabase(db)
	supply.SetDatabase(db)
}
