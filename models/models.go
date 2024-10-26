// models/models.go
package models

type ProductCategory string

const (
	Milk       ProductCategory = "Milk"
	Vegetables ProductCategory = "Vegetables"
	Meat       ProductCategory = "Meat"
	Bread      ProductCategory = "Bread"
	Sugar      ProductCategory = "Sugar"
	Unknown    ProductCategory = "Unknown"
)

type Product struct {
	Category string `json:"Category"`
	Quantity int    `json:"Quantity"`
}

type Position struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

type DesiredProduct struct {
	Category ProductCategory `json:"Category"`
}

type Customer struct {
	ID              int            `json:"ID"`
	CurrentPosition Position       `json:"CurrentPosition"`
	DesiredProduct  DesiredProduct `json:"DesiredProduct"`
	// Дополнительные поля при необходимости
}
