---
name: tracekit-go-sdk
description: Sets up TraceKit APM in Go services for automatic distributed tracing, error capture, and code monitoring. Supports Gin, Echo, and net/http frameworks. Includes LLM instrumentation via http.RoundTripper transport wrapper for OpenAI and Anthropic API call monitoring. Use when the user asks to add TraceKit, add observability, instrument a Go service, or configure APM in a Go project.
---

# TraceKit Go SDK Setup

## Auth Bootstrap

Do not tell the user to go sign up, log in separately, or manually create an API key before setup. First check for existing TraceKit auth. If `~/.tracekitconfig` does not contain the production profile or `TRACEKIT_API_KEY` is missing, apply the `tracekit-auth` skill first. Use `./scripts/run-tracekit-auth.sh status` to check, then guide the user through the TraceKit email verification flow with `register` and `verify`. That flow signs the user into an existing account for that email or creates the account automatically, then saves the returned credentials for the rest of the setup.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to a Go service
- Add observability or APM to a Go application
- Instrument a Go service with distributed tracing
- Configure TraceKit API keys in a Go project
- Debug production Go services with live breakpoints
- Set up code monitoring in a Go app
- Monitor OpenAI or Anthropic API calls in a Go service
- Add LLM observability to a Go application

## Non-Negotiable Rules

1. **Never hardcode API keys** in code. Always use `os.Getenv("TRACEKIT_API_KEY")`.
2. **Always initialize TraceKit before registering routes**  - middleware and handlers must be added after `tracekit.NewSDK()`.
3. **Always call `defer sdk.Shutdown(context.Background())`** to flush pending traces on exit.
4. **Always include a verification step** confirming traces appear in `https://app.tracekit.dev/traces`.
5. **Always enable code monitoring** (`EnableCodeMonitoring: true`)  - it is TraceKit's differentiator.

## Detection

Before applying this skill, detect the project type:

1. **Check for `go.mod`**  - confirms this is a Go project.
2. **Detect framework** by scanning `go.mod` and import statements:
   - `github.com/gin-gonic/gin` => Gin framework (use Gin branch)
   - `github.com/labstack/echo` => Echo framework (use Echo branch)
   - No framework imports => plain `net/http` (use net/http branch)
3. **Only ask the user** if multiple frameworks are detected or if `go.mod` is missing.

## Step 1: Environment Setup

Set the `TRACEKIT_API_KEY` environment variable. This is the only required secret.

Add to your `.env` file or environment:

```bash
export TRACEKIT_API_KEY=ctxio_your_api_key_here
```

The OTLP endpoint is hardcoded in the SDK init  - no need to configure it separately.

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. Use `.env` files, deployment secret managers, or CI variables.

## Step 2: Install SDK

```bash
go get github.com/Tracekit-Dev/go-sdk
```

This installs the TraceKit Go SDK with built-in OpenTelemetry support, framework middleware, and code monitoring.

## Step 3: Initialize TraceKit

Add this to your `main()` function, **before** any route or handler registration:

```go
package main

import (
    "context"
    "log"
    "os"

    tracekit "github.com/Tracekit-Dev/go-sdk"
)

func main() {
    // Initialize TraceKit  - MUST be before routes
    sdk, err := tracekit.NewSDK(&tracekit.Config{
        APIKey:               os.Getenv("TRACEKIT_API_KEY"),
        ServiceName:          "my-go-service",
        Endpoint:             "https://app.tracekit.dev/v1/traces",
        EnableCodeMonitoring: true,
    })
    if err != nil {
        log.Fatalf("tracekit init failed: %v", err)
    }
    defer sdk.Shutdown(context.Background())

    // ... register routes and start server below
}
```

**Key points:**
- `ServiceName` should match your service's logical name (e.g., `"api-gateway"`, `"user-service"`)
- `EnableCodeMonitoring: true` enables live breakpoints and snapshots in production
- `defer sdk.Shutdown(...)` ensures all pending traces are flushed before the process exits

## Step 4: Framework Integration

Choose the branch matching your framework. Apply **one** of the following.

### Branch A: Gin

```go
import "github.com/gin-gonic/gin"

func main() {
    // ... TraceKit init from Step 3 ...

    r := gin.Default()

    // Add TraceKit middleware  - auto-traces all routes
    r.Use(sdk.GinMiddleware())

    r.GET("/api/users", func(c *gin.Context) {
        c.JSON(200, gin.H{"users": []string{"alice", "bob"}})
    })

    r.Run(":8080")
}
```

### Branch B: Echo

```go
import "github.com/labstack/echo/v4"

func main() {
    // ... TraceKit init from Step 3 ...

    e := echo.New()

    // Add TraceKit middleware  - auto-traces all routes
    e.Use(sdk.EchoMiddleware())

    e.GET("/api/users", func(c echo.Context) error {
        return c.JSON(200, map[string]any{"users": []string{"alice", "bob"}})
    })

    e.Start(":8080")
}
```

### Branch C: net/http (standard library)

```go
import "net/http"

func main() {
    // ... TraceKit init from Step 3 ...

    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"users": ["alice", "bob"]}`))
    })

    // Wrap your handler with TraceKit
    tracedHandler := sdk.HTTPHandler(mux)

    http.ListenAndServe(":8080", tracedHandler)
}
```

## Step 5: Error Capture

Capture errors explicitly where you handle them:

```go
result, err := someOperation()
if err != nil {
    sdk.CaptureException(err)
    // handle the error...
}
```

For adding context to traces, use manual spans:

```go
span := sdk.StartSpan("process-order", map[string]string{
    "order.id": orderID,
    "user.id":  userID,
})
defer span.End()

// ... your business logic ...
```

## Step 5b: Snapshot Capture (Code Monitoring)

For programmatic snapshots, **use the SnapshotClient directly**  - do not call through the SDK wrapper. The SDK uses `runtime.Caller(2)` internally to identify the call site. Adding extra layers shifts the frame count and causes snapshots to report the wrong source location.

Create a thin wrapper package (e.g., `internal/breakpoints/breakpoints.go`):

```go
package breakpoints

import (
    "context"

    "github.com/Tracekit-Dev/go-sdk/tracekit"
)

var snapshotClient *tracekit.SnapshotClient

// Init stores the snapshot client. No-op when sdk is nil.
func Init(s *tracekit.SDK) {
    if s != nil {
        snapshotClient = s.SnapshotClient()
    }
}

// Capture fires a code monitoring snapshot. No-op when tracing is disabled.
func Capture(ctx context.Context, name string, data map[string]interface{}) {
    if snapshotClient == nil {
        return
    }
    snapshotClient.CheckAndCaptureWithContext(ctx, name, data)
}
```

Initialize after SDK setup in `main()`:

```go
breakpoints.Init(sdk)
```

Use at call sites:

```go
breakpoints.Capture(ctx, "payment-failed", map[string]interface{}{
    "order_id": order.ID,
    "error":    err.Error(),
})
```

See the `tracekit-code-monitoring` skill for the full pattern across all languages.

### Code Monitoring v25.0 Features

The latest SDK release adds these code monitoring capabilities (configured via the dashboard, no code changes needed):

- **Logpoint mode** -- capture expressions only without full variable snapshots, reducing overhead
- **Per-breakpoint limits** -- individual max captures (default: 100) and rate limits per breakpoint
- **Dynamic stack traces** -- configurable stack depth per breakpoint (1-50 frames)
- **Idle auto-expiry** -- breakpoints auto-expire after inactivity (default: 24h), pinnable to prevent expiry
- **Conditional expressions** -- server-side evaluated conditions with full operator support (`==`, `!=`, `>`, `<`, `&&`, `||`)

These features are available when `EnableCodeMonitoring` is set to `true`. No SDK code changes required -- all configuration is done through the TraceKit dashboard.

For full details, see the `tracekit-code-monitoring` skill.

## Step 6: Verification

After integrating, verify traces are flowing:

1. **Start your application** with `TRACEKIT_API_KEY` set in the environment.
2. **Hit your endpoints 3-5 times**  - e.g., `curl http://localhost:8080/api/users`.
3. **Open** `https://app.tracekit.dev/traces`.
4. **Confirm** new spans and your service name appear within 30-60 seconds.

If traces do not appear, see Troubleshooting below.

## Troubleshooting

### Traces not appearing in dashboard

- **Check `TRACEKIT_API_KEY`:** Ensure the env var is set in the runtime environment (not just in your shell). Print it: `fmt.Println(os.Getenv("TRACEKIT_API_KEY"))`.
- **Check outbound access:** Your service must reach `https://app.tracekit.dev/v1/traces`. Verify with: `curl -X POST https://app.tracekit.dev/v1/traces` (expect 401  - means the endpoint is reachable).
- **Check init order:** `tracekit.NewSDK()` must be called **before** registering routes and middleware. If init happens after routes, requests are not traced.

### Init order wrong

Symptoms: Server starts fine but no traces appear despite traffic.

Fix: Move `tracekit.NewSDK()` to the very beginning of `main()`, before `gin.Default()`, `echo.New()`, or `http.NewServeMux()`.

### Missing environment variable

Symptoms: `tracekit init failed` error on startup, or traces appear without an API key (rejected by backend).

Fix: Ensure `TRACEKIT_API_KEY` is exported in your shell, `.env` file, Docker Compose, or deployment config.

### Service name collisions

Symptoms: Traces appear under the wrong service in the dashboard.

Fix: Use a unique `ServiceName` per deployed service. Avoid generic names like `"app"` or `"server"`.

## LLM Instrumentation (Manual Setup)

TraceKit can instrument OpenAI and Anthropic API calls made via Go's `net/http` using an `http.RoundTripper` transport wrapper.

### When To Use

Add this when the user:
- Uses OpenAI or Anthropic APIs in their Go service
- Wants to monitor LLM cost, tokens, and latency
- Asks about AI observability in Go

### Setup

Wrap the HTTP client's transport with `NewLLMTransport` and pass it to the OpenAI or Anthropic Go SDK:

```go
import (
    tracekit "github.com/Tracekit-Dev/go-sdk/tracekit"
    "net/http"
)

// Default config (content capture off)
transport := tracekit.NewLLMTransport(nil)
httpClient := &http.Client{Transport: transport}

// With content capture enabled
transport := tracekit.NewLLMTransport(nil, tracekit.WithCaptureContent(true))
httpClient := &http.Client{Transport: transport}

// With full custom config
transport := tracekit.NewLLMTransport(nil, tracekit.WithLLMConfig(tracekit.LLMConfig{
    Enabled:        true,
    OpenAI:         true,
    Anthropic:      true,
    CaptureContent: false,
}))
httpClient := &http.Client{Transport: transport}
```

Pass this `httpClient` to your OpenAI or Anthropic SDK's HTTP client configuration. If `nil` is passed as the base transport, `http.DefaultTransport` is used.

### Environment Variable

Set `TRACEKIT_LLM_CAPTURE_CONTENT=true` to enable prompt/completion capture without code changes.

### Captured Attributes

LLM spans include: `gen_ai.system`, `gen_ai.request.model`, `gen_ai.usage.input_tokens`, `gen_ai.usage.output_tokens`, `gen_ai.response.finish_reasons`. Streaming responses produce a single span with accumulated token counts. Tool calls are recorded as `gen_ai.tool.call` span events.

### Verification

After adding the transport, make an LLM API call and verify the span appears in the TraceKit dashboard under **LLM Observability** (`/ai/llm`).

## Next Steps

Once your Go service is traced, consider:
- **Code Monitoring**  - Set live breakpoints and capture snapshots in production without redeploying (already enabled via `EnableCodeMonitoring: true`)
- **Distributed Tracing**  - Connect traces across multiple services for full request visibility
- **Frontend Observability**  - Add `@tracekit/browser` to your frontend for end-to-end trace correlation

## References

- Go SDK docs: `https://app.tracekit.dev/docs/languages/go`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
