# WHOOP Model Context Protocol (MCP) Go Server

A Model Context Protocol (MCP) server written in Go that integrates with **v2** of the WHOOP API. It exposes metrics for physiological cycles, sleep, recovery, workouts, profile, and body measurements to LLM applications (like Claude Desktop or Cursor).

## Features

- **OAuth 2.0 with Automatic Token Refresh:** Integrates a temporary local HTTP server to handle authorization callbacks and save rotating tokens to `~/.config/whoop-mcp/config.json`. Automatically handles background access token refresh on API calls.
- **Official MCP SDK:** Built on top of the official `modelcontextprotocol/go-sdk`.
- **Rich v2 API Data Support:** Exposes full WHOOP v2 health and activity telemetry.

---

## Prerequisites

1. **Go:** Version 1.25+ installed on your machine.
2. **WHOOP Developer Account & App Registration:**
   - Log in to the [WHOOP Developer Dashboard](https://developer-dashboard.whoop.com).
   - Click **Create App** to register a new integration.
   - Enter your App Details (Name, Description).
   - **Important:** Add the Redirect URI: `http://127.0.0.1:8181/callback` to the redirect URIs list. This port and path are required by the local OAuth loopback server to securely receive authorization tokens.
   - Save the app configuration to generate your **Client ID** and **Client Secret**.

---

## Setup & Configuration

### 1. Environment Variables
Create a file named `.env` in the root of the project (if not already present) and populate your WHOOP client credentials:

```env
WHOOP_CLIENT_ID=your_client_id_here
WHOOP_CLIENT_SECRET=your_client_secret_here
```

### 2. Run OAuth Authorization
Start the temporary loopback HTTP server to authorize the server to read your WHOOP data:

```bash
go run cmd/whoop-mcp/main.go login
```

1. The terminal will print a login URL. Open it in your browser.
2. Log in with your WHOOP credentials and grant permissions.
3. Upon approval, you will be redirected to `http://127.0.0.1:8181/callback`, and your tokens will be saved to `~/.config/whoop-mcp/config.json`.

---

## Usage

### Run local Stdio Server
Once authorized, you can run the server in standard I/O (stdio) mode to interact with MCP-enabled clients:

```bash
go run cmd/whoop-mcp/main.go
```

### Exposed Tools

The server registers the following tools:

- `get_profile`: Retrieve basic profile information (e.g. name, user ID, email).
- `get_body_measurements`: Retrieve body measurements (e.g. height, weight, max HR).
- `get_cycles`: Retrieve daily physiological cycles (strain, sleep performance, recovery).
  - *Parameters:* `start` (string, optional), `end` (string, optional), `limit` (int, optional), `next_token` (string, optional)
- `get_sleeps`: Retrieve sleep activity records (sleep states, disturbances, efficiency).
  - *Parameters:* `start` (string, optional), `end` (string, optional), `limit` (int, optional), `next_token` (string, optional)
- `get_recoveries`: Retrieve daily recovery metrics (HRV, resting HR, recovery score).
  - *Parameters:* `start` (string, optional), `end` (string, optional), `limit` (int, optional), `next_token` (string, optional)
- `get_workouts`: Retrieve workout activities (strain, duration, HR, calories).
  - *Parameters:* `start` (string, optional), `end` (string, optional), `limit` (int, optional), `next_token` (string, optional)

---

## Client Configuration (e.g., Claude Desktop)

To use this server with **Claude Desktop**, add the following config to your `claude_desktop_config.json` (located at `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "whoop": {
      "command": "/opt/homebrew/bin/go",
      "args": [
        "run",
        "/Users/sarat/github/whoop-mcp/cmd/whoop-mcp/main.go"
      ],
      "env": {
        "WHOOP_CLIENT_ID": "your_client_id_here",
        "WHOOP_CLIENT_SECRET": "your_client_secret_here"
      }
    }
  }
}
```

*(Note: Change the `command` path to match your Go installation path, which you can find by running `which go` in your terminal.)*
