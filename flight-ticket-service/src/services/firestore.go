package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"flight-ticket-service/src/models"
	"google.golang.org/api/option"
)

type FirestoreService struct {
	client     *firestore.Client
	collection string
}

// NewFirestoreService creates a new Firestore service instance
func NewFirestoreService(projectID string, credentialsPath string) (*FirestoreService, error) {
	ctx := context.Background()
	
	var client *firestore.Client
	var err error
	
	if credentialsPath != "" {
		// Use service account key file
		client, err = firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
	} else {
		// Use default credentials (ADC)
		client, err = firestore.NewClient(ctx, projectID)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %v", err)
	}

	return &FirestoreService{
		client:     client,
		collection: "flight_tickets",
	}, nil
}

// CreateTicket creates a new flight ticket in Firestore
func (fs *FirestoreService) CreateTicket(ctx context.Context, ticket *models.FlightTicket) error {
	_, err := fs.client.Collection(fs.collection).Doc(ticket.ConfirmationID).Set(ctx, ticket)
	if err != nil {
		return fmt.Errorf("failed to create ticket: %v", err)
	}
	
	log.Printf("Created ticket with confirmation ID: %s", ticket.ConfirmationID)
	return nil
}

// GetTicket retrieves a flight ticket by confirmation ID
func (fs *FirestoreService) GetTicket(ctx context.Context, confirmationID string) (*models.FlightTicket, error) {
	doc, err := fs.client.Collection(fs.collection).Doc(confirmationID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %v", err)
	}
	
	var ticket models.FlightTicket
	if err := doc.DataTo(&ticket); err != nil {
		return nil, fmt.Errorf("failed to parse ticket data: %v", err)
	}
	
	return &ticket, nil
}

// UpdateTicket updates an existing flight ticket
func (fs *FirestoreService) UpdateTicket(ctx context.Context, confirmationID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	
	// Build the update array
	var updateArray []firestore.Update
	for field, value := range updates {
		updateArray = append(updateArray, firestore.Update{
			Path:  field,
			Value: value,
		})
	}
	
	_, err := fs.client.Collection(fs.collection).Doc(confirmationID).Update(ctx, updateArray)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %v", err)
	}
	
	log.Printf("Updated ticket with confirmation ID: %s", confirmationID)
	return nil
}

// DeleteTicket deletes a flight ticket (or marks as cancelled)
func (fs *FirestoreService) DeleteTicket(ctx context.Context, confirmationID string) error {
	// Instead of deleting, we'll mark as cancelled for audit purposes
	updates := map[string]interface{}{
		"status":     "CANCELLED",
		"updated_at": time.Now(),
	}
	
	return fs.UpdateTicket(ctx, confirmationID, updates)
}

// ListTickets retrieves all flight tickets (with optional filtering)
func (fs *FirestoreService) ListTickets(ctx context.Context, limit int) ([]*models.FlightTicket, error) {
	query := fs.client.Collection(fs.collection).OrderBy("created_at", firestore.Desc)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list tickets: %v", err)
	}
	
	var tickets []*models.FlightTicket
	for _, doc := range docs {
		var ticket models.FlightTicket
		if err := doc.DataTo(&ticket); err != nil {
			log.Printf("Failed to parse ticket %s: %v", doc.Ref.ID, err)
			continue
		}
		tickets = append(tickets, &ticket)
	}
	
	return tickets, nil
}

// Close closes the Firestore client
func (fs *FirestoreService) Close() error {
	return fs.client.Close()
}
