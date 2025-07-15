package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"flight-ticket-service/src/models"
	"flight-ticket-service/src/services"

	"github.com/go-chi/chi/v5"
)

type TicketHandler struct {
	firestoreService *services.FirestoreService
}

func NewTicketHandler(firestoreService *services.FirestoreService) *TicketHandler {
	return &TicketHandler{
		firestoreService: firestoreService,
	}
}

// CreateTicket handles POST /ticket
// @Summary Create a new flight ticket
// @Description Create a new flight ticket with the provided details
// @Tags tickets
// @Accept json
// @Produce json
// @Param ticket body models.CreateTicketRequest true "Ticket creation request"
// @Success 201 {object} models.FlightTicket "Successfully created ticket"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /ticket [post]
func (h *TicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid JSON payload"})
		return
	}

	// Validate required fields
	if req.Origin == "" || req.Destination == "" || req.DepartureDate == "" || req.DepartureTime == "" || req.Passengers <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "Missing required fields",
			Message: "origin, destination, departure_date, departure_time, and passengers are required",
		})
		return
	}

	// Parse date and time
	departureDate, err := time.Parse("2006-01-02", req.DepartureDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "Invalid departure_date format",
			Message: "Use YYYY-MM-DD format",
		})
		return
	}

	// Parse time and combine with date
	timeOnly, err := time.Parse("15:04", req.DepartureTime)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "Invalid departure_time format",
			Message: "Use HH:MM format",
		})
		return
	}

	// Combine date and time into a single timestamp
	departureTime := time.Date(
		departureDate.Year(),
		departureDate.Month(),
		departureDate.Day(),
		timeOnly.Hour(),
		timeOnly.Minute(),
		0, // seconds
		0, // nanoseconds
		time.UTC,
	)

	// Create new ticket
	ticket := models.NewFlightTicket(
		req.Origin,
		req.Destination,
		departureDate,
		departureTime,
		req.FlightNumber,
		req.Passengers,
	)

	if ticket == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "Invalid airport codes",
			Message: "Use 3-letter IATA codes",
		})
		return
	}

	// Save to Firestore
	if err := h.firestoreService.CreateTicket(r.Context(), ticket); err != nil {
		log.Printf("Failed to create ticket: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to create ticket"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

// GetTicket handles GET /ticket/{confirmationID}
// @Summary Get a flight ticket by confirmation ID
// @Description Retrieve a flight ticket using its confirmation ID
// @Tags tickets
// @Accept json
// @Produce json
// @Param confirmationID path string true "Ticket confirmation ID" example("ABC123")
// @Success 200 {object} models.FlightTicket "Successfully retrieved ticket"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 404 {object} models.ErrorResponse "Ticket not found"
// @Router /ticket/{confirmationID} [get]
func (h *TicketHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	confirmationID := chi.URLParam(r, "confirmationID")
	if confirmationID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Confirmation ID is required"})
		return
	}

	ticket, err := h.firestoreService.GetTicket(r.Context(), confirmationID)
	if err != nil {
		log.Printf("Failed to get ticket %s: %v", confirmationID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Ticket not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

// UpdateTicket handles PUT /ticket/{confirmationID}
// @Summary Update a flight ticket
// @Description Update an existing flight ticket with new information
// @Tags tickets
// @Accept json
// @Produce json
// @Param confirmationID path string true "Ticket confirmation ID" example("ABC123")
// @Param ticket body models.UpdateTicketRequest true "Ticket update request"
// @Success 200 {object} models.FlightTicket "Successfully updated ticket"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /ticket/{confirmationID} [put]
func (h *TicketHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	confirmationID := chi.URLParam(r, "confirmationID")
	if confirmationID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Confirmation ID is required"})
		return
	}

	var req models.UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid JSON payload"})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Origin != "" {
		if !models.ValidateAirportCode(req.Origin) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid origin airport code"})
			return
		}
		updates["origin"] = req.Origin
	}

	if req.Destination != "" {
		if !models.ValidateAirportCode(req.Destination) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid destination airport code"})
			return
		}
		updates["destination"] = req.Destination
	}

	if req.DepartureDate != "" {
		departureDate, err := time.Parse("2006-01-02", req.DepartureDate)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Invalid departure_date format",
				Message: "Use YYYY-MM-DD format",
			})
			return
		}
		updates["departure_date"] = departureDate
	}

	if req.DepartureTime != "" {
		// Parse time only
		timeOnly, err := time.Parse("15:04", req.DepartureTime)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Invalid departure_time format",
				Message: "Use HH:MM format",
			})
			return
		}
		
		// If we also have a departure date, combine them
		if req.DepartureDate != "" {
			departureDate, err := time.Parse("2006-01-02", req.DepartureDate)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(models.ErrorResponse{
					Error:   "Invalid departure_date format",
					Message: "Use YYYY-MM-DD format",
				})
				return
			}
			
			departureTime := time.Date(
				departureDate.Year(),
				departureDate.Month(),
				departureDate.Day(),
				timeOnly.Hour(),
				timeOnly.Minute(),
				0, 0, time.UTC,
			)
			updates["departure_time"] = departureTime
		} else {
			// If no date provided, use today's date with the time
			today := time.Now().UTC()
			departureTime := time.Date(
				today.Year(),
				today.Month(),
				today.Day(),
				timeOnly.Hour(),
				timeOnly.Minute(),
				0, 0, time.UTC,
			)
			updates["departure_time"] = departureTime
		}
	}

	if req.FlightNumber != "" {
		updates["flight_number"] = req.FlightNumber
	}

	if req.Passengers > 0 {
		updates["passengers"] = req.Passengers
	}

	if req.Status != "" {
		if req.Status != "CONFIRMED" && req.Status != "CANCELLED" && req.Status != "PENDING" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Invalid status",
				Message: "Use CONFIRMED, CANCELLED, or PENDING",
			})
			return
		}
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "No fields to update"})
		return
	}

	// Update ticket
	if err := h.firestoreService.UpdateTicket(r.Context(), confirmationID, updates); err != nil {
		log.Printf("Failed to update ticket %s: %v", confirmationID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to update ticket"})
		return
	}

	// Get updated ticket
	ticket, err := h.firestoreService.GetTicket(r.Context(), confirmationID)
	if err != nil {
		log.Printf("Failed to get updated ticket %s: %v", confirmationID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Ticket updated but failed to retrieve"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

// DeleteTicket handles DELETE /ticket/{confirmationID}
// @Summary Cancel a flight ticket
// @Description Cancel (soft delete) a flight ticket by setting its status to CANCELLED
// @Tags tickets
// @Accept json
// @Produce json
// @Param confirmationID path string true "Ticket confirmation ID" example("ABC123")
// @Success 200 {object} models.SuccessResponse "Successfully cancelled ticket"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /ticket/{confirmationID} [delete]
func (h *TicketHandler) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	confirmationID := chi.URLParam(r, "confirmationID")
	if confirmationID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Confirmation ID is required"})
		return
	}

	if err := h.firestoreService.DeleteTicket(r.Context(), confirmationID); err != nil {
		log.Printf("Failed to cancel ticket %s: %v", confirmationID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to cancel ticket"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Message:        "Ticket cancelled successfully",
		ConfirmationID: confirmationID,
	})
}

// ListTickets handles GET /tickets
// @Summary List all flight tickets
// @Description Retrieve a list of all flight tickets with optional pagination
// @Tags tickets
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of tickets to return" default(50) example(10)
// @Success 200 {object} models.TicketListResponse "Successfully retrieved tickets"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /tickets [get]
func (h *TicketHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	tickets, err := h.firestoreService.ListTickets(r.Context(), limit)
	if err != nil {
		log.Printf("Failed to list tickets: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve tickets"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TicketListResponse{
		Tickets: tickets,
		Count:   len(tickets),
	})
}
