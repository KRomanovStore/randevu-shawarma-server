package writeoff

import (
	"time"
)

type WriteOff struct {
	ID        int                       `json:"id"`
	UserID    int                       `json:"userId"`
	CreatedAt time.Time                 `json:"createdAt"`
	Notes     string                    `json:"notes"`
	Products  []WriteOffProductRelation `json:"products"`
}

type WriteOffProductRelation struct {
	WriteOffID int     `json:"writeOffId"`
	ProductID  int     `json:"productId"`
	Quantity   float64 `json:"quantity"`
}
