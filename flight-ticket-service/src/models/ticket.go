package models

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// FlightTicket represents a flight ticket with standard airline format
// @Description Flight ticket information
type FlightTicket struct {
	ConfirmationID string    `json:"confirmation_id" firestore:"confirmation_id" example:"ABC123" description:"6-character alphanumeric confirmation ID"`
	Origin         string    `json:"origin" firestore:"origin" example:"JFK" description:"3-letter IATA origin airport code"`
	Destination    string    `json:"destination" firestore:"destination" example:"LAX" description:"3-letter IATA destination airport code"`
	DepartureDate  time.Time `json:"departure_date" firestore:"departure_date" example:"2024-12-25T00:00:00Z" description:"Departure date"`
	DepartureTime  time.Time `json:"departure_time" firestore:"departure_time" example:"2024-01-01T14:30:00Z" description:"Departure time"`
	FlightNumber   string    `json:"flight_number" firestore:"flight_number" example:"AA1234" description:"Flight number in airline format"`
	Passengers     int       `json:"passengers" firestore:"passengers" example:"2" description:"Number of passengers"`
	CreatedAt      time.Time `json:"created_at" firestore:"created_at" example:"2024-07-12T19:00:00Z" description:"Ticket creation timestamp"`
	UpdatedAt      time.Time `json:"updated_at" firestore:"updated_at" example:"2024-07-12T19:00:00Z" description:"Last update timestamp"`
	Status         string    `json:"status" firestore:"status" example:"CONFIRMED" enums:"CONFIRMED,CANCELLED,PENDING" description:"Ticket status"`
}

// CreateTicketRequest represents the request payload for creating a ticket
// @Description Request payload for creating a new flight ticket
type CreateTicketRequest struct {
	Origin        string `json:"origin" example:"JFK" description:"3-letter IATA origin airport code" validate:"required"`
	Destination   string `json:"destination" example:"LAX" description:"3-letter IATA destination airport code" validate:"required"`
	DepartureDate string `json:"departure_date" example:"2024-12-25" description:"Departure date in YYYY-MM-DD format" validate:"required"`
	DepartureTime string `json:"departure_time" example:"14:30" description:"Departure time in HH:MM format" validate:"required"`
	FlightNumber  string `json:"flight_number,omitempty" example:"AA1234" description:"Flight number (optional, will be generated if not provided)"`
	Passengers    int    `json:"passengers" example:"2" description:"Number of passengers" validate:"required,min=1"`
}

// UpdateTicketRequest represents the request payload for updating a ticket
// @Description Request payload for updating an existing flight ticket
type UpdateTicketRequest struct {
	Origin        string `json:"origin,omitempty" example:"JFK" description:"3-letter IATA origin airport code"`
	Destination   string `json:"destination,omitempty" example:"LAX" description:"3-letter IATA destination airport code"`
	DepartureDate string `json:"departure_date,omitempty" example:"2024-12-25" description:"Departure date in YYYY-MM-DD format"`
	DepartureTime string `json:"departure_time,omitempty" example:"14:30" description:"Departure time in HH:MM format"`
	FlightNumber  string `json:"flight_number,omitempty" example:"AA1234" description:"Flight number"`
	Passengers    int    `json:"passengers,omitempty" example:"2" description:"Number of passengers" validate:"min=1"`
	Status        string `json:"status,omitempty" example:"CONFIRMED" enums:"CONFIRMED,CANCELLED,PENDING" description:"Ticket status"`
}

// TicketListResponse represents the response for listing tickets
// @Description Response containing list of tickets
type TicketListResponse struct {
	Tickets []*FlightTicket `json:"tickets" description:"List of flight tickets"`
	Count   int             `json:"count" example:"10" description:"Number of tickets returned"`
}

// ErrorResponse represents an error response
// @Description Error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request" description:"Error message"`
	Message string `json:"message,omitempty" example:"Detailed error description" description:"Detailed error message"`
}

// SuccessResponse represents a success response
// @Description Success response
type SuccessResponse struct {
	Message        string `json:"message" example:"Ticket cancelled successfully" description:"Success message"`
	ConfirmationID string `json:"confirmation_id,omitempty" example:"ABC123" description:"Confirmation ID"`
}

// GenerateConfirmationID generates a standard airline confirmation ID (6 characters alphanumeric)
func GenerateConfirmationID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateFlightNumber generates a standard flight number format (2-letter airline code + 3-4 digit number)
func GenerateFlightNumber(airlineCode string) string {
	if len(airlineCode) != 2 {
		airlineCode = "AA" // Default to American Airlines
	}
	flightNum := rand.Intn(9000) + 1000 // Generate 4-digit number between 1000-9999
	return fmt.Sprintf("%s%d", strings.ToUpper(airlineCode), flightNum)
}

// ValidateAirportCode validates that airport code is 3 letters (IATA format)
func ValidateAirportCode(code string) bool {
	return len(code) == 3 && strings.ToUpper(code) == code
}

// NewFlightTicket creates a new flight ticket with generated confirmation ID
func NewFlightTicket(origin, destination string, departureDate, departureTime time.Time, flightNumber string, passengers int) *FlightTicket {
	now := time.Now()
	
	// Validate airport codes
	if !ValidateAirportCode(origin) || !ValidateAirportCode(destination) {
		return nil
	}
	
	// Generate flight number if not provided
	if flightNumber == "" {
		flightNumber = GenerateFlightNumber("AA")
	}
	
	return &FlightTicket{
		ConfirmationID: GenerateConfirmationID(),
		Origin:         strings.ToUpper(origin),
		Destination:    strings.ToUpper(destination),
		DepartureDate:  departureDate,
		DepartureTime:  departureTime,
		FlightNumber:   strings.ToUpper(flightNumber),
		Passengers:     passengers,
		CreatedAt:      now,
		UpdatedAt:      now,
		Status:         "CONFIRMED",
	}
}
