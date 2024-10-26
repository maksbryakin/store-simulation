// handlers/customers.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/maksbryakin/store-simulation/database"
	"github.com/maksbryakin/store-simulation/logger"
	"github.com/maksbryakin/store-simulation/models"
	"github.com/maksbryakin/store-simulation/simulation"
	"go.uber.org/zap"
)

var storeInstance *simulation.Store

// SetStore устанавливает ссылку на симуляцию магазина
func SetStore(s *simulation.Store) {
	storeInstance = s
}

// AddCustomerRequest структура для добавления покупателя
type AddCustomerRequest struct {
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	PurchaseAmount float64 `json:"purchase_amount"`
	PurchaseCount  int     `json:"purchase_count"`
	VisitCount     int     `json:"visit_count"`
	Category       string  `json:"category"`
}

// AddCustomerHandler обрабатывает добавление покупателей через API
// @Summary Добавление покупателя
// @Description Добавить нового покупателя в базу данных
// @Tags customers
// @Accept json
// @Produce json
// @Param customer body AddCustomerRequest true "Данные покупателя"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/customers [post]
func AddCustomerHandler(w http.ResponseWriter, r *http.Request) {
	var req AddCustomerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Logger.Warn("Некорректные данные при добавлении покупателя", zap.Error(err))
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	// Валидация данных (можно добавить более строгую валидацию)
	if req.FirstName == "" || req.LastName == "" || req.Category == "" {
		logger.Logger.Warn("Недостаточно данных для добавления покупателя")
		http.Error(w, "Недостаточно данных", http.StatusBadRequest)
		return
	}

	// Вставка покупателя в базу данных
	var customerID int
	err = database.DB.QueryRow(context.Background(), `
        INSERT INTO customers (first_name, last_name, purchase_amount, purchase_count, visit_count, category)
        VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
    `, req.FirstName, req.LastName, req.PurchaseAmount, req.PurchaseCount, req.VisitCount, req.Category).Scan(&customerID)

	if err != nil {
		logger.Logger.Error("Ошибка добавления покупателя в базу данных", zap.Error(err))
		http.Error(w, "Ошибка добавления покупателя", http.StatusInternalServerError)
		return
	}

	// Создание покупателя для симуляции
	newCustomer := &models.Customer{
		ID:              customerID,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		PurchaseAmount:  req.PurchaseAmount,
		PurchaseCount:   req.PurchaseCount,
		VisitCount:      req.VisitCount,
		DesiredProduct:  models.DesiredProduct{Category: req.Category},
		CurrentPosition: models.Position{X: 50, Y: 500}, // Начальная позиция (вход)
	}
	storeInstance.AddCustomer(newCustomer)

	logger.Logger.Info("Покупатель добавлен", zap.Int("id", customerID), zap.String("first_name", req.FirstName), zap.String("last_name", req.LastName))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Покупатель добавлен", "id": strconv.Itoa(customerID)})
}

// DeleteCustomerHandler обрабатывает удаление покупателей через API
// @Summary Удаление покупателя
// @Description Удалить покупателя из базы данных по ID
// @Tags customers
// @Accept json
// @Produce json
// @Param id path int true "ID покупателя"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/customers/{id} [delete]
func DeleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Logger.Warn("Некорректный ID покупателя", zap.String("id", idStr))
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	result, err := database.DB.Exec(context.Background(), "DELETE FROM customers WHERE id=$1", id)
	if err != nil {
		logger.Logger.Error("Ошибка удаления покупателя из базы данных", zap.Int("id", id), zap.Error(err))
		http.Error(w, "Ошибка удаления покупателя", http.StatusInternalServerError)
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Logger.Warn("Покупатель не найден", zap.Int("id", id))
		http.Error(w, "Покупатель не найден", http.StatusNotFound)
		return
	}

	// Удаление покупателя из симуляции
	storeInstance.RemoveCustomer(id)

	logger.Logger.Info("Покупатель удалён", zap.Int("id", id))
	json.NewEncoder(w).Encode(map[string]string{"message": "Покупатель удалён"})
}

// GetCustomersHandler обрабатывает получение списка покупателей
// @Summary Получение списка покупателей
// @Description Получить список всех покупателей из базы данных
// @Tags customers
// @Accept json
// @Produce json
// @Success 200 {array} models.Customer
// @Failure 500 {object} map[string]string
// @Router /api/customers [get]
func GetCustomersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(context.Background(), "SELECT id, first_name, last_name, purchase_amount, purchase_count, visit_count, category FROM customers")
	if err != nil {
		logger.Logger.Error("Ошибка получения покупателей из базы данных", zap.Error(err))
		http.Error(w, "Ошибка получения покупателей", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.PurchaseAmount, &c.PurchaseCount, &c.VisitCount, &c.DesiredProduct.Category)
		if err != nil {
			logger.Logger.Error("Ошибка обработки данных покупателя", zap.Error(err))
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			return
		}

		customers = append(customers, c)
	}

	if err := rows.Err(); err != nil {
		logger.Logger.Error("Ошибка чтения строк из базы данных", zap.Error(err))
		http.Error(w, "Ошибка чтения данных", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(customers)
}
