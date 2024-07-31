package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"

	"randevu-shawarma-server/orders"
	"randevu-shawarma-server/supply"
	"randevu-shawarma-server/users"
	"randevu-shawarma-server/writeoff"
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

func main() {
	db := initDB()
	defer db.Close()

	users.SetDatabase(db)
	supply.SetDatabase(db)
	writeoff.SetDatabase(db)
	orders.SetDatabase(db)

	router := httprouter.New()
	users.RegisterRoutes(router)
	supply.RegisterRoutes(router)
	writeoff.RegisterRoutes(router)
	orders.RegisterRoutes(router)

	log.Fatal(http.ListenAndServe(":8090", router))
}
