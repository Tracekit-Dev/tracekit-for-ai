package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const protocolVersion = "2024-11-05"

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type callToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

func runMCP() int {
	server := newMCPServer()
	reader := bufio.NewReader(os.Stdin)
	for {
		payload, err := readMessage(reader)
		if err != nil {
			if err == io.EOF {
				return 0
			}
			writeResponse(rpcResponse{JSONRPC: "2.0", Error: &rpcError{Code: -32700, Message: err.Error()}})
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			writeResponse(rpcResponse{JSONRPC: "2.0", Error: &rpcError{Code: -32700, Message: fmt.Sprintf("invalid JSON: %v", err)}})
			continue
		}
		resp := server.handle(req)
		if len(req.ID) == 0 {
			continue
		}
		writeResponse(resp)
	}
}

type mcpServer struct{ tools map[string]tool }

func newMCPServer() *mcpServer {
	tools := []tool{
		{Name: "tracekit_status", Description: "Get the authenticated TraceKit integration status for the current account.", InputSchema: schema("object", map[string]interface{}{}, []string{})},
		{Name: "tracekit_dashboard", Description: "Fetch TraceKit dashboard metrics and active alerts.", InputSchema: schema("object", map[string]interface{}{"window": enumStringProp("Dashboard window", []string{"1h", "6h", "24h"})}, []string{})},
		{Name: "tracekit_services", Description: "List monitored services with request, error-rate, and latency health metrics.", InputSchema: schema("object", map[string]interface{}{}, []string{})},
		{Name: "tracekit_service_detail", Description: "Get detailed metrics, top errors, and hot operations for a single service.", InputSchema: schema("object", map[string]interface{}{"service_name": stringProp("TraceKit service name")}, []string{"service_name"})},
		{Name: "tracekit_traces", Description: "Search recent traces with optional service, duration, and error filters.", InputSchema: schema("object", map[string]interface{}{"service": stringProp("Optional service filter"), "has_error": boolProp("Only return traces with errors"), "min_duration_ms": intProp("Minimum trace duration in milliseconds"), "time_window": enumStringProp("Time window", []string{"1h", "6h", "24h", "all"}), "limit": intProp("Max traces to return"), "offset": intProp("Pagination offset")}, []string{})},
		{Name: "tracekit_alert_rules", Description: "List configured TraceKit alert rules.", InputSchema: schema("object", map[string]interface{}{}, []string{})},
		{Name: "tracekit_triage_inbox", Description: "List triage inbox items with optional severity, status, and team filters.", InputSchema: schema("object", map[string]interface{}{"severity": stringProp("Optional severity filter"), "entity_type": stringProp("Optional entity type filter"), "status": stringProp("Optional status filter"), "team": stringProp("Optional team filter")}, []string{})},
	}
	lookup := make(map[string]tool, len(tools))
	for _, t := range tools {
		lookup[t.Name] = t
	}
	return &mcpServer{tools: lookup}
}

func (s *mcpServer) handle(req rpcRequest) rpcResponse {
	resp := rpcResponse{JSONRPC: "2.0", ID: parseID(req.ID)}
	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{"protocolVersion": protocolVersion, "capabilities": map[string]interface{}{"tools": map[string]interface{}{}}, "serverInfo": map[string]interface{}{"name": "tracekit-agent", "version": versionString()}}
	case "notifications/initialized":
		return rpcResponse{}
	case "ping":
		resp.Result = map[string]interface{}{}
	case "tools/list":
		order := []string{"tracekit_status", "tracekit_dashboard", "tracekit_services", "tracekit_service_detail", "tracekit_traces", "tracekit_alert_rules", "tracekit_triage_inbox"}
		toolList := make([]tool, 0, len(order))
		for _, name := range order {
			toolList = append(toolList, s.tools[name])
		}
		resp.Result = map[string]interface{}{"tools": toolList}
	case "tools/call":
		result, err := s.callTool(req.Params)
		if err != nil {
			resp.Result = toolErrorResult(err)
		} else {
			resp.Result = toolSuccessResult(result)
		}
	default:
		resp.Error = &rpcError{Code: -32601, Message: "method not found: " + req.Method}
	}
	return resp
}

func (s *mcpServer) callTool(raw json.RawMessage) (interface{}, error) {
	var params callToolParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("invalid tool call params: %w", err)
	}
	if _, ok := s.tools[params.Name]; !ok {
		return nil, fmt.Errorf("unknown tool %q", params.Name)
	}
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("TraceKit credentials not available: %w", err)
	}
	client := newTracekitClient(cfg)
	args := params.Arguments
	if args == nil {
		args = map[string]interface{}{}
	}
	switch params.Name {
	case "tracekit_status":
		return client.getStatus()
	case "tracekit_dashboard":
		return client.getDashboard(getString(args, "window", "24h"))
	case "tracekit_services":
		return client.getServices()
	case "tracekit_service_detail":
		serviceName := strings.TrimSpace(getString(args, "service_name", ""))
		if serviceName == "" {
			return nil, fmt.Errorf("service_name is required")
		}
		return client.getServiceDetail(serviceName)
	case "tracekit_traces":
		return client.getTraces(getString(args, "service", ""), getBool(args, "has_error", false), getInt(args, "min_duration_ms", 0), getString(args, "time_window", "24h"), getInt(args, "limit", 20), getInt(args, "offset", 0))
	case "tracekit_alert_rules":
		return client.getAlertRules()
	case "tracekit_triage_inbox":
		return client.getTriageInbox(getString(args, "severity", ""), getString(args, "entity_type", ""), getString(args, "status", ""), getString(args, "team", ""))
	default:
		return nil, fmt.Errorf("tool %q is not implemented", params.Name)
	}
}

func schema(typeName string, properties map[string]interface{}, required []string) map[string]interface{} {
	return map[string]interface{}{"type": typeName, "properties": properties, "required": required}
}
func stringProp(description string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "description": description}
}
func boolProp(description string) map[string]interface{} {
	return map[string]interface{}{"type": "boolean", "description": description}
}
func intProp(description string) map[string]interface{} {
	return map[string]interface{}{"type": "integer", "description": description}
}
func enumStringProp(description string, values []string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "description": description, "enum": values}
}
func toolSuccessResult(data interface{}) map[string]interface{} {
	encoded, _ := json.MarshalIndent(data, "", "  ")
	return map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": string(encoded)}}, "isError": false}
}
func toolErrorResult(err error) map[string]interface{} {
	return map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": err.Error()}}, "isError": true}
}
func getString(args map[string]interface{}, key, fallback string) string {
	if v, ok := args[key].(string); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}
func getBool(args map[string]interface{}, key string, fallback bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return fallback
}
func getInt(args map[string]interface{}, key string, fallback int) int {
	switch v := args[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		if p, err := strconv.Atoi(v); err == nil {
			return p
		}
	}
	return fallback
}
func parseID(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var out interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}
func readMessage(r *bufio.Reader) ([]byte, error) {
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		return line, nil
	}
}
func writeResponse(resp rpcResponse) {
	if resp.JSONRPC == "" {
		return
	}
	body, err := json.Marshal(resp)
	if err != nil {
		return
	}
	body = append(body, '\n')
	_, _ = os.Stdout.Write(body)
}
