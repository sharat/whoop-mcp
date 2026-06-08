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
   - **Important:** Add the default Redirect URI, `http://127.0.0.1:8181/callback`, to the redirect URIs list. Register the corresponding URI instead if you configure another callback port.
   - Save the app configuration to generate your **Client ID** and **Client Secret**.

---

## Setup & Configuration

### 1. Environment Variables
Create `.env` from the provided example:

```bash
cp .env.example .env
```

Edit `.env` and replace the placeholder values for `WHOOP_CLIENT_ID` and
`WHOOP_CLIENT_SECRET`. `WHOOP_CALLBACK_PORT` is optional and defaults to `8181`.

The `.env` file is loaded from the process working directory. When launching the
server from another directory or through an MCP client, set these variables in
the client configuration instead. Explicit environment variables take
precedence over `.env` and saved configuration values.

If you change `WHOOP_CALLBACK_PORT`, register the matching redirect URI in the
WHOOP Developer Dashboard. For example, port `9191` requires
`http://127.0.0.1:9191/callback`.

### 2. Run OAuth Authorization
Start the temporary loopback HTTP server to authorize the server to read your WHOOP data:

```bash
go run cmd/whoop-mcp/main.go login
```

1. The terminal will print a login URL. Open it in your browser.
2. Log in with your WHOOP credentials and grant permissions.
3. Upon approval, you will be redirected to the configured callback URI (port `8181` by default), and your tokens will be saved to `~/.config/whoop-mcp/config.json`.

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

To use this server with **Claude Desktop**, add the following config to your `claude_desktop_config.json` (located at `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS).

### Option 1: Run from Source (requires Go)
```json
{
  "mcpServers": {
    "whoop": {
      "command": "go",
      "args": [
        "run",
        "/PATH/TO/whoop-mcp/cmd/whoop-mcp/main.go"
      ],
      "env": {
        "WHOOP_CLIENT_ID": "your_client_id_here",
        "WHOOP_CLIENT_SECRET": "your_client_secret_here"
      }
    }
  }
}
```

### Option 2: Use Pre-built Binary (no Go required)
Download from [Releases](https://github.com/sharat/whoop-mcp/releases) or build locally:
```bash
go build -o whoop-mcp ./cmd/whoop-mcp
```
Then configure:
```json
{
  "mcpServers": {
    "whoop": {
      "command": "/PATH/TO/whoop-mcp/whoop-mcp",
      "env": {
        "WHOOP_CLIENT_ID": "your_client_id_here",
        "WHOOP_CLIENT_SECRET": "your_client_secret_here"
      }
    }
  }
}
```

### Option 3: Installed via `go install`
```bash
go install github.com/sharat/whoop-mcp@latest
```
Then configure (binary at `~/go/bin/whoop-mcp`):
```json
{
  "mcpServers": {
    "whoop": {
      "command": "~/go/bin/whoop-mcp",
      "env": {
        "WHOOP_CLIENT_ID": "your_client_id_here",
        "WHOOP_CLIENT_SECRET": "your_client_secret_here"
      }
    }
  }
}
```

*(Note: Replace `/PATH/TO/whoop-mcp` with the absolute path to your cloned repository or downloaded binary. Find your Go path with `which go` — use just `go` if it's in your PATH, or the full path like `/opt/homebrew/bin/go`.)*
