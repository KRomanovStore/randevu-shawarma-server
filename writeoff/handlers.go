package writeoff

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"randevu-shawarma-server/users"
	"randevu-shawarma-server/warehouse"

	"github.com/julienschmidt/httprouter"
)

var db *sql.DB

// SetDatabase sets the database connection
func SetDatabase(database *sql.DB) {
	db = database
}

// RegisterRoutes registers all write-off routes
func RegisterRoutes(router *httprouter.Router) {
	router.POST("/write-off", users.Authenticate(CreateWriteOff))
}

func CreateWriteOff(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var newWriteOff WriteOff
	err := json.NewDecoder(r.Body).Decode(&newWriteOff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert new write off
	err = tx.QueryRow(
		"INSERT INTO public.\"Write_off\" (user_id, created_at, notes) VALUES ($1, $2, $3) RETURNING id",
		newWriteOff.UserID, time.Now(), newWriteOff.Notes,
	).Scan(&newWriteOff.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert write off products and update warehouse
	for _, product := range newWriteOff.Products {
		// Insert into write_off_product_relations
		_, err = tx.Exec(
			"INSERT INTO public.\"Write_off_product_relations\" (write_off_id, product_id, quantity) VALUES ($1, $2, $3)",
			newWriteOff.ID, product.ProductID, product.Quantity,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update warehouse
		var currentStock sql.NullFloat64

		err = tx.QueryRow(
			"SELECT current_stock FROM public.\"Warehouse\" WHERE product_id = $1",
			product.ProductID,
		).Scan(&currentStock)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if currentStock.Valid {
			newStock := currentStock.Float64 - product.Quantity
			if newStock < 0 {
				http.Error(w, "Insufficient stock", http.StatusBadRequest)
				return
			}

			_, err = tx.Exec(
				"UPDATE public.\"Warehouse\" SET current_stock = $1 WHERE product_id = $2",
				newStock, product.ProductID,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Product not found in warehouse", http.StatusBadRequest)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	warehouse.GetWarehouse(w, r, ps)
}
