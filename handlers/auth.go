// handlers/auth.go
package handlers

import (
	"context"
	"encoding/json"
	"github.com/maksbryakin/store-simulation/database"
	"github.com/maksbryakin/store-simulation/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// RegisterRequest структура для регистрации
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterHandler обрабатывает регистрацию пользователей
// @Summary Регистрация пользователя
// @Description Зарегистрировать нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Данные пользователя"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Logger.Warn("Некорректные данные при регистрации", zap.Error(err))
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	// Валидация данных
	if req.Username == "" || req.Password == "" {
		logger.Logger.Warn("Недостаточно данных для регистрации")
		http.Error(w, "Недостаточно данных", http.StatusBadRequest)
		return
	}

	// Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Logger.Error("Ошибка хэширования пароля", zap.Error(err))
		http.Error(w, "Ошибка регистрации", http.StatusInternalServerError)
		return
	}

	// Вставка пользователя в базу данных
	_, err = database.DB.Exec(context.Background(), `
        INSERT INTO users (username, password_hash)
        VALUES ($1, $2)
    `, req.Username, string(hashedPassword))
	if err != nil {
		logger.Logger.Error("Ошибка добавления пользователя в базу данных", zap.Error(err))
		http.Error(w, "Ошибка регистрации", http.StatusInternalServerError)
		return
	}

	logger.Logger.Info("Пользователь зарегистрирован", zap.String("username", req.Username))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь зарегистрирован"})
}

// LoginRequest структура для логина
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler обрабатывает логин пользователей
// @Summary Логин пользователя
// @Description Аутентифицировать пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param user body LoginRequest true "Данные пользователя"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Logger.Warn("Некорректные данные при логине", zap.Error(err))
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	// Валидация данных
	if req.Username == "" || req.Password == "" {
		logger.Logger.Warn("Недостаточно данных для логина")
		http.Error(w, "Недостаточно данных", http.StatusBadRequest)
		return
	}

	// Получение хэшированного пароля из базы данных
	var hashedPassword string
	err = database.DB.QueryRow(context.Background(), `
        SELECT password_hash FROM users WHERE username=$1
    `, req.Username).Scan(&hashedPassword)
	if err != nil {
		logger.Logger.Warn("Пользователь не найден", zap.String("username", req.Username), zap.Error(err))
		http.Error(w, "Неверные учётные данные", http.StatusUnauthorized)
		return
	}

	// Сравнение паролей
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		logger.Logger.Warn("Неверный пароль", zap.String("username", req.Username), zap.Error(err))
		http.Error(w, "Неверные учётные данные", http.StatusUnauthorized)
		return
	}

	logger.Logger.Info("Пользователь аутентифицирован", zap.String("username", req.Username))
	json.NewEncoder(w).Encode(map[string]string{"message": "Аутентификация успешна"})
}
