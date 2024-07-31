package supply

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"randevu-shawarma-server/users"

	"github.com/julienschmidt/httprouter"
)

var db *sql.DB

// SetDatabase sets the database connection
func SetDatabase(database *sql.DB) {
	db = database
}

// RegisterRoutes registers all supply routes
func RegisterRoutes(router *httprouter.Router) {
	router.POST("/supply", users.Authenticate(CreateSupply))
}

func CreateSupply(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var newSupply Supply
	err := json.NewDecoder(r.Body).Decode(&newSupply)
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

	// Insert new supply
	err = tx.QueryRow(
		"INSERT INTO public.\"Supply\" (user_id, created_at) VALUES ($1, $2) RETURNING id",
		newSupply.UserID, time.Now(),
	).Scan(&newSupply.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert supply products and update warehouse
	for _, product := range newSupply.Products {
		// Insert into supply_product_relations
		_, err = tx.Exec(
			"INSERT INTO public.\"Supply_product_relations\" (supply_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)",
			newSupply.ID, product.ProductID, product.Quantity, product.Price,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update warehouse
		var currentStock sql.NullFloat64
		var averageCost sql.NullString

		err = tx.QueryRow(
			"SELECT current_stock, average_cost FROM public.\"Warehouse\" WHERE product_id = $1",
			product.ProductID,
		).Scan(&currentStock, &averageCost)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert averageCost to float64 if it exists
		var avgCostFloat64 float64
		if averageCost.Valid {
			avgCostFloat64, err = parseMoneyToFloat(averageCost.String)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		productPriceFloat64, err := parseMoneyToFloat(product.Price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newStock := product.Quantity
		newCost := productPriceFloat64

		if currentStock.Valid && averageCost.Valid {
			totalCost := (currentStock.Float64 * avgCostFloat64) + (product.Quantity * productPriceFloat64)
			newStock += currentStock.Float64
			newCost = totalCost / newStock
		}

		newCostStr := formatFloatToMoney(newCost)

		if currentStock.Valid {
			_, err = tx.Exec(
				"UPDATE public.\"Warehouse\" SET current_stock = $1, average_cost = $2 WHERE product_id = $3",
				newStock, newCostStr, product.ProductID,
			)
		} else {
			_, err = tx.Exec(
				"INSERT INTO public.\"Warehouse\" (product_id, current_stock, average_cost) VALUES ($1, $2, $3)",
				product.ProductID, newStock, newCostStr,
			)
		}

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
	json.NewEncoder(w).Encode(newSupply)
}

func parseMoneyToFloat(moneyStr string) (float64, error) {
	cleanedStr := strings.Replace(moneyStr, "$", "", -1)
	return strconv.ParseFloat(cleanedStr, 64)
}

func formatFloatToMoney(f float64) string {
	return "$" + strconv.FormatFloat(f, 'f', 2, 64)
}
