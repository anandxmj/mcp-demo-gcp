#!/usr/bin/env python3
"""
Test script for the HTTP mode of the MCP server.
This simulates how the server would behave in Cloud Run.
"""

import asyncio
import httpx
import json
from datetime import datetime

async def test_health_endpoint():
    """Test the health endpoint."""
    async with httpx.AsyncClient() as client:
        try:
            response = await client.get("http://localhost:8080/health")
            print(f"Health check status: {response.status_code}")
            print(f"Response: {response.json()}")
            return response.status_code == 200
        except Exception as e:
            print(f"Health check failed: {e}")
            return False

async def test_mcp_tools():
    """Test MCP tools via HTTP."""
    async with httpx.AsyncClient() as client:
        try:
            # Test health_check tool
            payload = {
                "method": "tools/call",
                "params": {
                    "name": "health_check",
                    "arguments": {}
                }
            }
            
            response = await client.post(
                "http://localhost:8080/mcp/v1/tools/call",
                json=payload,
                headers={"Content-Type": "application/json"}
            )
            
            print(f"MCP tool call status: {response.status_code}")
            if response.status_code == 200:
                print(f"Response: {response.json()}")
            else:
                print(f"Error: {response.text}")
                
        except Exception as e:
            print(f"MCP tool test failed: {e}")

async def main():
    """Main test function."""
    print("Testing Flight Ticket Tools MCP Server in HTTP mode")
    print("=" * 50)
    
    print("\n1. Testing health endpoint...")
    health_ok = await test_health_endpoint()
    
    if health_ok:
        print("\n2. Testing MCP tools...")
        await test_mcp_tools()
    else:
        print("Health check failed, skipping MCP tool tests")
    
    print("\nTest completed!")

if __name__ == "__main__":
    asyncio.run(main())
