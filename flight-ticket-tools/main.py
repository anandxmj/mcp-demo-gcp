from mcp.server.fastmcp import FastMCP
from mcp.server.session import ServerSession
from mcp.server.stdio import stdio_server
from mcp.types import ServerCapabilities, Tool
from os import getenv
import sys
import signal
import httpx
import json
import asyncio
from typing import Optional, Dict, Any, List
from datetime import datetime
from starlette.applications import Starlette
from starlette.responses import JSONResponse, Response
from starlette.routing import Route
from starlette.middleware.cors import CORSMiddleware
from starlette.requests import Request
import uvicorn

# Configuration
PORT = int(getenv("PORT", "8080"))
ENVIRONMENT = getenv("ENVIRONMENT", "local")  # "local" or "cloudrun"

# Base URL for the Flight Ticket Service API
BASE_URL = "[HTTPS URL for flight-ticket-service]"

# Initialize MCP server
mcp = FastMCP("FlightTicketTools")

# Health endpoint function for Cloud Run (not using MCP decorator)
async def health_endpoint():
    """Health check endpoint for Cloud Run."""
    return {
        "status": "healthy",
        "service": "flight-ticket-tools",
        "timestamp": datetime.utcnow().isoformat(),
        "environment": ENVIRONMENT
    }

@mcp.tool()
def health_check() -> Dict[str, Any]:
    """
    Check the health status of the Flight Ticket Service.
    
    Returns:
        Dict containing service health information including status, service name, version, and timestamp.
    """
    try:
        with httpx.Client() as client:
            response = client.get(f"{BASE_URL}/health")
            response.raise_for_status()
            return response.json()
    except httpx.RequestError as e:
        return {"error": f"Failed to check health: {str(e)}"}
    except httpx.HTTPStatusError as e:
        try:
            error_data = e.response.json()
            return {"error": error_data}
        except:
            return {"error": f"HTTP {e.response.status_code}: {e.response.text}"}

@mcp.tool()
def create_flight_ticket(
    origin: str,
    destination: str,
    departure_date: str,
    departure_time: str,
    passengers: int,
    flight_number: Optional[str] = None
) -> Dict[str, Any]:
    """
    Create a new flight ticket with the provided details.
    
    Args:
        origin: Origin airport code (e.g., "JFK")
        destination: Destination airport code (e.g., "LAX")
        departure_date: Departure date in YYYY-MM-DD format (e.g., "2024-12-25")
        departure_time: Departure time in HH:MM format (e.g., "14:30")
        passengers: Number of passengers (minimum 1)
        flight_number: Flight number (e.g., "AA1234") - optional
    
    Returns:
        Dict containing the created flight ticket information or error details.
    """
    ticket_data = {
        "origin": origin,
        "destination": destination,
        "departure_date": departure_date,
        "departure_time": departure_time,
        "passengers": passengers
    }
    
    if flight_number:
        ticket_data["flight_number"] = flight_number
    
    try:
        with httpx.Client() as client:
            response = client.post(f"{BASE_URL}/ticket", json=ticket_data)
            response.raise_for_status()
            return response.json()
    except httpx.RequestError as e:
        return {"error": f"Failed to create ticket: {str(e)}"}
    except httpx.HTTPStatusError as e:
        try:
            error_data = e.response.json()
            return {"error": error_data}
        except:
            return {"error": f"HTTP {e.response.status_code}: {e.response.text}"}

@mcp.tool()
def get_flight_ticket(confirmation_id: str) -> Dict[str, Any]:
    """
    Retrieve a flight ticket using its confirmation ID.
    
    Args:
        confirmation_id: Ticket confirmation ID (e.g., "ABC123")
    
    Returns:
        Dict containing the flight ticket information or error details.
    """
    try:
        with httpx.Client() as client:
            response = client.get(f"{BASE_URL}/ticket/{confirmation_id}")
            response.raise_for_status()
            return response.json()
    except httpx.RequestError as e:
        return {"error": f"Failed to get ticket: {str(e)}"}
    except httpx.HTTPStatusError as e:
        try:
            error_data = e.response.json()
            return {"error": error_data}
        except:
            return {"error": f"HTTP {e.response.status_code}: {e.response.text}"}

@mcp.tool()
def update_flight_ticket(
    confirmation_id: str,
    origin: Optional[str] = None,
    destination: Optional[str] = None,
    departure_date: Optional[str] = None,
    departure_time: Optional[str] = None,
    passengers: Optional[int] = None,
    flight_number: Optional[str] = None,
    status: Optional[str] = None
) -> Dict[str, Any]:
    """
    Update an existing flight ticket with new information.
    
    Args:
        confirmation_id: Ticket confirmation ID (e.g., "ABC123")
        origin: New origin airport code (e.g., "JFK") - optional
        destination: New destination airport code (e.g., "LAX") - optional
        departure_date: New departure date in YYYY-MM-DD format (e.g., "2024-12-25") - optional
        departure_time: New departure time in HH:MM format (e.g., "14:30") - optional
        passengers: New number of passengers (minimum 1) - optional
        flight_number: New flight number (e.g., "AA1234") - optional
        status: New status ("CONFIRMED", "CANCELLED", or "PENDING") - optional
    
    Returns:
        Dict containing the updated flight ticket information or error details.
    """
    update_data = {}
    
    if origin is not None:
        update_data["origin"] = origin
    if destination is not None:
        update_data["destination"] = destination
    if departure_date is not None:
        update_data["departure_date"] = departure_date
    if departure_time is not None:
        update_data["departure_time"] = departure_time
    if passengers is not None:
        update_data["passengers"] = passengers
    if flight_number is not None:
        update_data["flight_number"] = flight_number
    if status is not None:
        update_data["status"] = status
    
    try:
        with httpx.Client() as client:
            response = client.put(f"{BASE_URL}/ticket/{confirmation_id}", json=update_data)
            response.raise_for_status()
            return response.json()
    except httpx.RequestError as e:
        return {"error": f"Failed to update ticket: {str(e)}"}
    except httpx.HTTPStatusError as e:
        try:
            error_data = e.response.json()
            return {"error": error_data}
        except:
            return {"error": f"HTTP {e.response.status_code}: {e.response.text}"}

@mcp.tool()
def cancel_flight_ticket(confirmation_id: str) -> Dict[str, Any]:
    """
    Cancel (soft delete) a flight ticket by setting its status to CANCELLED.
    
    Args:
        confirmation_id: Ticket confirmation ID (e.g., "ABC123")
    
    Returns:
        Dict containing success message and confirmation ID or error details.
    """
    try:
        with httpx.Client() as client:
            response = client.delete(f"{BASE_URL}/ticket/{confirmation_id}")
            response.raise_for_status()
            return response.json()
    except httpx.RequestError as e:
        return {"error": f"Failed to cancel ticket: {str(e)}"}
    except httpx.HTTPStatusError as e:
        try:
            error_data = e.response.json()
            return {"error": error_data}
        except:
            return {"error": f"HTTP {e.response.status_code}: {e.response.text}"}

@mcp.tool()
def list_flight_tickets(limit: Optional[int] = 50) -> Dict[str, Any]:
    """
    Retrieve a list of all flight tickets with optional pagination.
    
    Args:
        limit: Maximum number of tickets to return (default: 50)
    
    Returns:
        Dict containing list of tickets with count or error details.
    """
    params = {}
    if limit is not None:
        params["limit"] = limit
    
    try:
        with httpx.Client() as client:
            response = client.get(f"{BASE_URL}/tickets", params=params)
            response.raise_for_status()
            return response.json()
    except httpx.RequestError as e:
        return {"error": f"Failed to list tickets: {str(e)}"}
    except httpx.HTTPStatusError as e:
        try:
            error_data = e.response.json()
            return {"error": error_data}
        except:
            return {"error": f"HTTP {e.response.status_code}: {e.response.text}"}

# Streamable HTTP Transport Implementation
async def handle_mcp_message(request: Request):
    """Handle MCP messages over streamable HTTP transport."""
    try:
        # Read the request body
        body = await request.body()
        if not body:
            return Response(
                content=json.dumps({
                    "jsonrpc": "2.0",
                    "error": {"code": -32600, "message": "Invalid Request"}
                }),
                media_type="application/json",
                status_code=400
            )
        
        # Parse JSON-RPC message
        try:
            message = json.loads(body.decode())
        except json.JSONDecodeError:
            return Response(
                content=json.dumps({
                    "jsonrpc": "2.0",
                    "error": {"code": -32700, "message": "Parse error"}
                }),
                media_type="application/json",
                status_code=400
            )
        
        # Handle MCP protocol messages manually since FastMCP doesn't expose server directly
        if message.get("method") == "initialize":
            response = {
                "jsonrpc": "2.0",
                "id": message.get("id"),
                "result": {
                    "protocolVersion": "2024-11-05",
                    "capabilities": {
                        "tools": {"listChanged": False},
                        "resources": {"subscribe": False, "listChanged": False},
                        "prompts": {"listChanged": False},
                        "experimental": {}
                    },
                    "serverInfo": {
                        "name": "FlightTicketTools",
                        "version": "1.0.0"
                    }
                }
            }
        elif message.get("method") == "tools/list":
            # List available tools
            tools = [
                {
                    "name": "health_check",
                    "description": "Check the health status of the Flight Ticket Service",
                    "inputSchema": {
                        "type": "object",
                        "properties": {},
                        "required": []
                    }
                },
                {
                    "name": "create_flight_ticket",
                    "description": "Create a new flight ticket with the provided details",
                    "inputSchema": {
                        "type": "object",
                        "properties": {
                            "origin": {"type": "string", "description": "Origin airport code (e.g., 'JFK')"},
                            "destination": {"type": "string", "description": "Destination airport code (e.g., 'LAX')"},
                            "departure_date": {"type": "string", "description": "Departure date in YYYY-MM-DD format"},
                            "departure_time": {"type": "string", "description": "Departure time in HH:MM format"},
                            "passengers": {"type": "integer", "description": "Number of passengers (minimum 1)"},
                            "flight_number": {"type": "string", "description": "Flight number (optional)"}
                        },
                        "required": ["origin", "destination", "departure_date", "departure_time", "passengers"]
                    }
                },
                {
                    "name": "get_flight_ticket",
                    "description": "Retrieve a flight ticket using its confirmation ID",
                    "inputSchema": {
                        "type": "object",
                        "properties": {
                            "confirmation_id": {"type": "string", "description": "Ticket confirmation ID"}
                        },
                        "required": ["confirmation_id"]
                    }
                },
                {
                    "name": "update_flight_ticket",
                    "description": "Update an existing flight ticket with new information",
                    "inputSchema": {
                        "type": "object",
                        "properties": {
                            "confirmation_id": {"type": "string", "description": "Ticket confirmation ID"},
                            "origin": {"type": "string", "description": "New origin airport code (optional)"},
                            "destination": {"type": "string", "description": "New destination airport code (optional)"},
                            "departure_date": {"type": "string", "description": "New departure date (optional)"},
                            "departure_time": {"type": "string", "description": "New departure time (optional)"},
                            "passengers": {"type": "integer", "description": "New number of passengers (optional)"},
                            "flight_number": {"type": "string", "description": "New flight number (optional)"},
                            "status": {"type": "string", "description": "New status (optional)"}
                        },
                        "required": ["confirmation_id"]
                    }
                },
                {
                    "name": "cancel_flight_ticket",
                    "description": "Cancel a flight ticket by setting its status to CANCELLED",
                    "inputSchema": {
                        "type": "object",
                        "properties": {
                            "confirmation_id": {"type": "string", "description": "Ticket confirmation ID"}
                        },
                        "required": ["confirmation_id"]
                    }
                },
                {
                    "name": "list_flight_tickets",
                    "description": "Retrieve a list of all flight tickets with optional pagination",
                    "inputSchema": {
                        "type": "object",
                        "properties": {
                            "limit": {"type": "integer", "description": "Maximum number of tickets to return (default: 50)"}
                        },
                        "required": []
                    }
                }
            ]
            response = {
                "jsonrpc": "2.0",
                "id": message.get("id"),
                "result": {"tools": tools}
            }
        elif message.get("method") == "tools/call":
            # Handle tool calls
            tool_name = message.get("params", {}).get("name")
            arguments = message.get("params", {}).get("arguments", {})
            
            try:
                if tool_name == "health_check":
                    result = health_check()
                elif tool_name == "create_flight_ticket":
                    result = create_flight_ticket(**arguments)
                elif tool_name == "get_flight_ticket":
                    result = get_flight_ticket(**arguments)
                elif tool_name == "update_flight_ticket":
                    result = update_flight_ticket(**arguments)
                elif tool_name == "cancel_flight_ticket":
                    result = cancel_flight_ticket(**arguments)
                elif tool_name == "list_flight_tickets":
                    result = list_flight_tickets(**arguments)
                else:
                    result = {"error": f"Unknown tool: {tool_name}"}
                
                response = {
                    "jsonrpc": "2.0",
                    "id": message.get("id"),
                    "result": {
                        "content": [
                            {
                                "type": "text",
                                "text": json.dumps(result, indent=2)
                            }
                        ]
                    }
                }
            except Exception as e:
                response = {
                    "jsonrpc": "2.0",
                    "id": message.get("id"),
                    "error": {
                        "code": -32603,
                        "message": f"Tool execution error: {str(e)}"
                    }
                }
        else:
            response = {
                "jsonrpc": "2.0",
                "id": message.get("id"),
                "error": {
                    "code": -32601,
                    "message": f"Method not found: {message.get('method')}"
                }
            }
        
        return Response(
            content=json.dumps(response),
            media_type="application/json",
            headers={
                "Access-Control-Allow-Origin": "*",
                "Access-Control-Allow-Methods": "POST, OPTIONS",
                "Access-Control-Allow-Headers": "Content-Type, Authorization"
            }
        )
        
    except Exception as e:
        error_response = {
            "jsonrpc": "2.0",
            "id": message.get("id") if 'message' in locals() else None,
            "error": {
                "code": -32603,
                "message": f"Internal error: {str(e)}"
            }
        }
        return Response(
            content=json.dumps(error_response),
            media_type="application/json",
            status_code=500
        )

async def handle_options(request: Request):
    """Handle CORS preflight requests."""
    return Response(
        content="",
        headers={
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "POST, OPTIONS",
            "Access-Control-Allow-Headers": "Content-Type, Authorization",
            "Access-Control-Max-Age": "86400"
        }
    )

async def handle_health_check(request: Request):
    """Handle health check requests."""
    return JSONResponse(await health_endpoint())

async def main():
    """Main entry point that handles both local and Cloud Run environments."""
    
    def signal_handler(sig, frame):
        print("Shutting down gracefully...")
        sys.exit(0)
    
    # Set up signal handlers
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    if ENVIRONMENT == "cloudrun":
        # Cloud Run mode: HTTP server with streamable MCP transport
        print(f"Starting remote MCP server with streamable-http transport on port {PORT}")
        
        # Create Starlette app with MCP endpoints
        app = Starlette(routes=[
            Route('/health', handle_health_check, methods=['GET']),
            Route('/message', handle_mcp_message, methods=['POST']),
            Route('/message', handle_options, methods=['OPTIONS']),
        ])
        
        # Add CORS middleware
        app.add_middleware(
            CORSMiddleware,
            allow_origins=["*"],
            allow_credentials=True,
            allow_methods=["*"],
            allow_headers=["*"],
        )
        
        # Run the server
        config = uvicorn.Config(app, host="0.0.0.0", port=PORT, log_level="info")
        server = uvicorn.Server(config)
        await server.serve()
        
    else:
        # Local mode: use stdio transport
        print("Starting MCP server in stdio mode")
        try:
            # Handle broken pipe gracefully for stdio mode
            signal.signal(signal.SIGPIPE, signal.SIG_DFL)
            await mcp.run_stdio_async()
        except BrokenPipeError:
            # Handle broken pipe gracefully
            pass
        except KeyboardInterrupt:
            pass
        except Exception as e:
            print(f"Error running stdio server: {e}")
            sys.exit(1)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("Interrupted by user")
    except Exception as e:
        print(f"Unexpected error: {e}")
        sys.exit(1)
