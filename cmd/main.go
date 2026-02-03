package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"subscription-service/internal/database"
	"subscription-service/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	database.InitDB(os.Getenv("DB_URL"))

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		var s models.Subscription

		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		start := time.Time(s.StartDate)
		var end *time.Time
		if s.EndDate != nil {
			t := time.Time(*s.EndDate)
			end = &t
		}

		err := database.DB.QueryRow(
			"INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES ($1,$2,$3,$4,$5) RETURNING id",
			s.ServiceName, s.Price, s.UserID, start, end,
		).Scan(&s.ID)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(201)
		json.NewEncoder(w).Encode(s)
	})

	// Ручка подсчета суммы
	r.Get("/subscriptions/total", func(w http.ResponseWriter, r *http.Request) {
		// 1. Получаем user_id из параметров запроса (из URL)
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}

		// 2. Делаем запрос в базу: "Дай мне цены всех подписок этого юзера"
		// (В реальном ТЗ тут нужно еще учитывать даты, но для старта хватит суммы)
		rows, err := database.DB.Query("SELECT price FROM subscriptions WHERE user_id = $1", userID)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// 3. Считаем сумму
		totalCost := 0
		for rows.Next() {
			var price int
			if err := rows.Scan(&price); err != nil {
				continue
			}
			totalCost += price
		}

		// 4. Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":    userID,
			"total_cost": totalCost,
			"currency":   "RUB",
		})
	})

	// Ручка удаления подписки по ID
	r.Delete("/subscriptions/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id") // Берем ID из ссылки

		// Удаляем из базы
		res, err := database.DB.Exec("DELETE FROM subscriptions WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Проверяем, была ли удалена запись (или такого ID не было)
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Subscription not found", 404)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Deleted"))
	})

	log.Println("Start on :8080")
	http.ListenAndServe(":8080", r)
}
