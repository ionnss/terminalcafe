package models

type Order struct {
	Items    []OrderItem
	Customer Customer
}

type OrderItem struct {
	Product  Product
	Quantity int
}

func (o *Order) Total() float64 {
	var total float64
	for _, item := range o.Items {
		total += item.Product.Price * float64(item.Quantity)
	}
	return total
}
