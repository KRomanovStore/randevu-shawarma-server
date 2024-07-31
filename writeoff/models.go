package writeoff

import (
	"time"
)

type WriteOff struct {
	ID        int                       `json:"id"`
	UserID    int                       `json:"user_id"`
	CreatedAt time.Time                 `json:"created_at"`
	Notes     string                    `json:"notes"`
	Products  []WriteOffProductRelation `json:"products"`
}

type WriteOffProductRelation struct {
	WriteOffID int     `json:"write_off_id"`
	ProductID  int     `json:"product_id"`
	Quantity   float64 `json:"quantity"`
}
