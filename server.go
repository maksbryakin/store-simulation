// server.go
package main

import (
	"encoding/json"
	"github.com/maksbryakin/store-simulation/models"
	"github.com/maksbryakin/store-simulation/simulation"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Структуры данных для обмена с клиентом
type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type StartData struct {
	CustomerCount int `json:"customerCount"`
	Products      struct {
		Milk       int `json:"Milk"`
		Vegetables int `json:"Vegetables"`
		Meat       int `json:"Meat"`
		Bread      int `json:"Bread"`
		Sugar      int `json:"Sugar"`
	} `json:"products"`
}

type CustomerData struct {
	ID              int                   `json:"ID"`
	CurrentPosition models.Position       `json:"CurrentPosition"`
	DesiredProduct  models.DesiredProduct `json:"DesiredProduct"`
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешить все источники (для разработки)
	},
}

var clients = make(map[*websocket.Conn]bool)
var mutex = sync.Mutex{}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Обновляем соединение до WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка обновления WebSocket:", err)
		return
	}
	defer ws.Close()

	// Регистрация клиента
	mutex.Lock()
	clients[ws] = true
	mutex.Unlock()

	log.Println("Новый клиент подключён")

	for {
		var msg Message
		// Читаем сообщение от клиента
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("Ошибка чтения сообщения:", err)
			mutex.Lock()
			delete(clients, ws)
			mutex.Unlock()
			break
		}

		if msg.Action == "start" {
			var startData StartData
			err := json.Unmarshal(msg.Data, &startData)
			if err != nil {
				log.Println("Ошибка разбора JSON данных start:", err)
				continue
			}

			log.Printf("Запуск симуляции с %d покупателями и продуктами: %+v", startData.CustomerCount, startData.Products)

			// Инициализация продуктов и магазина
			products := []models.Product{
				{Category: "Milk", Quantity: startData.Products.Milk},
				{Category: "Vegetables", Quantity: startData.Products.Vegetables},
				{Category: "Meat", Quantity: startData.Products.Meat},
				{Category: "Bread", Quantity: startData.Products.Bread},
				{Category: "Sugar", Quantity: startData.Products.Sugar},
			}

			store := simulation.NewStore(products)

			// Создание покупателей
			for i := 0; i < startData.CustomerCount; i++ {
				desiredCategory := getRandomCategory(startData.Products)
				customer := simulation.NewCustomer(i+1, desiredCategory)
				store.AddCustomer(customer)
			}

			// Запуск симуляции
			go store.StartSimulation()

			// Отправка обновлений клиенту
			go func(ws *websocket.Conn, store *simulation.Store) {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()

				for {
					<-ticker.C

					customers := store.GetCustomers()
					if len(customers) == 0 {
						// Симуляция закончена, закрыть соединение
						ws.WriteJSON([]CustomerData{})
						ws.Close()
						log.Println("Все покупатели ушли. Соединение закрыто.")
						return
					}

					// Преобразуем покупателей в структуру для отправки клиенту
					var customerData []CustomerData
					for _, c := range customers {
						customerData = append(customerData, CustomerData{
							ID:              c.ID,
							CurrentPosition: c.CurrentPosition,
							DesiredProduct:  c.DesiredProduct,
						})
					}

					// Отправляем данные клиенту
					err := ws.WriteJSON(customerData)
					if err != nil {
						log.Println("Ошибка отправки сообщения клиенту:", err)
						mutex.Lock()
						delete(clients, ws)
						mutex.Unlock()
						return
					}

					// Логирование отправки данных
					log.Printf("Отправлено %d клиентов клиенту", len(customerData))
				}
			}(ws, store)
		}
	}
}

func getRandomCategory(products struct {
	Milk       int `json:"Milk"`
	Vegetables int `json:"Vegetables"`
	Meat       int `json:"Meat"`
	Bread      int `json:"Bread"`
	Sugar      int `json:"Sugar"`
}) string {
	categories := []string{}
	for category, count := range map[string]int{
		"Milk":       products.Milk,
		"Vegetables": products.Vegetables,
		"Meat":       products.Meat,
		"Bread":      products.Bread,
		"Sugar":      products.Sugar,
	} {
		for i := 0; i < count; i++ {
			categories = append(categories, category)
		}
	}
	if len(categories) == 0 {
		return "Unknown"
	}
	return categories[rand.Intn(len(categories))]
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/ws", handleConnections)

	// Обслуживание статических файлов
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Println("Сервер запущен на :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
