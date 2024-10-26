// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/maksbryakin/store-simulation/docs" // Документация Swagger
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/maksbryakin/store-simulation/database"
	"github.com/maksbryakin/store-simulation/handlers"
	"github.com/maksbryakin/store-simulation/logger"
	"github.com/maksbryakin/store-simulation/metrics"
	"github.com/maksbryakin/store-simulation/models"
	"github.com/maksbryakin/store-simulation/simulation"

	"go.uber.org/zap"
)

// @title Store Simulation API
// @version 1.0
// @description API для симуляции магазина
// @host localhost:8080
// @BasePath /

// Структуры данных для обмена с клиентом
type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type StartData struct {
	CustomerCount int `json:"customerCount"`
}

type CustomerData struct {
	ID              int                   `json:"id"`
	CurrentPosition models.Position       `json:"current_position"`
	DesiredProduct  models.DesiredProduct `json:"desired_product"`
}

type SimulationStats struct {
	Customers            []CustomerData `json:"customers"`
	AveragePurchaseCount float64        `json:"average_purchase_count"`
	CustomerCategories   map[string]int `json:"customer_categories"`
	StoreLoad            float64        `json:"store_load"`
	GoroutineCount       int            `json:"goroutine_count"`
	ChannelCount         int            `json:"channel_count"`
	TechnicalLogs        []string       `json:"technical_logs"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешить все источники (для разработки)
	},
}

var clients = make(map[*websocket.Conn]bool)
var mutexSync = sync.Mutex{}
var store *simulation.Store

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Обновляем соединение до WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.Error("Ошибка обновления WebSocket", zap.Error(err))
		return
	}
	defer ws.Close()

	// Регистрация клиента
	mutexSync.Lock()
	clients[ws] = true
	mutexSync.Unlock()

	logger.Logger.Info("Новый клиент подключён")

	for {
		var msg Message
		// Читаем сообщение от клиента
		err := ws.ReadJSON(&msg)
		if err != nil {
			logger.Logger.Warn("Ошибка чтения сообщения", zap.Error(err))
			mutexSync.Lock()
			delete(clients, ws)
			mutexSync.Unlock()
			break
		}

		if msg.Action == "start" {
			var startData StartData
			err := json.Unmarshal(msg.Data, &startData)
			if err != nil {
				logger.Logger.Warn("Ошибка разбора JSON данных start", zap.Error(err))
				continue
			}

			logger.Logger.Info("Запуск симуляции", zap.Int("customerCount", startData.CustomerCount))

			// Инициализация отделов
			departments := []simulation.Department{
				{Name: "Молочный отдел", Position: models.Position{X: 200, Y: 100}, Width: 100, Height: 200},
				{Name: "Отдел овощей", Position: models.Position{X: 400, Y: 100}, Width: 100, Height: 200},
				{Name: "Отдел мяса", Position: models.Position{X: 600, Y: 100}, Width: 100, Height: 200},
				{Name: "Отдел хлеба", Position: models.Position{X: 200, Y: 350}, Width: 100, Height: 200},
				{Name: "Отдел сахара", Position: models.Position{X: 400, Y: 350}, Width: 100, Height: 200},
			}

			// Инициализация магазина
			store = simulation.NewStore(departments)
			store.IncrementChannel()

			// Установка ссылки на магазин в обработчики
			handlers.SetStore(store)

			// Создание покупателей
			for i := 0; i < startData.CustomerCount; i++ {
				desiredCategory := getRandomCategory(departments)
				customer := &models.Customer{
					ID:              i + 1,
					DesiredProduct:  models.DesiredProduct{Category: desiredCategory},
					CurrentPosition: models.Position{X: 50, Y: 500}, // Начальная позиция (вход)
				}
				store.AddCustomer(customer)
			}

			// Запуск симуляции
			go store.StartSimulation()

			// Отправка обновлений клиенту
			go func(ws *websocket.Conn, store *simulation.Store) {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						customers := store.GetCustomers()
						if len(customers) == 0 {
							// Симуляция закончена, закрыть соединение
							ws.WriteJSON(SimulationStats{})
							ws.Close()
							logger.Logger.Info("Все покупатели ушли. Соединение закрыто.")
							return
						}

						// Расчёт статистики
						var totalPurchaseCount int
						customerCategories := make(map[string]int)
						for _, c := range customers {
							totalPurchaseCount += c.PurchaseCount
							category := c.DesiredProduct.Category
							customerCategories[category]++
						}
						averagePurchaseCount := 0.0
						if len(customers) > 0 {
							averagePurchaseCount = float64(totalPurchaseCount) / float64(len(customers))
						}

						// Пример загрузки магазина (можно определить свою метрику)
						storeLoad := float64(len(customers)) / 100.0 // Предположим, что вместимость 100 покупателей

						// Получение количества горутин и каналов
						goroutineCount := store.GetGoroutineCount()
						channelCount := store.GetChannelCount()

						// Получение технических логов
						technicalLogs := store.GetTechnicalLogs()

						// Подготовка данных для отправки
						stats := SimulationStats{
							Customers:            []CustomerData{},
							AveragePurchaseCount: averagePurchaseCount,
							CustomerCategories:   customerCategories,
							StoreLoad:            storeLoad,
							GoroutineCount:       goroutineCount,
							ChannelCount:         channelCount,
							TechnicalLogs:        technicalLogs,
						}

						for _, c := range customers {
							stats.Customers = append(stats.Customers, CustomerData{
								ID:              c.ID,
								CurrentPosition: c.CurrentPosition,
								DesiredProduct:  c.DesiredProduct,
							})
						}

						// Отправляем данные клиенту
						err := ws.WriteJSON(stats)
						if err != nil {
							logger.Logger.Warn("Ошибка отправки сообщения клиенту", zap.Error(err))
							mutexSync.Lock()
							delete(clients, ws)
							mutexSync.Unlock()
							return
						}

						// Логирование отправки данных
						logger.Logger.Debug("Отправлено клиентов клиенту", zap.Int("count", len(stats.Customers)))

						// Обновление метрик
						metrics.SetCustomerCount(float64(len(customers)))
						metrics.SetGoroutineCount(float64(goroutineCount))
						metrics.SetChannelCount(float64(channelCount))
					}
				}
			}(ws, store)
		}
	}
}

func getRandomCategory(departments []simulation.Department) string {
	if len(departments) == 0 {
		return "Unknown"
	}
	index := rand.Intn(len(departments))
	return departments[index].Name
}

func main() {
	// Загрузка переменных окружения из .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Не удалось загрузить файл .env")
	}

	// Инициализация логгера
	logger.InitLogger()
	defer logger.Logger.Sync()

	// Инициализация базы данных
	err = database.InitDB()
	if err != nil {
		logger.Logger.Fatal("Не удалось подключиться к базе данных", zap.Error(err))
	}
	defer database.DB.Close() // Закрытие соединения при завершении работы приложения

	// Автоматическое создание таблиц, если они не существуют
	createTablesIfNotExist()

	// Инициализация метрик
	metrics.InitMetrics()
	metrics.StartMetricsServer()

	// Настройка маршрутов
	router := mux.NewRouter()
	router.HandleFunc("/ws", handleConnections)
	router.HandleFunc("/api/register", handlers.RegisterHandler).Methods("POST")
	router.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/api/customers", handlers.AddCustomerHandler).Methods("POST")
	router.HandleFunc("/api/customers/{id}", handlers.DeleteCustomerHandler).Methods("DELETE")
	router.HandleFunc("/api/customers", handlers.GetCustomersHandler).Methods("GET")
	router.HandleFunc("/api/accidents/start", StartAccidentsHandler).Methods("POST")
	router.HandleFunc("/api/accidents/stop", StopAccidentsHandler).Methods("POST")

	// Обслуживание Swagger документации
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Обслуживание статических файлов
	fs := http.FileServer(http.Dir("./static"))
	router.PathPrefix("/").Handler(fs)

	port := "8080"
	logger.Logger.Info("Сервер запущен", zap.String("port", port))
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		logger.Logger.Fatal("ListenAndServe ошибка", zap.Error(err))
	}
}

// createTablesIfNotExist создает необходимые таблицы, если они не существуют
func createTablesIfNotExist() {
	queries := []string{
		// Таблица пользователей
		`CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(50) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL
        );`,
		// Таблица покупателей
		`CREATE TABLE IF NOT EXISTS customers (
            id SERIAL PRIMARY KEY,
            first_name VARCHAR(50) NOT NULL,
            last_name VARCHAR(50) NOT NULL,
            purchase_amount DECIMAL(10,2) NOT NULL,
            purchase_count INTEGER NOT NULL,
            visit_count INTEGER NOT NULL,
            category VARCHAR(50) NOT NULL
        );`,
	}

	for _, query := range queries {
		_, err := database.DB.Exec(context.Background(), query)
		if err != nil {
			logger.Logger.Fatal("Ошибка создания таблицы", zap.Error(err))
		}
	}

	logger.Logger.Info("Таблицы в базе данных созданы или уже существуют")
}

// Аварийный функционал

var accidentFrequency = 10 * time.Second
var accidentTicker *time.Ticker
var accidentQuit chan struct{}

// StartAccidentsHandler запускает генерацию аварий
// @Summary Запуск аварий
// @Description Запустить автоматическую генерацию аварий
// @Tags accidents
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/accidents/start [post]
func StartAccidentsHandler(w http.ResponseWriter, r *http.Request) {
	if accidentTicker != nil {
		logger.Logger.Warn("Аварии уже запущены")
		http.Error(w, "Аварии уже запущены", http.StatusBadRequest)
		return
	}

	accidentTicker = time.NewTicker(accidentFrequency)
	accidentQuit = make(chan struct{})

	go func() {
		for {
			select {
			case <-accidentTicker.C:
				triggerAccident()
			case <-accidentQuit:
				return
			}
		}
	}()

	logger.Logger.Info("Аварии запущены")
	json.NewEncoder(w).Encode(map[string]string{"message": "Аварии запущены"})
}

// StopAccidentsHandler останавливает генерацию аварий
// @Summary Остановка аварий
// @Description Остановить автоматическую генерацию аварий
// @Tags accidents
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/accidents/stop [post]
func StopAccidentsHandler(w http.ResponseWriter, r *http.Request) {
	if accidentTicker == nil {
		logger.Logger.Warn("Аварии не запущены")
		http.Error(w, "Аварии не запущены", http.StatusBadRequest)
		return
	}

	accidentTicker.Stop()
	close(accidentQuit)
	accidentTicker = nil

	logger.Logger.Info("Аварии остановлены")
	json.NewEncoder(w).Encode(map[string]string{"message": "Аварии остановлены"})
}

// triggerAccident генерирует случайную аварию и влияет на симуляцию
func triggerAccident() {
	if store == nil {
		logger.Logger.Warn("Симуляция не запущена")
		return
	}

	accident := generateRandomAccident()
	store.LogMessage(accident)

	// Влияние на симуляцию, например, уменьшение числа покупателей на 10%
	store.Mutex.Lock()
	affected := int(0.1 * float64(len(store.Customers)))
	for i := 0; i < affected; i++ {
		if len(store.Customers) == 0 {
			break
		}
		c := store.Customers[rand.Intn(len(store.Customers))]
		store.RemoveCustomer(c.ID)
	}
	store.Mutex.Unlock()

	store.LogMessage(fmt.Sprintf("Авария применена: %s. Удалено %d покупателей.", accident, affected))
}

// generateRandomAccident генерирует описание случайной аварии
func generateRandomAccident() string {
	types := []string{
		"Пожар",
		"Взлом",
		"Электросбой",
		"Наводнение",
		"Засорение канализации",
		"Несчастный случай",
		"Сбой системы охлаждения",
		"Перегрузка сети",
		"Проблемы с безопасностью",
		"Сбой в кассовой системе",
	}
	descriptions := []string{
		"В магазине произошёл пожар в зоне овощей.",
		"В магазин попытались проникнуть мошенники.",
		"Произошёл электросбой, временно закрыты некоторые отделы.",
		"В магазине началось наводнение из-за протечки.",
		"Засорение канализации вызвало неприятные запахи.",
		"Несчастный случай на кассе, требуется помощь.",
		"Система охлаждения вышла из строя, продукты начинают портиться.",
		"Перегрузка электрической сети привела к отключению света.",
		"Проблемы с системой безопасности, некоторые камеры не работают.",
		"Сбой в кассовой системе, невозможно провести оплату.",
	}
	index := rand.Intn(len(types))
	return fmt.Sprintf("Авария: %s - %s", types[index], descriptions[index])
}
