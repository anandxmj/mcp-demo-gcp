#!/usr/bin/env python3

import sys
import json
import httpx
import asyncio
from typing import Dict, Any

# Configuration
REMOTE_MCP_URL = "HTTPS URL for flight-ticket-tools MCP Server"

async def forward_message(message: Dict[str, Any]) -> Dict[str, Any]:
    """Forward MCP message to remote server and return response."""
    try:
        async with httpx.AsyncClient() as client:
            response = await client.post(
                REMOTE_MCP_URL,
                json=message,
                headers={"Content-Type": "application/json"},
                timeout=30.0
            )
            response.raise_for_status()
            return response.json()
    except Exception as e:
        return {
            "jsonrpc": "2.0",
            "id": message.get("id"),
            "error": {
                "code": -32603,
                "message": f"Proxy error: {str(e)}"
            }
        }

async def main():
    """Main loop to handle stdio communication."""
    while True:
        try:
            # Read line from stdin
            line = sys.stdin.readline()
            if not line:
                break
                
            line = line.strip()
            if not line:
                continue
                
            # Parse JSON-RPC message
            message = json.loads(line)
            
            # Forward to remote MCP server
            response = await forward_message(message)
            
            # Send response to stdout
            print(json.dumps(response), flush=True)
            
        except json.JSONDecodeError as e:
            error_response = {
                "jsonrpc": "2.0",
                "id": None,
                "error": {
                    "code": -32700,
                    "message": f"Parse error: {str(e)}"
                }
            }
            print(json.dumps(error_response), flush=True)
        except Exception as e:
            error_response = {
                "jsonrpc": "2.0",
                "id": None,
                "error": {
                    "code": -32603,
                    "message": f"Internal error: {str(e)}"
                }
            }
            print(json.dumps(error_response), flush=True)

if __name__ == "__main__":
    asyncio.run(main())
