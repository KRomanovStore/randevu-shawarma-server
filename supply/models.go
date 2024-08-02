package supply

import (
	"time"
)

type Supply struct {
	ID        int                     `json:"id"`
	UserID    int                     `json:"userId"`
	CreatedAt time.Time               `json:"createdAt"`
	Products  []SupplyProductRelation `json:"products"`
}

type SupplyProductRelation struct {
	SupplyID  int     `json:"supplyId"`
	ProductID int     `json:"productId"`
	Quantity  float64 `json:"quantity"`
	Price     string  `json:"price"`
}
