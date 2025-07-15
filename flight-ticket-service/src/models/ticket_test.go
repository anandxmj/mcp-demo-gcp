package models

import (
	"testing"
	"time"
)

func TestGenerateConfirmationID(t *testing.T) {
	id := GenerateConfirmationID()
	
	if len(id) != 6 {
		t.Errorf("Expected confirmation ID length of 6, got %d", len(id))
	}
	
	// Test that multiple calls generate different IDs
	id2 := GenerateConfirmationID()
	if id == id2 {
		t.Errorf("Expected different confirmation IDs, got same: %s", id)
	}
}

func TestGenerateFlightNumber(t *testing.T) {
	flightNum := GenerateFlightNumber("AA")
	
	if len(flightNum) < 5 || len(flightNum) > 6 {
		t.Errorf("Expected flight number length between 5-6, got %d", len(flightNum))
	}
	
	if flightNum[:2] != "AA" {
		t.Errorf("Expected flight number to start with 'AA', got %s", flightNum[:2])
	}
}

func TestValidateAirportCode(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		{"JFK", true},
		{"LAX", true},
		{"ORD", true},
		{"jfk", false}, // lowercase
		{"JFKX", false}, // too long
		{"JF", false},   // too short
		{"", false},     // empty
	}
	
	for _, test := range tests {
		result := ValidateAirportCode(test.code)
		if result != test.expected {
			t.Errorf("ValidateAirportCode(%s) = %v, expected %v", test.code, result, test.expected)
		}
	}
}

func TestNewFlightTicket(t *testing.T) {
	departureDate := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
	departureTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	
	ticket := NewFlightTicket("JFK", "LAX", departureDate, departureTime, "AA1234", 2)
	
	if ticket == nil {
		t.Fatal("Expected ticket to be created, got nil")
	}
	
	if ticket.Origin != "JFK" {
		t.Errorf("Expected origin 'JFK', got %s", ticket.Origin)
	}
	
	if ticket.Destination != "LAX" {
		t.Errorf("Expected destination 'LAX', got %s", ticket.Destination)
	}
	
	if ticket.FlightNumber != "AA1234" {
		t.Errorf("Expected flight number 'AA1234', got %s", ticket.FlightNumber)
	}
	
	if ticket.Passengers != 2 {
		t.Errorf("Expected 2 passengers, got %d", ticket.Passengers)
	}
	
	if ticket.Status != "CONFIRMED" {
		t.Errorf("Expected status 'CONFIRMED', got %s", ticket.Status)
	}
	
	if len(ticket.ConfirmationID) != 6 {
		t.Errorf("Expected confirmation ID length of 6, got %d", len(ticket.ConfirmationID))
	}
}

func TestNewFlightTicketInvalidAirportCode(t *testing.T) {
	departureDate := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
	departureTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	
	// Test invalid origin
	ticket := NewFlightTicket("INVALID", "LAX", departureDate, departureTime, "AA1234", 2)
	if ticket != nil {
		t.Error("Expected nil ticket for invalid origin airport code")
	}
	
	// Test invalid destination
	ticket = NewFlightTicket("JFK", "invalid", departureDate, departureTime, "AA1234", 2)
	if ticket != nil {
		t.Error("Expected nil ticket for invalid destination airport code")
	}
}
