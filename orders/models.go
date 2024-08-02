package orders

import (
	"time"
)

type Order struct {
	ID         int                 `json:"id"`
	UserID     int                 `json:"userId"`
	Name       string              `json:"name"`
	CreatedAt  time.Time           `json:"createdAt"`
	Processing bool                `json:"processing"`
	Sold       bool                `json:"sold"`
	Dishes     []OrderDishRelation `json:"dishes"`
}

type OrderDishRelation struct {
	OrderID  int `json:"orderId"`
	DishID   int `json:"dishId"`
	Quantity int `json:"quantity"`
}

type OrderDishRelationView struct {
	DishID   int    `json:"dishId"`
	DishName string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    string `json:"price"`
}

type OrderView struct {
	ID         int                     `json:"id"`
	UserID     int                     `json:"userId"`
	Name       string                  `json:"name"`
	TotalPrice string                  `json:"TotalPrice"`
	Dishes     []OrderDishRelationView `json:"dishes"`
}
