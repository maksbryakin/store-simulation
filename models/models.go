// models/models.go
package models

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type DesiredProduct struct {
	Category string `json:"category"`
	// Можно добавить дополнительные поля, если необходимо
}

type Customer struct {
	ID              int            `json:"id"`
	FirstName       string         `json:"first_name"`
	LastName        string         `json:"last_name"`
	PurchaseAmount  float64        `json:"purchase_amount"`
	PurchaseCount   int            `json:"purchase_count"`
	VisitCount      int            `json:"visit_count"`
	DesiredProduct  DesiredProduct `json:"desired_product"`
	CurrentPosition Position       `json:"current_position"`
}

type Product struct {
	Category string `json:"category"`
	Quantity int    `json:"quantity"`
}
