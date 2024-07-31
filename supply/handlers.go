package supply

import (
	"encoding/json"
	"fmt"
)

// Custom UnmarshalJSON to enforce price as a string
func (spr *SupplyProductRelation) UnmarshalJSON(data []byte) error {
	type Alias SupplyProductRelation
	aux := &struct {
		Price interface{} `json:"price"`
		*Alias
	}{
		Alias: (*Alias)(spr),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.Price.(type) {
	case string:
		spr.Price = v
	case float64:
		spr.Price = fmt.Sprintf("$%.2f", v)
	case int:
		spr.Price = fmt.Sprintf("$%d.00", v)
	default:
		return fmt.Errorf("unexpected type for price field: %T", v)
	}

	return nil
}
