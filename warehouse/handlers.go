package warehouse

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

// RegisterRoutes registers all warehouse routes
func RegisterRoutes(router *httprouter.Router) {
	router.GET("/warehouse", users.Authenticate(GetWarehouse))
}

func GetWarehouse(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := `
	SELECT w.id, w.product_id, p.name, w.current_stock, w.average_cost 
	FROM public."Warehouse" w
	INNER JOIN public."Products" p ON w.product_id = p.id
	`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var warehouseItems []WarehouseItem
	for rows.Next() {
		var item WarehouseItem
		err := rows.Scan(&item.ID, &item.ProductID, &item.ProductName, &item.CurrentStock, &item.AverageCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		warehouseItems = append(warehouseItems, item)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(warehouseItems)
}
