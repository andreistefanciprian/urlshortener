# urlshortener MCP Server

An [MCP](https://modelcontextprotocol.io) server that exposes URL shortener operations to AI assistants (e.g. VSCode). It proxies requests to the `api-gateway` service.

## Tools

| Tool | Description |
|------|-------------|
| `create_short_url` | Create a short URL from a long URL. Optional `expires_in` param (days, default 7). |
| `list_urls` | List all short URLs. |
| `delete_short_url` | Delete a short URL by its 7-character short code. |

## Play with MCP server on local machine and code editor

```bash
docker compose up --build
```

## MCP Client configuration

A ready-to-use config is already provided at [`.vscode/mcp.json`](../.vscode/mcp.json) in the repo root. VSCode will pick it up automatically.

```json
{
  "servers": {
    "urlshortener": {
      "type": "http",
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

Create new VSCode session and ask it to create a short URL for you based on long URL ...