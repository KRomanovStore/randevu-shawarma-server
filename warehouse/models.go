package warehouse

type WarehouseItem struct {
	ID           int     `json:"id"`
	ProductID    int     `json:"productId"`
	ProductName  string  `json:"productName"`
	CurrentStock float64 `json:"currentStock"`
	AverageCost  string  `json:"averageCost"`
}
