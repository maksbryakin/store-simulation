// simulation/simulation.go
package simulation

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/maksbryakin/store-simulation/database"
	"github.com/maksbryakin/store-simulation/logger"
	"github.com/maksbryakin/store-simulation/models"
)

// Department представляет отдел магазина
type Department struct {
	Name     string
	Position models.Position
	Width    int
	Height   int
}

// Store представляет магазин
type Store struct {
	Departments     []Department
	Customers       []*models.Customer
	CustomerChannel chan *models.Customer
	Mutex           sync.Mutex
	Logs            []string
	GoroutineCount  int
	GoroutineMutex  sync.Mutex
	ChannelCount    int
	ChannelMutex    sync.Mutex
}

// NewStore инициализирует новый магазин и загружает покупателей из базы данных
func NewStore(departments []Department) *Store {
	store := &Store{
		Departments:     departments,
		CustomerChannel: make(chan *models.Customer, 1000),
		Customers:       []*models.Customer{},
		Logs:            []string{},
	}

	// Загрузка покупателей из базы данных
	store.LoadCustomersFromDB()

	return store
}

// LoadCustomersFromDB загружает покупателей из базы данных и добавляет их в канал
func (s *Store) LoadCustomersFromDB() {
	rows, err := database.DB.Query(context.Background(), `
        SELECT id, first_name, last_name, purchase_amount, purchase_count, visit_count, category
        FROM customers
    `)
	if err != nil {
		s.LogMessage(fmt.Sprintf("Ошибка загрузки покупателей из базы данных: %v", err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var c models.Customer
		err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.PurchaseAmount, &c.PurchaseCount, &c.VisitCount, &c.DesiredProduct.Category)
		if err != nil {
			s.LogMessage(fmt.Sprintf("Ошибка чтения строки покупателя: %v", err))
			continue
		}

		// Определение начальной позиции (вход в магазин)
		c.CurrentPosition = models.Position{X: 50, Y: 500}
		s.Customers = append(s.Customers, &c)
		s.CustomerChannel <- &c
	}

	s.LogMessage(fmt.Sprintf("Загружено %d покупателей из базы данных", len(s.Customers)))
}

// AddCustomer добавляет покупателя в магазин и канал
func (s *Store) AddCustomer(c *models.Customer) {
	s.Mutex.Lock()
	s.Customers = append(s.Customers, c)
	s.Mutex.Unlock()
	s.CustomerChannel <- c
	s.LogMessage(fmt.Sprintf("Добавлен покупатель %d", c.ID))
}

// RemoveCustomer удаляет покупателя из магазина по ID
func (s *Store) RemoveCustomer(id int) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	for i, customer := range s.Customers {
		if customer.ID == id {
			s.Customers = append(s.Customers[:i], s.Customers[i+1:]...)
			s.LogMessage(fmt.Sprintf("Удалён покупатель %d", id))
			break
		}
	}
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
		s.IncrementGoroutine()
		go func(c *models.Customer) {
			defer s.DecrementGoroutine()
			s.HandleCustomer(c)
		}(customer)
	}
}

// HandleCustomer обрабатывает поведение покупателя
func (s *Store) HandleCustomer(c *models.Customer) {
	s.LogMessage(fmt.Sprintf("Покупатель %d начал движение к %s", c.ID, c.DesiredProduct.Category))
	departmentPosition := s.GetDepartmentPosition(c.DesiredProduct.Category)
	exitPosition := models.Position{X: 750, Y: 575} // Позиция выхода

	// Перемещение к отделу
	s.MoveTowards(c, departmentPosition)

	// Симуляция покупки продукта
	time.Sleep(2 * time.Second)
	s.LogMessage(fmt.Sprintf("Покупатель %d совершил покупку и направляется к выходу", c.ID))

	// Перемещение к выходу
	s.MoveTowards(c, exitPosition)
	s.LogMessage(fmt.Sprintf("Покупатель %d вышел из магазина", c.ID))

	// Удаление клиента из списка после выхода
	s.RemoveCustomer(c.ID)
}

// MoveTowards перемещает покупателя к целевой позиции
func (s *Store) MoveTowards(c *models.Customer, target models.Position) {
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
			distance := CalculateDistance(c.CurrentPosition, other.CurrentPosition)
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

// CalculateDistance рассчитывает расстояние между двумя точками
func CalculateDistance(p1, p2 models.Position) float64 {
	dx := float64(p1.X - p2.X)
	dy := float64(p1.Y - p2.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// GetDepartmentPosition возвращает позицию отдела по категории
func (s *Store) GetDepartmentPosition(category string) models.Position {
	for _, dept := range s.Departments {
		if contains(dept.Name, category) {
			return models.Position{
				X: dept.Position.X + dept.Width/2,
				Y: dept.Position.Y + dept.Height/2,
			}
		}
	}
	// Если отдел не найден, направляемся к выходу
	return models.Position{X: 750, Y: 575}
}

// contains проверяет, содержит ли строка str подстроку substr (без учёта регистра)
func contains(str, substr string) bool {
	return len(substr) == 0 || (len(str) >= len(substr) && (str == substr || contains(str[1:], substr)))
}

// LogMessage добавляет сообщение в лог
func (s *Store) LogMessage(message string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	s.Logs = append(s.Logs, logEntry)
	if len(s.Logs) > 1000 { // Ограничение на количество логов
		s.Logs = s.Logs[1:]
	}

	logger.Logger.Info("Лог сообщения", zap.String("message", logEntry))
}

// GetTechnicalLogs возвращает последние N логов
func (s *Store) GetTechnicalLogs() []string {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if len(s.Logs) < 10 {
		return s.Logs
	}
	return s.Logs[len(s.Logs)-10:]
}

// Goroutine счетчики

func (s *Store) IncrementGoroutine() {
	s.GoroutineMutex.Lock()
	s.GoroutineCount++
	s.GoroutineMutex.Unlock()
}

func (s *Store) DecrementGoroutine() {
	s.GoroutineMutex.Lock()
	if s.GoroutineCount > 0 {
		s.GoroutineCount--
	}
	s.GoroutineMutex.Unlock()
}

func (s *Store) GetGoroutineCount() int {
	s.GoroutineMutex.Lock()
	defer s.GoroutineMutex.Unlock()
	return s.GoroutineCount
}

// Каналы счетчики

func (s *Store) IncrementChannel() {
	s.ChannelMutex.Lock()
	s.ChannelCount++
	s.ChannelMutex.Unlock()
}

func (s *Store) DecrementChannel() {
	s.ChannelMutex.Lock()
	if s.ChannelCount > 0 {
		s.ChannelCount--
	}
	s.ChannelMutex.Unlock()
}

func (s *Store) GetChannelCount() int {
	s.ChannelMutex.Lock()
	defer s.ChannelMutex.Unlock()
	return s.ChannelCount
}
