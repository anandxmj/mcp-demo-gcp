# Flight Ticket Service

A Go-based REST API service for managing flight tickets using Google Cloud Firestore in the us-east1 region.

## Features

- Create flight tickets with standard airline format
- Retrieve tickets by confirmation ID
- Update existing tickets
- Cancel tickets (soft delete)
- List all tickets with pagination
- Standard airline confirmation IDs (6-character alphanumeric)
- IATA airport codes validation
- Standard flight number formats
- **OpenAPI 3.0 specification generation**
- **Interactive Swagger UI documentation**
- **Comprehensive API documentation with examples**

## Flight Ticket Structure

```json
{
  "confirmation_id": "ABC123",
  "origin": "JFK",
  "destination": "LAX",
  "departure_date": "2024-12-25T00:00:00Z",
  "departure_time": "2024-01-01T14:30:00Z",
  "flight_number": "AA1234",
  "passengers": 2,
  "created_at": "2024-07-12T19:00:00Z",
  "updated_at": "2024-07-12T19:00:00Z",
  "status": "CONFIRMED"
}
```

## Prerequisites

1. Go 1.24.5 or later
2. Google Cloud Project with Firestore enabled
3. Firestore database created in us-east1 region
4. Service account key with Firestore permissions

## Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd flight-ticket-service
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up Google Cloud Firestore**
   - Create a Google Cloud Project
   - Enable Firestore API
   - Create a Firestore database in **us-east1** region
   - Create a service account with Firestore permissions
   - Download the service account key JSON file

4. **Configure environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   export GOOGLE_CLOUD_PROJECT=your-project-id
   export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
   export PORT=8080
   ```

5. **Generate API documentation (optional)**
   ```bash
   make swagger-install  # Install swag CLI tool
   make swagger-gen      # Generate OpenAPI spec and Swagger docs
   ```

6. **Build and run**
   ```bash
   go build -o server src/cmd/server/server.go
   ./server
   ```

## API Documentation

### Swagger UI
Once the server is running, you can access the interactive API documentation at:
- **Swagger UI**: http://localhost:8080/swagger/
- **OpenAPI JSON**: http://localhost:8080/swagger/doc.json

### API Endpoints

#### Create Flight Ticket
```bash
POST /ticket
Content-Type: application/json

{
  "origin": "JFK",
  "destination": "LAX",
  "departure_date": "2024-12-25",
  "departure_time": "14:30",
  "flight_number": "AA1234",
  "passengers": 2
}
```

#### Get Flight Ticket
```bash
GET /ticket/{confirmation_id}
```

#### Update Flight Ticket
```bash
PUT /ticket/{confirmation_id}
Content-Type: application/json

{
  "passengers": 3,
  "status": "CONFIRMED"
}
```

#### Cancel Flight Ticket
```bash
DELETE /ticket/{confirmation_id}
```

#### List All Tickets
```bash
GET /tickets?limit=50
```

#### Health Check
```bash
GET /health
```

## Development Commands

### Using Mage (Recommended)

```bash
# Show all available commands
mage -l

# Local development
mage Build                    # Build Go application locally
mage Run                      # Run application locally on port 6000
mage Clean                    # Clean up build artifacts

# Docker commands
mage DockerBuild             # Build Docker image
mage DockerRun               # Run Docker container locally
mage DockerStop              # Stop running containers
mage DockerPush              # Push image to Artifact Registry

# Cloud Run deployment
mage Setup                   # Setup Artifact Registry (run once)
mage SetupServiceAccount     # Setup service account for Firestore
mage Deploy                  # Deploy to Cloud Run (basic)
mage DeployWithServiceAccount # Deploy with service account (recommended)
mage FullPipeline            # Complete pipeline: Setup -> Build -> Push -> Deploy

# Monitoring and debugging
mage Status                  # Get service URL and status
mage Logs                    # View Cloud Run logs
```

### Using Make (Alternative)

```bash
# Show all available commands
make help

# Complete setup for development
make setup

# Generate OpenAPI documentation
make docs

# Build the application
make build

# Run the application
make run

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Clean build artifacts
make clean
```

### Manual Commands

```bash
# Install Swagger CLI tool
go install github.com/swaggo/swag/cmd/swag@latest

# Generate OpenAPI specification
$(go env GOPATH)/bin/swag init -g src/cmd/server/server.go -o docs

# Build
go build -o server src/cmd/server/server.go

# Run
./server
```

## Example Usage

1. **Create a ticket**
   ```bash
   curl -X POST http://localhost:8080/ticket \
     -H "Content-Type: application/json" \
     -d '{
       "origin": "JFK",
       "destination": "LAX", 
       "departure_date": "2024-12-25",
       "departure_time": "14:30",
       "passengers": 2
     }'
   ```

2. **Get a ticket**
   ```bash
   curl http://localhost:8080/ticket/ABC123
   ```

3. **Update a ticket**
   ```bash
   curl -X PUT http://localhost:8080/ticket/ABC123 \
     -H "Content-Type: application/json" \
     -d '{"passengers": 3}'
   ```

4. **Cancel a ticket**
   ```bash
   curl -X DELETE http://localhost:8080/ticket/ABC123
   ```

5. **List tickets**
   ```bash
   curl http://localhost:8080/tickets?limit=10
   ```

6. **Health check**
   ```bash
   curl http://localhost:8080/health
   ```

## API Documentation Features

### OpenAPI Specification
- **Complete OpenAPI 3.0 specification** generated automatically from code annotations
- **Request/Response schemas** with examples and validation rules
- **Parameter documentation** with types, constraints, and examples
- **Error response documentation** with appropriate HTTP status codes

### Swagger UI Features
- **Interactive API testing** - Test endpoints directly from the browser
- **Request/Response examples** - See sample payloads and responses
- **Schema validation** - Understand data structures and constraints
- **Authentication documentation** (when applicable)
- **Export capabilities** - Download OpenAPI spec in JSON/YAML format

### Documentation Structure
- **Organized by tags** - Endpoints grouped logically (tickets, health)
- **Detailed descriptions** - Each endpoint includes purpose and usage
- **Parameter documentation** - Path, query, and body parameters explained
- **Response codes** - All possible HTTP status codes documented
- **Model definitions** - Complete data structure documentation

## Data Formats

- **Airport Codes**: 3-letter IATA codes (e.g., JFK, LAX, ORD)
- **Dates**: YYYY-MM-DD format
- **Times**: HH:MM format (24-hour)
- **Flight Numbers**: Standard airline format (e.g., AA1234, UA567)
- **Confirmation IDs**: 6-character alphanumeric (auto-generated)

## Status Values

- `CONFIRMED`: Ticket is confirmed and active
- `PENDING`: Ticket is pending confirmation
- `CANCELLED`: Ticket has been cancelled

## Error Handling

The API returns appropriate HTTP status codes with structured error responses:
- `200`: Success
- `201`: Created
- `400`: Bad Request (validation errors)
- `404`: Not Found
- `500`: Internal Server Error

### Error Response Format
```json
{
  "error": "Error message",
  "message": "Detailed error description (optional)"
}
```

## Development

### Project Structure
```
flight-ticket-service/
├── src/
│   ├── cmd/server/          # Main application entry point
│   ├── handlers/            # HTTP request handlers
│   ├── models/              # Data models and structures
│   └── services/            # Business logic and external services
├── docs/                    # Generated OpenAPI documentation
├── Makefile                 # Development commands
├── Dockerfile               # Container configuration
└── README.md               # This file
```

### Adding New Endpoints

1. **Add handler function** with Swagger annotations:
   ```go
   // @Summary Endpoint summary
   // @Description Detailed description
   // @Tags tag-name
   // @Accept json
   // @Produce json
   // @Param param-name path string true "Parameter description"
   // @Success 200 {object} ResponseModel
   // @Router /endpoint [method]
   func HandlerFunction(w http.ResponseWriter, r *http.Request) {
       // Implementation
   }
   ```

2. **Register route** in server.go
3. **Regenerate documentation**:
   ```bash
   make swagger-gen
   ```

### Using Mage (if available)
```bash
mage build
mage run
```

## Docker Support

```bash
# Build image
docker build -t flight-ticket-service .

# Run container
docker run -p 8080:8080 \
  -e GOOGLE_CLOUD_PROJECT=your-project-id \
  -e GOOGLE_APPLICATION_CREDENTIALS=/app/credentials.json \
  -v /path/to/credentials.json:/app/credentials.json \
  flight-ticket-service
```

## Testing

### Unit Tests
```bash
make test
```

### Integration Tests
```bash
make test-coverage
```

### API Testing
Use the Swagger UI at http://localhost:8080/swagger/ for interactive testing, or use curl/Postman with the provided examples.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Update documentation (Swagger annotations)
5. Regenerate API docs: `make swagger-gen`
6. Submit a pull request

## License

MIT License
