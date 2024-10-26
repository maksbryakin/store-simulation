// store_layout.go
package simulation

import "github.com/maksbryakin/store-simulation/models"

type StoreLayout struct {
	Sections map[models.ProductCategory]models.Position
}

func NewStoreLayout() *StoreLayout {
	return &StoreLayout{
		Sections: map[models.ProductCategory]models.Position{
			models.Milk:       {X: 10, Y: 10},
			models.Vegetables: {X: 20, Y: 10},
			models.Meat:       {X: 30, Y: 10},
			models.Bread:      {X: 40, Y: 10},
			models.Sugar:      {X: 50, Y: 10},
		},
	}
}

func (sl *StoreLayout) GetSectionPosition(category models.ProductCategory) models.Position {
	return sl.Sections[category]
}
