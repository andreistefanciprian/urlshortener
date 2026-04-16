import httpx
import os
from mcp.server.fastmcp import FastMCP

BASE_URL = os.getenv("API_GATEWAY_URL", "http://api-gateway:8080").rstrip("/")

http_client = httpx.AsyncClient()

mcp = FastMCP("urlshortener", host="0.0.0.0", port=3000)


@mcp.tool()
async def create_short_url(long_url: str, expires_in: int = 7) -> dict:
    """Create a short URL from a long URL. expires_in is days until expiry (default 7)."""
    r = await http_client.post(
        f"{BASE_URL}/create",
        json={"longUrl": long_url, "expiresIn": expires_in},
    )
    r.raise_for_status()
    return r.json()


@mcp.tool()
async def list_urls() -> list:
    """List all short URLs in the system."""
    r = await http_client.get(f"{BASE_URL}/")
    r.raise_for_status()
    return r.json()


@mcp.tool()
async def delete_short_url(short_code: str) -> dict:
    """Delete a short URL by its 7-character short code (e.g. gt0PFD8)."""
    r = await http_client.delete(f"{BASE_URL}/{short_code}")
    r.raise_for_status()
    return r.json()


if __name__ == "__main__":
    mcp.run(transport="sse")
