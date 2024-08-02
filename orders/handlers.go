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
	router.GET("/orders", users.Authenticate(GetOrders))
	router.POST("/orders", users.Authenticate(CreateOrder))
	router.PUT("/orders", users.Authenticate(UpdateOrder))
}

func GetOrders(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := `
		SELECT o.id, o.user_id, o.name, SUM(d.price * odr.quantity) AS total_price
		FROM public."Orders" o
		JOIN public."Order_dish_relations" odr ON o.id = odr.order_id
		JOIN public."Dishes" d ON odr.dish_id = d.id
		WHERE o.processing = true
		GROUP BY o.id, o.user_id, o.name
	`
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []OrderView
	for rows.Next() {
		var order OrderView
		err := rows.Scan(&order.ID, &order.UserID, &order.Name, &order.TotalPrice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		dishQuery := `
			SELECT d.id, d.name, odr.quantity, d.price
			FROM public."Order_dish_relations" odr
			JOIN public."Dishes" d ON odr.dish_id = d.id
			WHERE odr.order_id = $1
		`
		dishRows, err := db.Query(dishQuery, order.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dishRows.Close()

		var dishes []OrderDishRelationView
		for dishRows.Next() {
			var dish OrderDishRelationView
			err := dishRows.Scan(&dish.DishID, &dish.DishName, &dish.Quantity, &dish.Price)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			dishes = append(dishes, dish)
		}
		order.Dishes = dishes
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func CreateOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	GetOrders(w, r, ps)
}

func UpdateOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var updateData struct {
		OrderID int  `json:"orderId"`
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
		// Fetch products
		query := `
		WITH product_quantities AS (
			SELECT dr.product_id, SUM(dr.quantity * odr.quantity) AS total_quantity
			FROM public."Order_dish_relations" odr
			INNER JOIN public."Dish_recipe" dr ON odr.dish_id = dr.dish_id
			WHERE odr.order_id = $1
			GROUP BY dr.product_id
			UNION ALL
			SELECT pr.product_id, SUM(pr.quantity * odr.quantity) AS total_quantity
			FROM public."Order_dish_relations" odr
			INNER JOIN public."Dishes_Preparations" dp ON odr.dish_id = dp.dishes_id
			INNER JOIN public."Preparation_recipe" pr ON dp.preparations_id = pr.preparation_id
			WHERE odr.order_id = $1
			GROUP BY pr.product_id
		)
		SELECT product_id, SUM(total_quantity) AS total_quantity
		FROM product_quantities
		GROUP BY product_id
		`

		rows, err := db.Query(query, updateData.OrderID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var productID int
		var totalQuantity float64

		// Update warehouse inventory
		for rows.Next() {
			err := rows.Scan(&productID, &totalQuantity)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec(
				"UPDATE public.\"Warehouse\" SET current_stock = current_stock - $1 WHERE product_id = $2",
				totalQuantity, productID,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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

	GetOrders(w, r, ps)
}
