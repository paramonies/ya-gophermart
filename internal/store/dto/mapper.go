package dto

import "fmt"

type OrderStatus string

const (
	OrderNew        OrderStatus = "NEW"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderInvalid    OrderStatus = "INVALID"
	OrderProcessed  OrderStatus = "PROCESSED"
)

func OrderStatusToStore(status string) (OrderStatus, error) {
	switch status {
	case "REGISTERED":
		return OrderNew, nil
	case "INVALID":
		return OrderInvalid, nil
	case "PROCESSING":
		return OrderProcessing, nil
	case "PROCESSED":
		return OrderProcessed, nil
	default:
		return "", fmt.Errorf("status %s for order not found", status)
	}
}
