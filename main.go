// main.go
// Example: Creating a Temperature MCP Server
//
// This file demonstrates how to build a simple MCP (Model Context Protocol) server in Go.
// The server exposes a tool to fetch the temperature for a given location. It uses the
// mark3labs/mcp-go library to define the server and tool, and proxies requests to an
// underlying HTTP temperature service.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func init() {
	logDir := filepath.Join(os.Getenv("HOME"), "Library", "Logs", "mcp-temperature-server")
	logFile := filepath.Join(logDir, "server.log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "[init] Failed to create log directory: %v\n", err)
		os.Exit(1)
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[init] Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	log.SetOutput(f)
}

func main() {
	// Step 1: Create a new MCP server instance.
	// The server will be named "Temperature Service 🌡️" and versioned as 1.0.0.
	// The WithToolCapabilities(false) disables auto-discovery of tools (explicit registration only).
	s := server.NewMCPServer(
		"Temperature Service 🌡️",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Step 2: Define the "get_temperature" tool.
	// This tool takes a single required string parameter: "location".
	// The tool's description and parameter details are provided for discoverability and documentation.
	tool := mcp.NewTool("get_temperature",
		mcp.WithDescription("Get the temperature for a given location"),
		mcp.WithString("location",
			mcp.Required(),
			mcp.Description("Name of the location to get the temperature for"),
		),
	)

	// Step 3: Register the tool and its handler with the MCP server.
	// The handler function (temperatureHandler) will be called whenever the tool is invoked.
	s.AddTool(tool, temperatureHandler)

	// Step 4: Start the MCP server using stdio (standard input/output).
	// This allows the server to communicate with clients via pipes or process integration.
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// temperatureHandler handles incoming requests to the "get_temperature" tool.
// It expects a "location" parameter and an optional "unit" parameter (defaults to "metric"), queries the underlying HTTP service, and returns the result.
func temperatureHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Debug: Log received arguments
	log.Printf("[temperatureHandler] Received Params: %+v", request.Params.Arguments)

	// Extract the "location" argument from the request parameters.
	location, ok := request.Params.Arguments["location"].(string)
	if !ok || location == "" {
		log.Println("[temperatureHandler] ERROR: location must be a non-empty string")
		return nil, errors.New("location must be a non-empty string")
	}

	// Extract the optional "unit" argument, normalize to 'metric' or 'imperial', default to 'metric'.
	unit, ok := request.Params.Arguments["unit"].(string)
	if !ok || unit == "" {
		unit = "metric"
	} else {
		switch u := strings.ToLower(unit); u {
		case "celsius", "c", "metric":
			unit = "metric"
		case "fahrenheit", "f", "imperial":
			unit = "imperial"
		default:
			unit = "metric"
		}
	}

	// Set the API key for authentication in the header
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Println("[temperatureHandler] WARNING: WEATHER_API_KEY is not set!")
	}

	// Step 1: Prepare the request URL for the HTTP temperature service.
	endpoint := "http://localhost:8080/temperature"
	reqUrl := fmt.Sprintf("%s?location=%s&units=%s&appid=%s", endpoint, url.QueryEscape(location), url.QueryEscape(unit), apiKey)
	log.Printf("[temperatureHandler] Requesting URL: %s", reqUrl)

	// Step 2: Make an HTTP GET request to the temperature service.
	resp, err := http.Get(reqUrl)
	if err != nil {
		log.Printf("[temperatureHandler] ERROR: failed to query temperature service: %v", err)
		return nil, fmt.Errorf("failed to query temperature service: %w", err)
	}
	log.Printf("[temperatureHandler] HTTP response status: %s", resp.Status)
	defer resp.Body.Close()

	// Step 3: Check for a successful response.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("temperature service returned status: %s", resp.Status)
	}

	// Step 4: Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Step 5: Return the temperature result as plain text.
	return mcp.NewToolResultText(fmt.Sprintf("Temperature for %s: %s", location, string(body))), nil
}
