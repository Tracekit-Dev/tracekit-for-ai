package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultEndpoint = "https://app.tracekit.dev"
	defaultTag      = "prod"
)

type tracekitConfig struct {
	APIKey                string `json:"api_key"`
	UserID                string `json:"user_id"`
	Endpoint              string `json:"endpoint"`
	ServiceName           string `json:"service_name,omitempty"`
	Enabled               string `json:"enabled,omitempty"`
	CodeMonitoringEnabled string `json:"code_monitoring_enabled,omitempty"`
	Tag                   string `json:"tag,omitempty"`
}

type configFile struct {
	Active   string                     `json:"active"`
	Profiles map[string]*tracekitConfig `json:"profiles"`
}

func loadConfig() (*tracekitConfig, error) {
	apiKey := strings.TrimSpace(os.Getenv("TRACEKIT_API_KEY"))
	userID := strings.TrimSpace(os.Getenv("TRACEKIT_USER_ID"))
	endpoint := strings.TrimSpace(os.Getenv("TRACEKIT_ENDPOINT"))
	if apiKey != "" {
		if endpoint == "" {
			endpoint = defaultEndpoint
		}
		return &tracekitConfig{APIKey: apiKey, UserID: userID, Endpoint: normalizeURL(endpoint)}, nil
	}

	path, err := globalConfigPath()
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var cfgFile configFile
	if err := json.Unmarshal(content, &cfgFile); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	if cfgFile.Active == "" || len(cfgFile.Profiles) == 0 {
		return nil, fmt.Errorf("no active TraceKit profile found in %s", path)
	}

	profile, ok := cfgFile.Profiles[cfgFile.Active]
	if !ok || profile == nil {
		return nil, fmt.Errorf("active TraceKit profile %q not found in %s", cfgFile.Active, path)
	}
	if profile.APIKey == "" {
		return nil, fmt.Errorf("active TraceKit profile %q is missing api_key", cfgFile.Active)
	}
	if profile.Endpoint == "" {
		profile.Endpoint = cfgFile.Active
	}
	profile.Endpoint = normalizeURL(profile.Endpoint)
	return profile, nil
}

func readConfigFile(path string) (*configFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &configFile{Active: defaultEndpoint, Profiles: map[string]*tracekitConfig{}}, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var cfgFile configFile
	if err := json.Unmarshal(content, &cfgFile); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	if cfgFile.Profiles == nil {
		cfgFile.Profiles = map[string]*tracekitConfig{}
	}
	if cfgFile.Active == "" {
		cfgFile.Active = defaultEndpoint
	}
	return &cfgFile, nil
}

func saveConfigFile(path string, payload *configFile) error {
	if payload.Profiles == nil {
		payload.Profiles = map[string]*tracekitConfig{}
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func globalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve home directory: %w", err)
	}
	return filepath.Join(home, ".tracekitconfig"), nil
}

func normalizeURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimRight(trimmed, "/")
	if trimmed == "" {
		return defaultEndpoint
	}
	return trimmed
}

func maskAPIKey(value string) string {
	if len(value) <= 10 {
		return value
	}
	return value[:10] + "..." + value[len(value)-6:]
}
