package orders

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"randevu-shawarma-server/users"

	"github.com/julienschmidt/httprouter"
)

var db *sql.DB

// SetDatabase sets the database connection
func SetDatabase(database *sql.DB) {
	db = database
}

// RegisterRoutes registers all orders routes
func RegisterRoutes(router *httprouter.Router) {
	router.POST("/orders", users.Authenticate(CreateOrder))
	router.PUT("/orders", users.Authenticate(UpdateOrder))
}

func CreateOrder(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var newOrder Order
	err := json.NewDecoder(r.Body).Decode(&newOrder)
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

	// Insert new order
	err = tx.QueryRow(
		"INSERT INTO public.\"Orders\" (user_id, name, created_at, processing, sold) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		newOrder.UserID, newOrder.Name, time.Now(), true, false,
	).Scan(&newOrder.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert order dishes
	for _, dish := range newOrder.Dishes {
		_, err = tx.Exec(
			"INSERT INTO public.\"Order_dish_relations\" (order_id, dish_id, quantity) VALUES ($1, $2, $3)",
			newOrder.ID, dish.DishID, dish.Quantity,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newOrder)
}

func UpdateOrder(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var updateData struct {
		OrderID int  `json:"order_id"`
		Sold    bool `json:"sold"`
	}
	err := json.NewDecoder(r.Body).Decode(&updateData)
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

	if updateData.Sold {
		// Fetch order dishes
		rows, err := tx.Query(
			"SELECT dish_id, quantity FROM public.\"Order_dish_relations\" WHERE order_id = $1",
			updateData.OrderID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var dishID int
		var quantity int

		// Update warehouse inventory
		for rows.Next() {
			err := rows.Scan(&dishID, &quantity)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Fetch dish recipe
			recipeRows, err := tx.Query(
				"SELECT product_id, quantity FROM public.\"Dish_recipe\" WHERE dish_id = $1",
				dishID,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer recipeRows.Close()

			var productID int
			var productQuantity float64

			for recipeRows.Next() {
				err := recipeRows.Scan(&productID, &productQuantity)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// Update warehouse
				_, err = tx.Exec(
					"UPDATE public.\"Warehouse\" SET current_stock = current_stock - $1 WHERE product_id = $2",
					productQuantity*float64(quantity), productID,
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		// Update order status
		_, err = tx.Exec(
			"UPDATE public.\"Orders\" SET processing = $1, sold = $2 WHERE id = $3",
			false, true, updateData.OrderID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Update order status to processing false
		_, err = tx.Exec(
			"UPDATE public.\"Orders\" SET processing = $1 WHERE id = $2",
			false, updateData.OrderID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
