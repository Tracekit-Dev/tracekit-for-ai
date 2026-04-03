package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type tracekitClient struct {
	BaseURL    string
	APIKey     string
	UserID     string
	HTTPClient *http.Client
}

type errorResponse struct {
	Error string `json:"error"`
}

type registerRequest struct {
	Email            string                 `json:"email"`
	Name             string                 `json:"name,omitempty"`
	OrganizationName string                 `json:"organization_name"`
	ServiceName      string                 `json:"service_name"`
	Source           string                 `json:"source"`
	SourceMetadata   map[string]interface{} `json:"source_metadata,omitempty"`
}

type registerResponse struct {
	VerificationRequired bool   `json:"verification_required"`
	SessionID            string `json:"session_id"`
	Message              string `json:"message"`
}

type verifyRequest struct {
	SessionID string `json:"session_id"`
	Code      string `json:"code"`
}

type verifyResponse struct {
	APIKey         string `json:"api_key"`
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	ServiceName    string `json:"service_name"`
	DashboardURL   string `json:"dashboard_url"`
}

type dashboardData struct {
	Stats struct {
		HealthScore float64 `json:"health_score"`
		Services    int     `json:"services"`
		TotalTraces int     `json:"total_traces"`
		Errors24h   int     `json:"errors_24h"`
		AvgResponse int     `json:"avg_response"`
		Traces24h   int     `json:"traces_24h"`
		ErrorRate   float64 `json:"error_rate"`
	} `json:"stats"`
	Services []struct {
		Name        string  `json:"name"`
		Traces      int     `json:"traces"`
		Errors      int     `json:"errors"`
		ErrorRate   float64 `json:"error_rate"`
		AvgResponse int     `json:"avg_response"`
	} `json:"services"`
	Alerts struct {
		Count int `json:"count"`
		Items []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Severity    string `json:"severity"`
			Message     string `json:"message"`
			TriggeredAt string `json:"triggered_at"`
			Duration    string `json:"duration"`
		} `json:"items"`
	} `json:"alerts"`
	Timestamp string `json:"timestamp"`
}

type serviceHealthListResponse struct {
	Services []serviceHealth `json:"services"`
}

type serviceHealth struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	RequestCount int64   `json:"request_count"`
	ErrorRate    float64 `json:"error_rate"`
	P95Latency   float64 `json:"p95_latency"`
	AvgLatency   float64 `json:"avg_latency"`
	LastSeen     string  `json:"last_seen"`
}

type serviceDetail struct {
	Name         string         `json:"name"`
	RequestCount int64          `json:"request_count"`
	ErrorRate    float64        `json:"error_rate"`
	P50Latency   float64        `json:"p50_latency"`
	P95Latency   float64        `json:"p95_latency"`
	P99Latency   float64        `json:"p99_latency"`
	AvgLatency   float64        `json:"avg_latency"`
	TopErrors    []errorSummary `json:"top_errors"`
	Operations   []serviceOp    `json:"operations"`
}

type errorSummary struct {
	Message string `json:"message"`
	Count   int64  `json:"count"`
}

type serviceOp struct {
	OperationName string  `json:"operation_name"`
	Count         int64   `json:"count"`
	ErrorRate     float64 `json:"error_rate"`
	P95Latency    float64 `json:"p95_latency"`
	AvgLatency    float64 `json:"avg_latency"`
}

type traceListResponse struct {
	Traces     []traceSummary `json:"traces"`
	TotalCount int            `json:"total_count"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
}

type traceSummary struct {
	ID            string `json:"id"`
	TraceID       string `json:"trace_id"`
	ServiceName   string `json:"service_name"`
	OperationName string `json:"operation_name"`
	StartTime     string `json:"start_time"`
	DurationMs    int    `json:"duration_ms"`
	StatusCode    string `json:"status_code"`
	HasError      bool   `json:"has_error"`
	SpanCount     int    `json:"span_count"`
}

type alertRulesResponse struct {
	Rules []alertRule `json:"rules"`
}

type alertRule struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Enabled    bool    `json:"enabled"`
	AlertType  string  `json:"alert_type"`
	ScopeType  string  `json:"scope_type"`
	ScopeValue json.RawMessage `json:"scope_value,omitempty"`
	Metric     string  `json:"metric"`
	Operator   string  `json:"operator"`
	Threshold  float64 `json:"threshold"`
	TimeWindow int     `json:"time_window"`
	Cooldown   int     `json:"cooldown"`
	Severity   string  `json:"severity"`
}

type triageInboxResponse struct {
	Items      []triageItem `json:"items"`
	TotalCount int          `json:"total_count"`
}

type triageItem struct {
	ID              string `json:"id"`
	EntityType      string `json:"entity_type"`
	Title           string `json:"title"`
	Severity        string `json:"severity"`
	Status          string `json:"status"`
	ServiceName     string `json:"service_name"`
	EscalationLevel int    `json:"escalation_level"`
	Timestamp       string `json:"timestamp"`
}

type traceDetailResponse struct {
	Trace traceSummary `json:"trace"`
	Spans []span       `json:"spans"`
}

type span struct {
	ID            string  `json:"id"`
	SpanID        string  `json:"span_id"`
	ParentSpanID  *string `json:"parent_span_id"`
	ServiceName   string  `json:"service_name"`
	OperationName string  `json:"operation_name"`
	Kind          string  `json:"kind"`
	StartTime     string  `json:"start_time"`
	DurationMs    int     `json:"duration_ms"`
	StatusCode    string  `json:"status_code"`
}

type serviceErrorsResponse struct {
	Errors []serviceError `json:"errors"`
}

type serviceError struct {
	SpanID     string  `json:"span_id"`
	Operation  string  `json:"operation"`
	Message    string  `json:"message"`
	DurationMs float64 `json:"duration_ms"`
	Timestamp  string  `json:"timestamp"`
}

type alertHistoryResponse struct {
	History []alertHistoryEntry `json:"history"`
}

type alertHistoryEntry struct {
	ID           string  `json:"id"`
	AlertRuleID  string  `json:"alert_rule_id"`
	TriggeredAt  string  `json:"triggered_at"`
	CurrentValue float64 `json:"current_value"`
	Threshold    float64 `json:"threshold"`
	Message      string  `json:"message"`
	Status       string  `json:"status"`
}

type channelsResponse struct {
	Channels []channel `json:"channels"`
}

type channel struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

type createAlertRuleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	AlertType   string   `json:"alert_type"`
	ScopeType   string   `json:"scope_type"`
	ScopeValue  string   `json:"scope_value,omitempty"`
	Metric      string   `json:"metric"`
	Operator    string   `json:"operator"`
	Threshold   float64  `json:"threshold"`
	TimeWindow  int      `json:"time_window"`
	Cooldown    int      `json:"cooldown"`
	Severity    string   `json:"severity"`
	ChannelIDs  []string `json:"channel_ids"`
}

type releaseListResponse struct {
	Releases []map[string]interface{} `json:"releases"`
	Total    int                      `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

func newTracekitClient(cfg *tracekitConfig) *tracekitClient {
	return &tracekitClient{BaseURL: normalizeURL(cfg.Endpoint), APIKey: strings.TrimSpace(cfg.APIKey), UserID: strings.TrimSpace(cfg.UserID), HTTPClient: &http.Client{Timeout: 30 * time.Second}}
}

func newAnonymousClient(endpoint string) *tracekitClient {
	return &tracekitClient{BaseURL: normalizeURL(endpoint), HTTPClient: &http.Client{Timeout: 30 * time.Second}}
}

func (c *tracekitClient) setAuthHeaders(req *http.Request) {
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}
	if c.UserID != "" {
		req.Header.Set("X-User-ID", c.UserID)
	}
}

func (c *tracekitClient) getJSON(path string, query url.Values, dest any) error {
	if c.APIKey == "" {
		return fmt.Errorf("TraceKit API key is missing; connect auth first")
	}
	endpoint := c.BaseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return parseAPIError(resp.StatusCode, body)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	return nil
}

func (c *tracekitClient) postJSON(path string, payload any, headers map[string]string, dest any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	c.setAuthHeaders(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, respBody)
	}
	if dest == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, dest); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	return nil
}

func parseAPIError(statusCode int, body []byte) error {
	var parsed errorResponse
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != "" {
		return fmt.Errorf("TraceKit API error (%d): %s", statusCode, parsed.Error)
	}
	return fmt.Errorf("TraceKit API error (%d): %s", statusCode, strings.TrimSpace(string(body)))
}

func (c *tracekitClient) register(req registerRequest) (*registerResponse, error) {
	var out registerResponse
	headers := map[string]string{}
	if req.Source != "" {
		headers["X-TraceKit-Source"] = req.Source
	}
	if err := c.postJSON("/v1/integrate/register", req, headers, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *tracekitClient) verify(req verifyRequest) (*verifyResponse, error) {
	var out verifyResponse
	if err := c.postJSON("/v1/integrate/verify", req, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *tracekitClient) getStatus() (map[string]any, error) {
	var out map[string]any
	err := c.getJSON("/v1/integrate/status", nil, &out)
	return out, err
}

func (c *tracekitClient) getDashboard(window string) (*dashboardData, error) {
	var out dashboardData
	query := url.Values{}
	if window != "" {
		query.Set("window", window)
	}
	err := c.getJSON("/v1/alerts/dashboard", query, &out)
	return &out, err
}

func (c *tracekitClient) getServices() (*serviceHealthListResponse, error) {
	var out serviceHealthListResponse
	err := c.getJSON("/v1/services", nil, &out)
	return &out, err
}

func (c *tracekitClient) getServiceDetail(serviceName string) (*serviceDetail, error) {
	var out serviceDetail
	err := c.getJSON("/v1/services/"+url.PathEscape(serviceName)+"/detail", nil, &out)
	return &out, err
}

func (c *tracekitClient) getTraces(service string, hasError bool, minDurationMs int, timeWindow string, limit, offset int) (*traceListResponse, error) {
	var out traceListResponse
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))
	query.Set("sort_by", "start_time")
	query.Set("sort_order", "desc")
	if service != "" {
		query.Set("service", service)
	}
	if hasError {
		query.Set("has_error", "true")
	}
	if minDurationMs > 0 {
		query.Set("min_duration", strconv.Itoa(minDurationMs))
	}
	if timeWindow != "" && timeWindow != "all" {
		now := time.Now().UTC()
		var from time.Time
		switch timeWindow {
		case "1h":
			from = now.Add(-1 * time.Hour)
		case "6h":
			from = now.Add(-6 * time.Hour)
		case "24h":
			from = now.Add(-24 * time.Hour)
		}
		if !from.IsZero() {
			query.Set("start_time_from", from.Format(time.RFC3339))
			query.Set("start_time_to", now.Format(time.RFC3339))
		}
	}
	err := c.getJSON("/v1/traces", query, &out)
	return &out, err
}

func (c *tracekitClient) getAlertRules() (*alertRulesResponse, error) {
	var out alertRulesResponse
	err := c.getJSON("/v1/alert-rules", nil, &out)
	return &out, err
}

func (c *tracekitClient) getTriageInbox(severity, entityType, status, team string) (*triageInboxResponse, error) {
	var out triageInboxResponse
	query := url.Values{}
	if severity != "" {
		query.Set("severity", severity)
	}
	if entityType != "" {
		query.Set("type", entityType)
	}
	if status != "" {
		query.Set("status", status)
	}
	if team != "" {
		query.Set("team", team)
	}
	err := c.getJSON("/v1/triage-inbox", query, &out)
	return &out, err
}

func (c *tracekitClient) getTraceDetail(traceID string) (*traceDetailResponse, error) {
	var out traceDetailResponse
	err := c.getJSON("/v1/traces/"+url.PathEscape(traceID), nil, &out)
	return &out, err
}

func (c *tracekitClient) getServiceErrors(serviceName string) (*serviceErrorsResponse, error) {
	var out serviceErrorsResponse
	err := c.getJSON("/v1/services/"+url.PathEscape(serviceName)+"/errors", nil, &out)
	return &out, err
}

func (c *tracekitClient) getAlertHistory(ruleID string) (*alertHistoryResponse, error) {
	var out alertHistoryResponse
	err := c.getJSON("/v1/alert-rules/"+url.PathEscape(ruleID)+"/history", nil, &out)
	return &out, err
}

func (c *tracekitClient) createAlertRule(req createAlertRuleRequest) (*alertRule, error) {
	var out alertRule
	if err := c.postJSON("/v1/alert-rules", req, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *tracekitClient) deleteAlertRule(ruleID string) error {
	return c.deleteRequest("/v1/alert-rules/" + url.PathEscape(ruleID))
}

func (c *tracekitClient) toggleAlertRule(ruleID string, enabled bool) error {
	body := map[string]bool{"enabled": enabled}
	return c.postJSON("/v1/alert-rules/"+url.PathEscape(ruleID)+"/toggle", body, nil, nil)
}

func (c *tracekitClient) getChannels() (*channelsResponse, error) {
	var out channelsResponse
	err := c.getJSON("/v1/channels", nil, &out)
	return &out, err
}

func (c *tracekitClient) listReleases(page, pageSize int, service string) (*releaseListResponse, error) {
	var out releaseListResponse
	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("page_size", strconv.Itoa(pageSize))
	if service != "" {
		query.Set("service", service)
	}
	err := c.getJSON("/v1/releases", query, &out)
	return &out, err
}

func (c *tracekitClient) triageAction(itemID, action, entityType, note, duration string) error {
	body := map[string]string{"entity_type": entityType}
	if note != "" {
		body["note"] = note
	}
	if duration != "" {
		body["duration"] = duration
	}
	return c.postJSON("/v1/triage-inbox/"+url.PathEscape(itemID)+"/"+action, body, nil, nil)
}

func (c *tracekitClient) deleteRequest(path string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, body)
	}
	return nil
}
