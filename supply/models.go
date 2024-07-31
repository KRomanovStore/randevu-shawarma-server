package supply

import (
	"time"
)

type Supply struct {
	ID        int                     `json:"id"`
	UserID    int                     `json:"user_id"`
	CreatedAt time.Time               `json:"created_at"`
	Products  []SupplyProductRelation `json:"products"`
}

type SupplyProductRelation struct {
	SupplyID  int     `json:"supply_id"`
	ProductID int     `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	Price     string  `json:"price"`
}
