---
name: tracekit-code-monitoring
description: Enable live breakpoints and snapshots for production debugging with TraceKit Code Monitoring. Works with all backend SDKs. Use when the user asks to debug production code, set breakpoints, capture variable state, or enable code monitoring.
---

# TraceKit Code Monitoring

## When To Use

Use this skill when the user asks to:
- Debug production code without redeploying
- Set breakpoints on a live service
- Capture snapshots of variable state in production
- Enable code monitoring or live debugging
- Inspect variables in production
- Add logpoints to production code
- Debug specific users or error conditions in production

**This is TraceKit's core differentiator.** Code monitoring lets developers debug production services with less than 5ms overhead per breakpoint hit -- no redeployment, no restart, no code changes required. Set a breakpoint, capture variable state, and link snapshots directly to distributed traces.

## Non-Negotiable Rules

1. **Never hardcode API keys** in code. Always use `TRACEKIT_API_KEY` env var (or the equivalent for your language/framework).
2. **Always include a verification step** confirming snapshots appear in `https://app.tracekit.dev/code-monitoring`.
3. **`enableCodeMonitoring: true` must be set in SDK init** -- all SDK skills include this by default, but verify it is present.
4. **Never set breakpoints on hot paths without rate limiting** -- unrestricted breakpoints on high-throughput code paths can generate excessive snapshot volume.

## Prerequisites

Code monitoring requires a backend SDK to be installed and initialized. Complete **one** of the following SDK skills first:

- `tracekit-go-sdk` -- Go services (Gin, Echo, net/http)
- `tracekit-node-sdk` -- Node.js services (Express, Fastify, Koa, NestJS)
- `tracekit-python-sdk` -- Python services (Django, Flask, FastAPI)
- `tracekit-php-sdk` -- PHP services
- `tracekit-laravel-sdk` -- Laravel applications
- `tracekit-java-sdk` -- Java services (Spring Boot, plain Java)
- `tracekit-dotnet-sdk` -- .NET services (ASP.NET Core, .NET Core)
- `tracekit-ruby-sdk` -- Ruby services (Rails, Sinatra)

The SDK must be initialized with `enableCodeMonitoring: true` (or the language-equivalent config key). All SDK skills already include this setting.

**If no SDK is detected**, redirect to the `tracekit-apm-setup` skill to choose and install the correct SDK first.

## Detection

Detect which backend SDK is installed by scanning project files:

| File to Check | Content to Match | SDK Detected |
|---|---|---|
| `go.mod` | `tracekit` or `Tracekit-Dev` | Go SDK |
| `package.json` | `@tracekit/node` | Node.js SDK |
| `requirements.txt` or `pyproject.toml` | `tracekit-apm` | Python SDK |
| `composer.json` | `tracekit/php-apm` | PHP SDK |
| `composer.json` | `tracekit/laravel-apm` | Laravel SDK |
| `pom.xml` or `build.gradle` | `tracekit` | Java SDK |
| `.csproj` | `TraceKit` | .NET SDK |
| `Gemfile` | `tracekit` | Ruby SDK |

If none of these are found, redirect to the `tracekit-apm-setup` skill.

## Step 1: Verify SDK Configuration

Confirm that `enableCodeMonitoring` is set to `true` in the SDK initialization. This is required for breakpoints and snapshots to work.

### Go

```go
sdk, err := tracekit.NewSDK(&tracekit.Config{
    APIKey:               os.Getenv("TRACEKIT_API_KEY"),
    ServiceName:          "my-go-service",
    Endpoint:             "https://app.tracekit.dev/v1/traces",
    EnableCodeMonitoring: true, // <-- Must be true
})
```

### Node.js

```javascript
const tracekit = require('@tracekit/node');

tracekit.init({
  apiKey: process.env.TRACEKIT_API_KEY,
  serviceName: 'my-node-service',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true, // <-- Must be true
});
```

### Python

```python
import tracekit_apm

tracekit_apm.init(
    api_key=os.environ["TRACEKIT_API_KEY"],
    service_name="my-python-service",
    endpoint="https://app.tracekit.dev/v1/traces",
    enable_code_monitoring=True,  # <-- Must be True
)
```

For PHP, Laravel, Java, .NET, and Ruby, the equivalent config key exists in each SDK. Refer to the respective SDK skill for the exact syntax.

## Step 2: Set Your First Breakpoint (Dashboard)

The fastest way to start is through the TraceKit dashboard:

1. Navigate to `https://app.tracekit.dev/code-monitoring`
2. Click **"New Breakpoint"**
3. Select the **service** from the dropdown (matches your `serviceName` config)
4. Select the **source file** and **line number** where you want to capture state
5. Choose the **breakpoint type** (see Step 3 for types)
6. Set **capture depth** (default: 3 levels of nested objects)
7. Set **rate limit** (default: 10 snapshots per minute)
8. Click **"Activate"**

The breakpoint takes effect within seconds -- no redeployment needed. The SDK polls for active breakpoints and applies them to the running process.

## Step 3: Breakpoint Types

### Line Breakpoint

Captures all local variables when execution hits the specified line.

**Use for:** Investigating a specific code path, seeing what state looks like at a particular point in your logic.

**Example:** Set a line breakpoint on the line where you process a payment to see the full payment object and user context.

### Conditional Breakpoint

Only triggers when a boolean expression evaluates to true. The expression has access to all local variables in scope.

**Use for:** Debugging issues that only affect specific users, requests, or error conditions.

**Example expressions:**
- `user.id == "abc123"` -- capture only for a specific user
- `response.statusCode >= 500` -- capture only on server errors
- `order.total > 10000` -- capture high-value transactions
- `err != nil` (Go) / `err !== null` (Node) / `err is not None` (Python) -- capture only when errors occur

### Logpoint

Injects a log message into the application output without capturing full variable state. Logpoints use template syntax with `{variable}` placeholders.

**Use for:** Lightweight tracing without the overhead of full snapshots. Good for understanding control flow or tracking specific values over time.

**Example log messages:**
- `User {user.id} hit endpoint with status {response.status}`
- `Order {order.id} processing took {duration}ms`
- `Cache {cache.hit ? "HIT" : "MISS"} for key {cache.key}`

Logpoints appear in the Code Monitoring > Logs tab and can be filtered by service, time range, and content.

## Step 4: Snapshot Configuration

Configure how much data breakpoints capture and how often they fire.

| Setting | Default | Description |
|---|---|---|
| **Capture depth** | 3 | How many levels of nested objects to capture. Increase for deeply nested data (max: 10). |
| **Rate limit** | 10/min | Max snapshots per minute per breakpoint. Prevents runaway capture on hot paths. |
| **Expiration** | 24 hours | Auto-remove the breakpoint after this duration. Prevents forgotten breakpoints. |
| **Max string length** | 256 chars | Truncate string values beyond this length. |

### Sensitive Data Redaction

Configure field patterns to exclude from snapshots. Any captured variable whose name matches a pattern is replaced with `[REDACTED]`.

Default redaction patterns:
- `password`
- `token`
- `secret`
- `authorization`
- `cookie`
- `creditCard` / `credit_card`

Add custom patterns in the dashboard under Code Monitoring > Settings > Redaction Patterns, or pass them in the SDK config:

```javascript
// Node.js example
tracekit.init({
  // ... other config
  codeMonitoring: {
    redactFields: ['ssn', 'bankAccount', 'apiSecret'],
  },
});
```

## Step 5: Programmatic Breakpoints (SDK API)

Set breakpoints programmatically from your code instead of the dashboard. Useful for CI/CD integration, automated debugging workflows, or temporary debug sessions.

### Go

```go
tracekit.SetBreakpoint(tracekit.BreakpointConfig{
    File:      "main.go",
    Line:      42,
    Condition: "err != nil",
    RateLimit: 5,          // max 5 snapshots/min
    ExpiresIn: 2 * time.Hour,
})
```

### Node.js

```javascript
const tracekit = require('@tracekit/node');

tracekit.setBreakpoint({
  file: 'src/handlers/payment.js',
  line: 42,
  condition: 'err !== null',
  rateLimit: 5,
  expiresIn: '2h',
});
```

### Python

```python
import tracekit_apm

tracekit_apm.set_breakpoint(
    file='app/handlers/payment.py',
    line=42,
    condition='err is not None',
    rate_limit=5,
    expires_in='2h',
)
```

Programmatic breakpoints appear in the dashboard alongside dashboard-created ones and can be managed from either location.

## Step 6: Viewing Snapshots

Once a breakpoint fires, snapshots are available in the dashboard:

1. Navigate to **Code Monitoring > Snapshots** at `https://app.tracekit.dev/code-monitoring`
2. **Filter** by breakpoint name, time range, service, or environment
3. Click a snapshot to **inspect captured variables**:
   - Local variables with their values at capture time
   - Call stack showing the execution path
   - Trace context linking to the full distributed trace
4. Click the **Trace ID** to jump to the full distributed trace waterfall view
5. Use **Compare** to diff two snapshots side by side (useful for investigating why the same code path produces different results)

### Snapshot Retention

Snapshots are retained for 30 days by default. Adjust in Dashboard > Settings > Data Retention.

## Step 7: Verification

Verify code monitoring is working end to end:

1. **Set a line breakpoint** on a frequently-hit endpoint handler (e.g., a health check or API route)
2. **Make a request** to that endpoint:
   ```bash
   curl http://localhost:8080/api/health
   ```
3. **Within 30 seconds**, check `https://app.tracekit.dev/code-monitoring` for a new snapshot
4. **Verify** the captured variables match expected state (request headers, local variables, etc.)
5. **Click the Trace ID** to confirm the snapshot is linked to a distributed trace
6. **Remove or deactivate** the test breakpoint once verified

If snapshots do not appear, see Troubleshooting below.

## Troubleshooting

### Snapshots not appearing

- **Check `enableCodeMonitoring: true`** is set in your SDK init config. Without this, the SDK does not poll for breakpoints.
- **Check the breakpoint is active** in the dashboard. Expired or paused breakpoints do not fire.
- **Check the service is sending data** -- visit `https://app.tracekit.dev/traces` to confirm traces are flowing from this service.
- **Check the service name matches** -- the breakpoint targets a specific service. Ensure `serviceName` in your SDK init matches the service selected when creating the breakpoint.
- **Check network connectivity** -- the SDK must reach `https://app.tracekit.dev` to fetch breakpoint configurations and upload snapshots.

### Too many snapshots

- **Add a condition** to narrow when the breakpoint fires (e.g., `user.id == "specific-user"`).
- **Lower the rate limit** (e.g., from 10/min to 2/min).
- **Set an expiration** so the breakpoint auto-removes after a set time.

### Missing variables in snapshots

- **Increase capture depth** if nested object values show as `[truncated]`. Default is 3 levels; increase to 5 or higher.
- **Check redaction patterns** -- variables matching redaction patterns are replaced with `[REDACTED]`.
- **Check compiler optimizations** -- in Go, variables optimized away by the compiler may not be capturable. Build with `-gcflags='-N -l'` for debugging.

### High latency after enabling breakpoints

- **Reduce capture depth** to minimize serialization overhead.
- **Add conditions** to reduce how often breakpoints fire.
- **Avoid hot loops** -- do not set unconditional breakpoints inside tight loops or high-frequency functions.
- **Use logpoints** instead of full snapshots for lightweight tracing.

## Next Steps

Once code monitoring is working, consider:
- **Distributed Tracing** (`tracekit-distributed-tracing` skill) -- See breakpoint snapshots linked to full request traces across services
- **Alerts** (`tracekit-alerts` skill) -- Set up alerts when breakpoints fire frequently, indicating a recurring issue

## References

- Code monitoring docs: `https://app.tracekit.dev/docs/code-monitoring`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
