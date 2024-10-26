// simulation/simulation.go
package simulation

import (
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/maksbryakin/store-simulation/models"
)

// Store представляет магазин
type Store struct {
	Products        []models.Product
	Customers       []*models.Customer
	CustomerChannel chan *models.Customer
	Mutex           sync.Mutex
}

// NewStore инициализирует новый магазин
func NewStore(products []models.Product) *Store {
	return &Store{
		Products:        products,
		CustomerChannel: make(chan *models.Customer, 1000),
		Customers:       []*models.Customer{},
	}
}

// AddCustomer добавляет покупателя в магазин
func (s *Store) AddCustomer(c *models.Customer) {
	s.Mutex.Lock()
	s.Customers = append(s.Customers, c)
	s.Mutex.Unlock()
	s.CustomerChannel <- c
}

// GetCustomers возвращает текущих покупателей
func (s *Store) GetCustomers() []*models.Customer {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	return s.Customers
}

// StartSimulation запускает обработку покупателей
func (s *Store) StartSimulation() {
	for customer := range s.CustomerChannel {
		go s.handleCustomer(customer)
	}
}

// handleCustomer обрабатывает поведение покупателя
func (s *Store) handleCustomer(c *models.Customer) {
	log.Printf("Покупатель %d начал движение к торговому залу", c.ID)
	aislePosition := getAislePosition(models.ProductCategory(c.DesiredProduct.Category))
	exitPosition := models.Position{X: 50, Y: 500} // Позиция Exit на клиенте

	// Перемещение к торговому залу
	s.moveTowards(c, aislePosition)

	// Симуляция покупки продукта
	time.Sleep(2 * time.Second)
	log.Printf("Покупатель %d совершил покупку и направляется к выходу", c.ID)

	// Перемещение к выходу
	s.moveTowards(c, exitPosition)
	log.Printf("Покупатель %d вышел из магазина", c.ID)

	// Удаление клиента из списка после выхода
	s.removeCustomer(c.ID)
	log.Printf("Покупатель %d удалён из списка", c.ID)
}

// moveTowards перемещает покупателя к целевой позиции
func (s *Store) moveTowards(c *models.Customer, target models.Position) {
	steps := 50
	deltaX := float64(target.X-c.CurrentPosition.X) / float64(steps)
	deltaY := float64(target.Y-c.CurrentPosition.Y) / float64(steps)

	for i := 0; i < steps; i++ {
		time.Sleep(50 * time.Millisecond) // Скорость движения

		s.Mutex.Lock()
		c.CurrentPosition.X += int(math.Round(deltaX))
		c.CurrentPosition.Y += int(math.Round(deltaY))

		// Проверка на столкновения (опционально)
		for _, other := range s.Customers {
			if other.ID == c.ID {
				continue
			}
			distance := calculateDistance(c.CurrentPosition, other.CurrentPosition)
			if distance < 30 { // Минимальное расстояние между клиентами
				// Изменение направления движения, чтобы избежать столкновения
				c.CurrentPosition.X += rand.Intn(5) - 2
				c.CurrentPosition.Y += rand.Intn(5) - 2
			}
		}

		// Ограничим позицию в пределах холста (800x600)
		if c.CurrentPosition.X < 0 {
			c.CurrentPosition.X = 0
		}
		if c.CurrentPosition.X > 800 {
			c.CurrentPosition.X = 800
		}
		if c.CurrentPosition.Y < 0 {
			c.CurrentPosition.Y = 0
		}
		if c.CurrentPosition.Y > 600 {
			c.CurrentPosition.Y = 600
		}
		s.Mutex.Unlock()
	}
}

// removeCustomer удаляет покупателя из списка
func (s *Store) removeCustomer(id int) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	for i, customer := range s.Customers {
		if customer.ID == id {
			s.Customers = append(s.Customers[:i], s.Customers[i+1:]...)
			break
		}
	}
}

// calculateDistance рассчитывает расстояние между двумя точками
func calculateDistance(p1, p2 models.Position) float64 {
	dx := float64(p1.X - p2.X)
	dy := float64(p1.Y - p2.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// getAislePosition возвращает позицию торгового зала по категории
func getAislePosition(category models.ProductCategory) models.Position {
	switch category {
	case models.Milk:
		return models.Position{X: 300, Y: 100}
	case models.Vegetables:
		return models.Position{X: 350, Y: 100}
	case models.Meat:
		return models.Position{X: 550, Y: 100}
	case models.Bread:
		return models.Position{X: 750, Y: 100}
	case models.Sugar:
		return models.Position{X: 150, Y: 300}
	default:
		return models.Position{X: 50, Y: 500} // Если категория неизвестна, направляемся к выходу
	}
}

// NewCustomer создаёт нового покупателя
func NewCustomer(id int, category string) *models.Customer {
	return &models.Customer{
		ID:              id,
		CurrentPosition: models.Position{X: rand.Intn(800), Y: rand.Intn(600)},
		DesiredProduct:  models.DesiredProduct{Category: models.ProductCategory(category)},
	}
}
