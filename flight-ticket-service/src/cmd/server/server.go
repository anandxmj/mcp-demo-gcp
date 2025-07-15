// @title Flight Ticket Service API
// @version 1.0
// @description A Go-based REST API service for managing flight tickets using Google Cloud Firestore
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @tag.name tickets
// @tag.description Flight ticket management operations

// @tag.name health
// @tag.description Health check operations

package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flight-ticket-service/src/handlers"
	"flight-ticket-service/src/services"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "flight-ticket-service/docs" // Import generated docs
)

func main() {
	// Initialize random seed for confirmation ID generation
	rand.Seed(time.Now().UnixNano())

	// Get environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT environment variable is required")
	}

	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	// Initialize Firestore service
	firestoreService, err := services.NewFirestoreService(projectID, credentialsPath)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore service: %v", err)
	}
	defer firestoreService.Close()

	// Initialize handlers
	ticketHandler := handlers.NewTicketHandler(firestoreService)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, specify your frontend domains
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Health check endpoint
	r.Get("/health", handlers.HealthCheck)

	// Root endpoint
	// @Summary API Information
	// @Description Get basic information about the Flight Ticket Service API
	// @Tags health
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]string "API information"
	// @Router / [get]
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Called /")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Flight Ticket Service API", "version": "1.0.0", "swagger": "/swagger/"}`))
	})

	// Swagger documentation endpoint
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // Use relative URL for Cloud Run compatibility
	))

	// Ticket endpoints
	r.Route("/ticket", func(r chi.Router) {
		r.Post("/", ticketHandler.CreateTicket)                        // Create new ticket
		r.Get("/{confirmationID}", ticketHandler.GetTicket)            // Get ticket by confirmation ID
		r.Put("/{confirmationID}", ticketHandler.UpdateTicket)         // Update ticket
		r.Delete("/{confirmationID}", ticketHandler.DeleteTicket)      // Cancel ticket
	})

	// List all tickets endpoint
	r.Get("/tickets", ticketHandler.ListTickets)

	// Start server
	go func() {
		log.Printf("Flight Ticket Service starting on port %s", port)
		log.Printf("Using Firestore in project: %s", projectID)
		log.Printf("Firestore region: us-east1")
		
		err := http.ListenAndServe(":"+port, r)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("Server Started on PORT %s", port)
	log.Println("API Endpoints:")
	log.Println("  POST   /ticket              - Create new flight ticket")
	log.Println("  GET    /ticket/{id}         - Get flight ticket by confirmation ID")
	log.Println("  PUT    /ticket/{id}         - Update flight ticket")
	log.Println("  DELETE /ticket/{id}         - Cancel flight ticket")
	log.Println("  GET    /tickets             - List all flight tickets")
	log.Println("  GET    /health              - Health check")
	log.Printf("  GET    /swagger/            - Swagger UI documentation")
	log.Printf("  GET    /swagger/doc.json    - OpenAPI specification")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	
	log.Println("Server shutting down gracefully...")
	
	// Close Firestore connection
	if err := firestoreService.Close(); err != nil {
		log.Printf("Error closing Firestore connection: %v", err)
	}
	
	log.Println("Server shutdown complete")
}
