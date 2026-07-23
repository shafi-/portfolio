package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
)

type TCPMCPClient struct {
	address string
}

func NewMCPClient(address string) MCPClient {
	return &TCPMCPClient{
		address: address,
	}
}

func (c *TCPMCPClient) Health(ctx context.Context) error {
	conn, err := net.DialTimeout("tcp", c.address, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	defer conn.Close()

	healthRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      "health",
			"arguments": map[string]interface{}{},
		},
	}

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if err := encoder.Encode(healthRequest); err != nil {
		return fmt.Errorf("failed to send health request: %w", err)
	}

	var healthResponse map[string]interface{}
	if err := decoder.Decode(&healthResponse); err != nil {
		return fmt.Errorf("failed to read health response: %w", err)
	}

	if result, ok := healthResponse["result"].(map[string]interface{}); ok {
		if status, ok := result["status"].(string); ok && status == "ok" {
			return nil
		}
	}

	return fmt.Errorf("health check failed: MCP server not healthy")
}

func (c *TCPMCPClient) ListTools(ctx context.Context) ([]string, error) {
	conn, err := net.DialTimeout("tcp", c.address, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	defer conn.Close()

	listRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if err := encoder.Encode(listRequest); err != nil {
		return nil, fmt.Errorf("failed to send list request: %w", err)
	}

	var listResponse map[string]interface{}
	if err := decoder.Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("failed to read list response: %w", err)
	}

	if result, ok := listResponse["result"].(map[string]interface{}); ok {
		if tools, ok := result["tools"].([]interface{}); ok {
			var toolNames []string
			for _, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					if name, ok := toolMap["name"].(string); ok {
						toolNames = append(toolNames, name)
					}
				}
			}
			return toolNames, nil
		}
	}

	return nil, fmt.Errorf("failed to parse tool list")
}

func (c *TCPMCPClient) RegisterTools(ctx context.Context, tools []ToolDef) error {
	conn, err := net.DialTimeout("tcp", c.address, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	defer conn.Close()

	registerRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/register",
		"params": map[string]interface{}{
			"tools": tools,
		},
	}

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if err := encoder.Encode(registerRequest); err != nil {
		return fmt.Errorf("failed to send register request: %w", err)
	}

	var registerResponse map[string]interface{}
	if err := decoder.Decode(&registerResponse); err != nil {
		return fmt.Errorf("failed to read register response: %w", err)
	}

	if _, ok := registerResponse["result"]; ok {
		return nil
	}

	return fmt.Errorf("failed to register tools")
}

type StdioMCPClient struct {
	binaryPath string
	args       []string
}

func NewStdioMCPClient(binaryPath string, args []string) MCPClient {
	return &StdioMCPClient{
		binaryPath: binaryPath,
		args:       args,
	}
}

func (c *StdioMCPClient) Health(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.binaryPath, c.args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	defer stdout.Close()

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}
	defer cmd.Process.Kill()

	healthRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      "health",
			"arguments": map[string]interface{}{},
		},
	}

	encoder := json.NewEncoder(stdin)
	decoder := json.NewDecoder(stdout)

	if err := encoder.Encode(healthRequest); err != nil {
		return fmt.Errorf("failed to send health request: %w", err)
	}

	var healthResponse map[string]interface{}
	if err := decoder.Decode(&healthResponse); err != nil {
		return fmt.Errorf("failed to read health response: %w", err)
	}

	if result, ok := healthResponse["result"].(map[string]interface{}); ok {
		if status, ok := result["status"].(string); ok && status == "ok" {
			return nil
		}
	}

	return fmt.Errorf("health check failed: MCP server not healthy")
}

func (c *StdioMCPClient) ListTools(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, c.binaryPath, c.args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	defer stdout.Close()

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}
	defer cmd.Process.Kill()

	listRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	encoder := json.NewEncoder(stdin)
	decoder := json.NewDecoder(stdout)

	if err := encoder.Encode(listRequest); err != nil {
		return nil, fmt.Errorf("failed to send list request: %w", err)
	}

	var listResponse map[string]interface{}
	if err := decoder.Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("failed to read list response: %w", err)
	}

	if result, ok := listResponse["result"].(map[string]interface{}); ok {
		if tools, ok := result["tools"].([]interface{}); ok {
			var toolNames []string
			for _, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					if name, ok := toolMap["name"].(string); ok {
						toolNames = append(toolNames, name)
					}
				}
			}
			return toolNames, nil
		}
	}

	return nil, fmt.Errorf("failed to parse tool list")
}

func (c *StdioMCPClient) RegisterTools(ctx context.Context, tools []ToolDef) error {
	cmd := exec.CommandContext(ctx, c.binaryPath, c.args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	defer stdout.Close()

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}
	defer cmd.Process.Kill()

	registerRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/register",
		"params": map[string]interface{}{
			"tools": tools,
		},
	}

	encoder := json.NewEncoder(stdin)
	decoder := json.NewDecoder(stdout)

	if err := encoder.Encode(registerRequest); err != nil {
		return fmt.Errorf("failed to send register request: %w", err)
	}

	var registerResponse map[string]interface{}
	if err := decoder.Decode(&registerResponse); err != nil {
		return fmt.Errorf("failed to read register response: %w", err)
	}

	if _, ok := registerResponse["result"]; ok {
		return nil
	}

	return fmt.Errorf("failed to register tools")
}

type IOReaderWriterMCPClient struct {
	reader io.Reader
	writer io.Writer
}

func NewIOReaderWriterMCPClient(reader io.Reader, writer io.Writer) MCPClient {
	return &IOReaderWriterMCPClient{
		reader: reader,
		writer: writer,
	}
}

func (c *IOReaderWriterMCPClient) Health(ctx context.Context) error {
	healthRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      "health",
			"arguments": map[string]interface{}{},
		},
	}

	encoder := json.NewEncoder(c.writer)
	decoder := json.NewDecoder(c.reader)

	if err := encoder.Encode(healthRequest); err != nil {
		return fmt.Errorf("failed to send health request: %w", err)
	}

	var healthResponse map[string]interface{}
	if err := decoder.Decode(&healthResponse); err != nil {
		return fmt.Errorf("failed to read health response: %w", err)
	}

	if result, ok := healthResponse["result"].(map[string]interface{}); ok {
		if status, ok := result["status"].(string); ok && status == "ok" {
			return nil
		}
	}

	return fmt.Errorf("health check failed: MCP server not healthy")
}

func (c *IOReaderWriterMCPClient) ListTools(ctx context.Context) ([]string, error) {
	listRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	encoder := json.NewEncoder(c.writer)
	decoder := json.NewDecoder(c.reader)

	if err := encoder.Encode(listRequest); err != nil {
		return nil, fmt.Errorf("failed to send list request: %w", err)
	}

	var listResponse map[string]interface{}
	if err := decoder.Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("failed to read list response: %w", err)
	}

	if result, ok := listResponse["result"].(map[string]interface{}); ok {
		if tools, ok := result["tools"].([]interface{}); ok {
			var toolNames []string
			for _, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					if name, ok := toolMap["name"].(string); ok {
						toolNames = append(toolNames, name)
					}
				}
			}
			return toolNames, nil
		}
	}

	return nil, fmt.Errorf("failed to parse tool list")
}

func (c *IOReaderWriterMCPClient) RegisterTools(ctx context.Context, tools []ToolDef) error {
	registerRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/register",
		"params": map[string]interface{}{
			"tools": tools,
		},
	}

	encoder := json.NewEncoder(c.writer)
	decoder := json.NewDecoder(c.reader)

	if err := encoder.Encode(registerRequest); err != nil {
		return fmt.Errorf("failed to send register request: %w", err)
	}

	var registerResponse map[string]interface{}
	if err := decoder.Decode(&registerResponse); err != nil {
		return fmt.Errorf("failed to read register response: %w", err)
	}

	if _, ok := registerResponse["result"]; ok {
		return nil
	}

	return fmt.Errorf("failed to register tools")
}