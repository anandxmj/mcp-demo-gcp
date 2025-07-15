package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status" example:"healthy" description:"Service health status"`
	Timestamp time.Time `json:"timestamp" example:"2024-07-13T05:00:00Z" description:"Health check timestamp"`
	Version   string    `json:"version" example:"1.0.0" description:"API version"`
	Service   string    `json:"service" example:"flight-ticket-service" description:"Service name"`
}

// HealthCheck handles GET /health
// @Summary Health check endpoint
// @Description Check the health status of the Flight Ticket Service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Service:   "flight-ticket-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
