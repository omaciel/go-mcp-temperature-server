# Temperature MCP Server Example

This project demonstrates how to build a simple MCP (Model Context Protocol) server in Go that provides temperature data for a given location. The server exposes a `get_temperature` tool via MCP, which proxies requests to an underlying HTTP temperature service.

## Features

- Implements an MCP server using the [`mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go) library.
- Registers a tool (`get_temperature`) that accepts a `location` parameter.
- Proxies temperature requests to a local or remote HTTP service.
- Well-documented code for educational purposes.

## Project Structure

- `main.go`: Main entry point. Sets up the MCP server, registers the tool, and implements the handler logic.
- `go.mod`, `go.sum`: Go module files for dependency management.

## Usage

### Prerequisites

- Go 1.18 or newer
- An HTTP temperature service running locally on `http://localhost:8080/temperature?location=<LOCATION>` (you can use your own implementation or [go-temperature-server](https://github.com/omaciel/go-temperature-server))

### Running the MCP Server

1. Clone the repository or copy the source files to your Go workspace.
2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Build the MCP server:

   ```sh
   go build -o mcp-temperature-server main.go
   ```

4. Ensure your HTTP temperature service is running on port 8080.
5. Start the MCP server:

   ```sh
   ./mcp-temperature-server
   ```

### Example Request

You can use an MCP-compatible client or integration to call the `get_temperature` tool with a location parameter, e.g.:

```json
{
  "tool": "get_temperature",
  "params": { "location": "Chapel Hill" }
}
```

### Example Response

The response will be in plain text, containing a JSON object with the location and temperature:

```sh
Temperature for Chapel Hill: {"location":"Chapel Hill","temperature":18.25}
```

---

### Example `mcp_config.json` for Windsurf IDE

To use this project with the Windsurf IDE, create a `.codeium/windsurf/mcp_config.json` file in your home directory (or project root) with the following content:

```json
{
  "mcpServers": {
    "temperature": {
      "command": "./ABSOLUTE/PATH/TO/YOUR/REPO/go-mcp-temperature-server/go-mcp-temperature-server",
      // <-- Modify this to match your checkout location and leave the initial ./ alone
      "args": [],
      "env": {
        "WEATHER_API_KEY": "YOUR_API_KEY_HERE"
      }
    }
  }
}
```

- Replace `YOUR_API_KEY_HERE` with your actual weather API key.
- The path to the server binary should match your project structure.
- This configuration ensures the MCP server is started with the correct environment variable for authentication.

## How It Works

- The MCP server defines a tool called `get_temperature`.
- When invoked, it extracts the `location` argument, then queries the HTTP service at `http://localhost:8080/temperature?location=<LOCATION>&units=metric&appid=<YOUR_API_KEY>`.
- The result is returned as plain text (with JSON content) to the MCP client.

## Customization

- To use a different HTTP temperature service, modify the `endpoint` variable in `main.go`.
- The backend temperature service expects the API key as the `appid` query parameter (e.g., `...&appid=YOUR_API_KEY`). If you receive a 500 Internal Server Error, check the backend service logs and ensure the API key is valid and passed as a query parameter.
- To add more tools or capabilities, register additional tools and handlers using the MCP server API.

## License

MIT or as specified in your project.

---

For more information on MCP, see [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go).
