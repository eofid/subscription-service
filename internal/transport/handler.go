package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"subscription-service/internal/domain"
	"subscription-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	svc *service.SubscriptionService
}

func NewHandler(svc *service.SubscriptionService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/subscriptions", h.create)
	r.Get("/subscriptions/total", h.getTotal)
	r.Put("/subscriptions/{id}", h.update)
	r.Delete("/subscriptions/{id}", h.delete)

	return r
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var sub domain.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	id, err := h.svc.Create(r.Context(), sub)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	sub.ID = id
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) getTotal(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", 400)
		return
	}

	total, err := h.svc.GetTotalCost(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"total":   total,
	})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	var sub domain.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	sub.ID = id

	if err := h.svc.Update(r.Context(), sub); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Subscription not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(`{"status":"updated"}`))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Subscription not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(`{"status":"deleted"}`))
}
