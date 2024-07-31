package orders

import (
	"time"
)

type Order struct {
	ID         int                 `json:"id"`
	UserID     int                 `json:"user_id"`
	Name       string              `json:"name"`
	CreatedAt  time.Time           `json:"created_at"`
	Processing bool                `json:"processing"`
	Sold       bool                `json:"sold"`
	Dishes     []OrderDishRelation `json:"dishes"`
}

type OrderDishRelation struct {
	OrderID  int `json:"order_id"`
	DishID   int `json:"dish_id"`
	Quantity int `json:"quantity"`
}
