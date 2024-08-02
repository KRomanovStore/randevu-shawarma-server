package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"

	"randevu-shawarma-server/dishes"
	"randevu-shawarma-server/orders"
	"randevu-shawarma-server/supply"
	"randevu-shawarma-server/users"
	"randevu-shawarma-server/warehouse"
	"randevu-shawarma-server/writeoff"
)

func initDB() *sql.DB {

	// Get environment variables
	// dbHost := os.Getenv("DB_HOST")
	// dbPort := os.Getenv("DB_PORT")
	// dbUser := os.Getenv("DB_USER")
	// dbPassword := os.Getenv("DB_PASSWORD")
	// dbName := os.Getenv("DB_NAME")

	// println(dbHost)
	// println(dbPort)
	// println(dbUser)
	// println(dbPassword)
	// println(dbName)

	// Connect to the database
	// connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	// 	dbHost, dbPort, dbUser, dbPassword, dbName)

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

func setupCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the headers for CORS
		w.Header().Set("Access-Control-Allow-Origin", "https://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	db := initDB()
	defer db.Close()

	users.SetDatabase(db)
	supply.SetDatabase(db)
	writeoff.SetDatabase(db)
	orders.SetDatabase(db)
	warehouse.SetDatabase(db)
	dishes.SetDatabase(db)

	router := httprouter.New()
	users.RegisterRoutes(router)
	supply.RegisterRoutes(router)
	writeoff.RegisterRoutes(router)
	orders.RegisterRoutes(router)
	warehouse.RegisterRoutes(router)
	dishes.RegisterRoutes(router)

	corsRouter := setupCORS(router)

	log.Fatal(http.ListenAndServe(":8090", corsRouter))
}
