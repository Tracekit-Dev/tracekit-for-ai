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
		{Name: "tracekit_trace_detail", Description: "Get full trace detail including all spans for a given trace ID.", InputSchema: schema("object", map[string]interface{}{"trace_id": stringProp("Trace ID to fetch")}, []string{"trace_id"})},
		{Name: "tracekit_service_errors", Description: "Get recent error spans for a specific service.", InputSchema: schema("object", map[string]interface{}{"service_name": stringProp("TraceKit service name")}, []string{"service_name"})},
		{Name: "tracekit_alert_history", Description: "Get firing history for a specific alert rule.", InputSchema: schema("object", map[string]interface{}{"rule_id": stringProp("Alert rule ID")}, []string{"rule_id"})},
		{Name: "tracekit_create_alert_rule", Description: "Create a new alert rule.", InputSchema: schema("object", map[string]interface{}{"name": stringProp("Alert rule name"), "alert_type": enumStringProp("Alert type", []string{"threshold", "anomaly"}), "scope_type": enumStringProp("Scope type", []string{"global", "service"}), "scope_value": stringProp("Service name when scope_type is service"), "metric": enumStringProp("Metric to monitor", []string{"error_rate", "latency_p95", "latency_p99", "latency_avg", "request_count"}), "operator": enumStringProp("Comparison operator", []string{">", "<", ">=", "<=", "=="}), "threshold": floatProp("Threshold value"), "time_window": intProp("Time window in minutes"), "cooldown": intProp("Cooldown in minutes between alerts"), "severity": enumStringProp("Alert severity", []string{"critical", "warning", "info"}), "channel_ids": arrayProp("Notification channel IDs")}, []string{"name", "alert_type", "metric", "operator", "threshold", "time_window", "severity"})},
		{Name: "tracekit_delete_alert_rule", Description: "Delete an alert rule by ID.", InputSchema: schema("object", map[string]interface{}{"rule_id": stringProp("Alert rule ID to delete")}, []string{"rule_id"})},
		{Name: "tracekit_toggle_alert_rule", Description: "Enable or disable an alert rule.", InputSchema: schema("object", map[string]interface{}{"rule_id": stringProp("Alert rule ID"), "enabled": boolProp("Whether the rule should be enabled")}, []string{"rule_id", "enabled"})},
		{Name: "tracekit_channels", Description: "List configured notification channels.", InputSchema: schema("object", map[string]interface{}{}, []string{})},
		{Name: "tracekit_releases", Description: "List releases with optional service filter.", InputSchema: schema("object", map[string]interface{}{"service": stringProp("Optional service filter"), "page": intProp("Page number (default 1)"), "page_size": intProp("Results per page (default 20)")}, []string{})},
		{Name: "tracekit_triage_acknowledge", Description: "Acknowledge a triage inbox item.", InputSchema: schema("object", map[string]interface{}{"item_id": stringProp("Triage item ID"), "entity_type": stringProp("Entity type of the item")}, []string{"item_id", "entity_type"})},
		{Name: "tracekit_triage_investigate", Description: "Mark a triage inbox item as under investigation.", InputSchema: schema("object", map[string]interface{}{"item_id": stringProp("Triage item ID"), "entity_type": stringProp("Entity type of the item")}, []string{"item_id", "entity_type"})},
		{Name: "tracekit_triage_resolve", Description: "Resolve a triage inbox item with an optional note.", InputSchema: schema("object", map[string]interface{}{"item_id": stringProp("Triage item ID"), "entity_type": stringProp("Entity type of the item"), "note": stringProp("Optional resolution note")}, []string{"item_id", "entity_type"})},
		{Name: "tracekit_triage_snooze", Description: "Snooze a triage inbox item for a specified duration.", InputSchema: schema("object", map[string]interface{}{"item_id": stringProp("Triage item ID"), "entity_type": stringProp("Entity type of the item"), "duration": enumStringProp("Snooze duration", []string{"1h", "4h", "24h", "7d"})}, []string{"item_id", "entity_type", "duration"})},
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
		order := []string{"tracekit_status", "tracekit_dashboard", "tracekit_services", "tracekit_service_detail", "tracekit_service_errors", "tracekit_traces", "tracekit_trace_detail", "tracekit_alert_rules", "tracekit_alert_history", "tracekit_create_alert_rule", "tracekit_delete_alert_rule", "tracekit_toggle_alert_rule", "tracekit_channels", "tracekit_releases", "tracekit_triage_inbox", "tracekit_triage_acknowledge", "tracekit_triage_investigate", "tracekit_triage_resolve", "tracekit_triage_snooze"}
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
	case "tracekit_trace_detail":
		traceID := strings.TrimSpace(getString(args, "trace_id", ""))
		if traceID == "" {
			return nil, fmt.Errorf("trace_id is required")
		}
		return client.getTraceDetail(traceID)
	case "tracekit_service_errors":
		serviceName := strings.TrimSpace(getString(args, "service_name", ""))
		if serviceName == "" {
			return nil, fmt.Errorf("service_name is required")
		}
		return client.getServiceErrors(serviceName)
	case "tracekit_alert_history":
		ruleID := strings.TrimSpace(getString(args, "rule_id", ""))
		if ruleID == "" {
			return nil, fmt.Errorf("rule_id is required")
		}
		return client.getAlertHistory(ruleID)
	case "tracekit_create_alert_rule":
		name := strings.TrimSpace(getString(args, "name", ""))
		if name == "" {
			return nil, fmt.Errorf("name is required")
		}
		req := createAlertRuleRequest{
			Name:       name,
			AlertType:  getString(args, "alert_type", "threshold"),
			ScopeType:  getString(args, "scope_type", "global"),
			ScopeValue: getString(args, "scope_value", ""),
			Metric:     getString(args, "metric", ""),
			Operator:   getString(args, "operator", ""),
			Threshold:  getFloat(args, "threshold", 0),
			TimeWindow: getInt(args, "time_window", 5),
			Cooldown:   getInt(args, "cooldown", 10),
			Severity:   getString(args, "severity", "warning"),
			ChannelIDs: getStringArray(args, "channel_ids"),
		}
		return client.createAlertRule(req)
	case "tracekit_delete_alert_rule":
		ruleID := strings.TrimSpace(getString(args, "rule_id", ""))
		if ruleID == "" {
			return nil, fmt.Errorf("rule_id is required")
		}
		if err := client.deleteAlertRule(ruleID); err != nil {
			return nil, err
		}
		return map[string]string{"status": "deleted", "rule_id": ruleID}, nil
	case "tracekit_toggle_alert_rule":
		ruleID := strings.TrimSpace(getString(args, "rule_id", ""))
		if ruleID == "" {
			return nil, fmt.Errorf("rule_id is required")
		}
		enabled := getBool(args, "enabled", true)
		if err := client.toggleAlertRule(ruleID, enabled); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "updated", "rule_id": ruleID, "enabled": enabled}, nil
	case "tracekit_channels":
		return client.getChannels()
	case "tracekit_releases":
		return client.listReleases(getInt(args, "page", 1), getInt(args, "page_size", 20), getString(args, "service", ""))
	case "tracekit_triage_acknowledge":
		return nil, client.triageAction(getString(args, "item_id", ""), "acknowledge", getString(args, "entity_type", ""), "", "")
	case "tracekit_triage_investigate":
		return nil, client.triageAction(getString(args, "item_id", ""), "investigate", getString(args, "entity_type", ""), "", "")
	case "tracekit_triage_resolve":
		return nil, client.triageAction(getString(args, "item_id", ""), "resolve", getString(args, "entity_type", ""), getString(args, "note", ""), "")
	case "tracekit_triage_snooze":
		return nil, client.triageAction(getString(args, "item_id", ""), "snooze", getString(args, "entity_type", ""), "", getString(args, "duration", "1h"))
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
func floatProp(description string) map[string]interface{} {
	return map[string]interface{}{"type": "number", "description": description}
}
func arrayProp(description string) map[string]interface{} {
	return map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": description}
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
func getFloat(args map[string]interface{}, key string, fallback float64) float64 {
	switch v := args[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			return p
		}
	}
	return fallback
}
func getStringArray(args map[string]interface{}, key string) []string {
	val, ok := args[key]
	if !ok || val == nil {
		return nil
	}
	if arr, ok := val.([]interface{}); ok {
		out := make([]string, 0, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
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
