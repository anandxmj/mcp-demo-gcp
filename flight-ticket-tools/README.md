# Flight Ticket Tools

A Model Context Protocol (MCP) server that provides tools for managing flight tickets through a REST API service.

## Deployment

This application can be deployed to Google Cloud Run or run locally using different MCP transport modes.

### Prerequisites

- [Go 1.24+](https://golang.org/dl/) (for Mage build tool)
- [Mage](https://magefile.org/) build tool
- [uv](https://docs.astral.sh/uv/) Python package manager
- [Docker](https://www.docker.com/)
- [Google Cloud CLI](https://cloud.google.com/sdk/docs/install)

### Setup

```bash
# Install dependencies and setup environment
mage setup

# Or manually:
# Install uv if not present
curl -LsSf https://astral.sh/uv/install.sh | sh

# Sync Python dependencies
uv sync

# Configure gcloud
gcloud auth login
gcloud config set project a3rlabs-sandbox
```

### Local Development

```bash
# Run locally in stdio mode (for MCP clients)
mage dev:local

# Run locally in HTTP mode (for testing Cloud Run behavior)
mage dev:http

# Run tests
mage dev:test
```

### Docker Operations

```bash
# Build Docker image
mage docker:build

# Run Docker image locally
mage docker:run

# Push to Artifact Registry
mage docker:push
```

### Cloud Run Deployment

```bash
# Deploy to Cloud Run (builds, pushes, and deploys)
mage cloudrun:deploy

# Update existing service
mage cloudrun:update

# Check service status
mage cloudrun:status

# View logs
mage cloudrun:logs

# Delete service
mage cloudrun:delete
```

### Configuration

The application automatically detects its environment:

- **Local mode** (`ENVIRONMENT=local`): Uses stdio transport for MCP communication
- **Cloud Run mode** (`ENVIRONMENT=cloudrun`): Uses HTTP transport with health checks

Environment variables:
- `ENVIRONMENT`: Set to "cloudrun" for Cloud Run deployment, "local" for local development
- `PORT`: Port number for HTTP server (default: 8080)

### Service Account

The Cloud Run deployment uses the `flight-ticket-service@a3rlabs-sandbox.iam.gserviceaccount.com` service account for secure access to Google Cloud resources.

### Health Checks

The application provides a `/health` endpoint for Cloud Run health checks that returns:
```json
{
  "status": "healthy",
  "service": "flight-ticket-tools",
  "timestamp": "2025-07-15T05:00:00.000000",
  "environment": "cloudrun"
}
```

## Available Tools

This MCP server provides the following tools for flight ticket management:

### 1. `health_check()`
Check the health status of the Flight Ticket Service.

**Returns:** Dict containing service health information including status, service name, version, and timestamp.

### 2. `create_flight_ticket(origin, destination, departure_date, departure_time, passengers, flight_number=None)`
Create a new flight ticket with the provided details.

**Parameters:**
- `origin` (str): Origin airport code (e.g., "JFK")
- `destination` (str): Destination airport code (e.g., "LAX")
- `departure_date` (str): Departure date in YYYY-MM-DD format (e.g., "2024-12-25")
- `departure_time` (str): Departure time in HH:MM format (e.g., "14:30")
- `passengers` (int): Number of passengers (minimum 1)
- `flight_number` (str, optional): Flight number (e.g., "AA1234")

**Returns:** Dict containing the created flight ticket information or error details.

### 3. `get_flight_ticket(confirmation_id)`
Retrieve a flight ticket using its confirmation ID.

**Parameters:**
- `confirmation_id` (str): Ticket confirmation ID (e.g., "ABC123")

**Returns:** Dict containing the flight ticket information or error details.

### 4. `update_flight_ticket(confirmation_id, origin=None, destination=None, departure_date=None, departure_time=None, passengers=None, flight_number=None, status=None)`
Update an existing flight ticket with new information.

**Parameters:**
- `confirmation_id` (str): Ticket confirmation ID (e.g., "ABC123")
- `origin` (str, optional): New origin airport code (e.g., "JFK")
- `destination` (str, optional): New destination airport code (e.g., "LAX")
- `departure_date` (str, optional): New departure date in YYYY-MM-DD format (e.g., "2024-12-25")
- `departure_time` (str, optional): New departure time in HH:MM format (e.g., "14:30")
- `passengers` (int, optional): New number of passengers (minimum 1)
- `flight_number` (str, optional): New flight number (e.g., "AA1234")
- `status` (str, optional): New status ("CONFIRMED", "CANCELLED", or "PENDING")

**Returns:** Dict containing the updated flight ticket information or error details.

### 5. `cancel_flight_ticket(confirmation_id)`
Cancel (soft delete) a flight ticket by setting its status to CANCELLED.

**Parameters:**
- `confirmation_id` (str): Ticket confirmation ID (e.g., "ABC123")

**Returns:** Dict containing success message and confirmation ID or error details.

### 6. `list_flight_tickets(limit=50)`
Retrieve a list of all flight tickets with optional pagination.

**Parameters:**
- `limit` (int, optional): Maximum number of tickets to return (default: 50)

**Returns:** Dict containing list of tickets with count or error details.

## API Service

The tools connect to a Flight Ticket Service API hosted at:
`https://flight-ticket-service-858333166396.us-east1.run.app`

This is a Go-based REST API service for managing flight tickets using Google Cloud Firestore.

## Usage

This MCP server can be used with any MCP-compatible client. The tools are automatically exposed and can be called to manage flight tickets programmatically.

## Dependencies

- `httpx>=0.28.1` - For making HTTP requests
- `mcp[cli]>=1.11.0` - Model Context Protocol framework

## Example Usage

```python
# Check service health
health_status = health_check()

# Create a new ticket
ticket = create_flight_ticket(
    origin="SFO",
    destination="NYC",
    departure_date="2025-08-15",
    departure_time="10:30",
    passengers=1,
    flight_number="UA456"
)

# Get ticket by confirmation ID
ticket_details = get_flight_ticket("ABC123")

# List all tickets
all_tickets = list_flight_tickets(limit=10)

# Update a ticket
updated_ticket = update_flight_ticket(
    confirmation_id="ABC123",
    passengers=2,
    status="CONFIRMED"
)

# Cancel a ticket
cancellation_result = cancel_flight_ticket("ABC123")
```
