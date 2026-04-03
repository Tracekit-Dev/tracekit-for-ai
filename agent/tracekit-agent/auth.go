package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

const defaultService = "tracekit-ai"

func runAuth(args []string) int {
	if len(args) == 0 {
		printAuthUsage()
		return 1
	}
	switch args[0] {
	case "status":
		return runAuthStatus(args[1:])
	case "register":
		return runAuthRegister(args[1:])
	case "verify":
		return runAuthVerify(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown auth command %q\n", args[0])
		printAuthUsage()
		return 1
	}
}

func runAuthStatus(args []string) int {
	fs := flag.NewFlagSet("auth status", flag.ContinueOnError)
	endpoint := fs.String("endpoint", defaultEndpoint, "TraceKit endpoint")
	configPath := fs.String("config-path", mustConfigPath(), "Config path")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 1
	}
	cfgFile, err := readConfigFile(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	e := normalizeURL(*endpoint)
	profile := cfgFile.Profiles[e]
	payload := map[string]any{"configured": profile != nil && profile.APIKey != "", "active": cfgFile.Active, "endpoint": e, "profile": profile}
	printJSON(payload)
	if configured, _ := payload["configured"].(bool); configured {
		return 0
	}
	return 1
}

func runAuthRegister(args []string) int {
	fs := flag.NewFlagSet("auth register", flag.ContinueOnError)
	email := fs.String("email", "", "Email address")
	name := fs.String("name", "", "Display name")
	org := fs.String("organization-name", "", "Organization name")
	service := fs.String("service-name", defaultService, "Service name")
	endpoint := fs.String("endpoint", defaultEndpoint, "TraceKit endpoint")
	source := fs.String("source", "tracekit-agent", "Source label")
	assistant := fs.String("assistant", "codex", "Assistant label")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if strings.TrimSpace(*email) == "" {
		fmt.Fprintln(os.Stderr, "--email is required")
		return 1
	}
	client := newAnonymousClient(*endpoint)
	resp, err := client.register(registerRequest{Email: *email, Name: strings.TrimSpace(*name), OrganizationName: firstNonEmpty(*org, *service, defaultService), ServiceName: firstNonEmpty(*service, defaultService), Source: *source, SourceMetadata: map[string]interface{}{"assistant": *assistant, "workflow": "tracekit-auth-skill"}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	printJSON(resp)
	return 0
}

func runAuthVerify(args []string) int {
	fs := flag.NewFlagSet("auth verify", flag.ContinueOnError)
	sessionID := fs.String("session-id", "", "Verification session id")
	code := fs.String("code", "", "Verification code")
	service := fs.String("service-name", defaultService, "Service name")
	endpoint := fs.String("endpoint", defaultEndpoint, "TraceKit endpoint")
	tag := fs.String("tag", defaultTag, "Profile tag")
	configPath := fs.String("config-path", mustConfigPath(), "Config path")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if strings.TrimSpace(*sessionID) == "" || strings.TrimSpace(*code) == "" {
		fmt.Fprintln(os.Stderr, "--session-id and --code are required")
		return 1
	}
	client := newAnonymousClient(*endpoint)
	resp, err := client.verify(verifyRequest{SessionID: *sessionID, Code: *code})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cfgFile, err := readConfigFile(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	e := normalizeURL(*endpoint)
	cfgFile.Active = e
	cfgFile.Profiles[e] = &tracekitConfig{APIKey: resp.APIKey, UserID: resp.UserID, Endpoint: e, ServiceName: firstNonEmpty(*service, resp.ServiceName, defaultService), Enabled: "true", CodeMonitoringEnabled: "true", Tag: *tag}
	if err := saveConfigFile(*configPath, cfgFile); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	printJSON(map[string]any{"saved": true, "config_path": *configPath, "endpoint": e, "dashboard_url": resp.DashboardURL, "service_name": cfgFile.Profiles[e].ServiceName, "user_id": resp.UserID, "api_key_masked": maskAPIKey(resp.APIKey)})
	return 0
}

func printAuthUsage() {
	fmt.Fprintln(os.Stderr, "usage: tracekit-agent auth <status|register|verify> [flags]")
}

func mustConfigPath() string {
	path, err := globalConfigPath()
	if err != nil {
		return ".tracekitconfig"
	}
	return path
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func printJSON(v any) {
	encoded, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(encoded))
}
