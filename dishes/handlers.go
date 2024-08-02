package dishes

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"randevu-shawarma-server/users"

	"github.com/julienschmidt/httprouter"
)

var db *sql.DB

// SetDatabase sets the database connection
func SetDatabase(database *sql.DB) {
	db = database
}

// RegisterRoutes registers all dishes routes
func RegisterRoutes(router *httprouter.Router) {
	router.GET("/dishes", users.Authenticate(GetDishes))
}

func GetDishes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rows, err := db.Query("SELECT id, name, price FROM public.\"Dishes\" WHERE is_active = true")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var dishes []DishItem
	for rows.Next() {
		var item DishItem
		err := rows.Scan(&item.ID, &item.Name, &item.Price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dishes = append(dishes, item)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dishes)
}
