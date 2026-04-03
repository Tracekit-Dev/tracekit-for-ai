---
name: tracekit-code-monitoring
description: Enable live breakpoints and snapshots for production debugging with TraceKit Code Monitoring. Works with all backend SDKs. Use when the user asks to debug production code, set breakpoints, capture variable state, or enable code monitoring.
---

# TraceKit Code Monitoring

## Auth Bootstrap

Do not tell the user to go sign up, log in separately, or manually create an API key before setup. First check for existing TraceKit auth. If `~/.tracekitconfig` does not contain the production profile or `TRACEKIT_API_KEY` is missing, apply the `tracekit-auth` skill first. Use `./scripts/run-tracekit-auth.sh status` to check, then guide the user through the TraceKit email verification flow with `register` and `verify`. That flow signs the user into an existing account for that email or creates the account automatically, then saves the returned credentials for the rest of the setup.

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

**Production-grade by default:** All SDKs include PII scrubbing (13 typed patterns, default-on), crash isolation, circuit breakers (auto-disable after 3 failures, auto-recover after 5 minutes), remote kill switch, and real-time breakpoint sync via Server-Sent Events (SSE).

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
| `package.json` | `@tracekit/node-apm` | Node.js SDK |
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
const tracekit = require('@tracekit/node-apm');

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

The breakpoint takes effect within seconds -- no redeployment needed. SDKs receive breakpoint updates in real-time via SSE (Server-Sent Events), with automatic fallback to polling if SSE is unavailable.

## Step 3: Breakpoint Types

### Line Breakpoint

Captures all local variables when execution hits the specified line.

**Use for:** Investigating a specific code path, seeing what state looks like at a particular point in your logic.

**Example:** Set a line breakpoint on the line where you process a payment to see the full payment object and user context.

### Conditional Breakpoint

Only triggers when a boolean expression evaluates to true. Conditions are **evaluated server-side** in a sandboxed expression engine -- raw expressions never run in your application. SDKs send metadata via a check-in endpoint, and the server decides whether to instruct capture.

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

### PII Scrubbing (Default On)

All SDKs automatically scrub sensitive data from snapshots before transmission. Detected values are replaced with typed markers like `[REDACTED:email]` for audit trails.

**13 built-in patterns** (enabled by default):

| Pattern | Detects | Marker |
|---|---|---|
| Email | `user@example.com` | `[REDACTED:email]` |
| SSN | `123-45-6789` | `[REDACTED:ssn]` |
| Credit Card | `4111111111111111` | `[REDACTED:credit_card]` |
| Phone | `+1-555-123-4567` | `[REDACTED:phone]` |
| AWS Access Key | `AKIA...` | `[REDACTED:aws_key]` |
| AWS Secret Key | `wJalrXUtnFE...` | `[REDACTED:aws_secret]` |
| OAuth Token | `ya29.a0...` | `[REDACTED:oauth_token]` |
| Stripe Key | `sk_live_...` | `[REDACTED:stripe_key]` |
| Sensitive Variable Name | `password`, `secret`, `api_key`, etc. | `[REDACTED:sensitive_name]` |
| JWT | `eyJhbGciOi...` | `[REDACTED:jwt]` |
| Private Key | `-----BEGIN...` | `[REDACTED:private_key]` |
| API Key (generic) | Various key patterns | `[REDACTED:api_key]` |
| Password value | In password-named fields | `[REDACTED:password]` |

Add custom patterns in the SDK config:

```javascript
// Node.js example
tracekit.init({
  apiKey: process.env.TRACEKIT_API_KEY,
  serviceName: 'my-service',
  enableCodeMonitoring: true,
  captureConfig: {
    piiScrubbing: true,  // default: true
    piiPatterns: [
      { pattern: /CUSTOM-\d+/, marker: '[REDACTED:custom_id]' },
    ],
  },
});
```

To disable PII scrubbing (not recommended for production):

```javascript
captureConfig: { piiScrubbing: false }
```

## Step 5: Programmatic Snapshots (SDK API)

Capture snapshots programmatically from your code to debug specific code paths. This is the recommended approach for adding observability to critical business logic, error handlers, and conditional paths.

**Important: Use the SnapshotClient directly.** All backend SDKs expose a `SnapshotClient` (or equivalent) that you should use instead of calling through the main SDK wrapper. This preserves correct stack frame attribution -- the SDK uses `runtime.Caller(2)` (or equivalent) internally to identify the call site. Adding extra layers between your code and the SnapshotClient causes the stack frame to resolve to the wrapper instead of the actual call site.

**Recommended pattern:** Create a thin wrapper package that holds the SnapshotClient and provides a simple `Capture()` function. This keeps the call chain short:

```
call site → your Capture() → SnapshotClient.CheckAndCaptureWithContext()
```

### Go

Create a `breakpoints` package (e.g., `internal/breakpoints/breakpoints.go`):

```go
package breakpoints

import (
    "context"

    "github.com/Tracekit-Dev/go-sdk/tracekit"
)

// snapshotClient holds the TraceKit snapshot client directly.
// We bypass SDK.CheckAndCaptureWithContext to keep the correct
// runtime.Caller frame count — the SDK uses runtime.Caller(2)
// internally, so the chain must be:
//
//    call site → Capture → SnapshotClient.CheckAndCaptureWithContext
//
// If we went through the SDK wrapper, Caller(2) would resolve to
// this file instead of the actual call site.
var snapshotClient *tracekit.SnapshotClient

// Init stores the TraceKit snapshot client for code monitoring.
// When sdk is nil (tracing disabled), Capture is a no-op.
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

Initialize in `main()` after SDK setup:

```go
import "myapp/internal/breakpoints"

func main() {
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

    // Initialize the snapshot wrapper
    breakpoints.Init(sdk)

    // ... register routes
}
```

Use at call sites:

```go
func handlePayment(ctx context.Context, order Order) error {
    result, err := processPayment(order)
    if err != nil {
        breakpoints.Capture(ctx, "payment-failed", map[string]interface{}{
            "order_id": order.ID,
            "error":    err.Error(),
            "amount":   order.Total,
        })
        return err
    }

    breakpoints.Capture(ctx, "payment-success", map[string]interface{}{
        "order_id":       order.ID,
        "transaction_id": result.TransactionID,
    })
    return nil
}
```

### Node.js

Create a `breakpoints` module (e.g., `src/lib/breakpoints.ts`):

```typescript
import * as tracekit from '@tracekit/node-apm-apm';

// Hold the snapshot client directly to preserve correct stack frame attribution.
// The SDK uses stack inspection internally — calling through the SDK wrapper
// adds an extra frame that shifts the reported call site.
let snapshotClient: tracekit.SnapshotClient | null = null;

export function init(sdk: tracekit.SDK): void {
  snapshotClient = sdk.snapshotClient();
}

export function capture(name: string, data: Record<string, unknown>): void {
  if (!snapshotClient) return;
  snapshotClient.checkAndCapture(name, data);
}
```

Initialize after SDK setup:

```typescript
import * as tracekit from '@tracekit/node-apm-apm';
import * as breakpoints from './lib/breakpoints';

const sdk = tracekit.init({
  apiKey: process.env.TRACEKIT_API_KEY!,
  serviceName: 'my-node-service',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
});

breakpoints.init(sdk);
```

Use at call sites:

```typescript
import { capture } from './lib/breakpoints';

async function handlePayment(order: Order) {
  try {
    const result = await processPayment(order);
    capture('payment-success', { orderId: order.id, transactionId: result.id });
  } catch (err) {
    capture('payment-failed', { orderId: order.id, error: String(err) });
    throw err;
  }
}
```

### Python

Create a `breakpoints` module (e.g., `app/breakpoints.py`):

```python
import tracekit

# Hold the snapshot client directly to preserve correct stack frame attribution.
# The SDK uses stack inspection internally — calling through the SDK wrapper
# adds an extra frame that shifts the reported call site.
_snapshot_client = None


def init(sdk):
    """Store the snapshot client. No-op when sdk is None (tracing disabled)."""
    global _snapshot_client
    if sdk is not None:
        _snapshot_client = sdk.snapshot_client()


def capture(name: str, data: dict):
    """Fire a code monitoring snapshot. No-op when tracing is disabled."""
    if _snapshot_client is None:
        return
    _snapshot_client.check_and_capture(name, data)
```

Initialize after SDK setup:

```python
import tracekit
from app.breakpoints import init as init_breakpoints

sdk = tracekit.init(
    api_key=os.getenv("TRACEKIT_API_KEY"),
    service_name="my-python-service",
    endpoint="https://app.tracekit.dev/v1/traces",
    enable_code_monitoring=True,
)

init_breakpoints(sdk)
```

Use at call sites:

```python
from app.breakpoints import capture

def handle_payment(order):
    try:
        result = process_payment(order)
        capture("payment-success", {"order_id": order.id, "transaction_id": result.id})
    except Exception as e:
        capture("payment-failed", {"order_id": order.id, "error": str(e)})
        raise
```

### PHP

Create a `Breakpoints` helper (e.g., `src/Breakpoints.php`):

```php
<?php

namespace App;

class Breakpoints
{
    private static $snapshotClient = null;

    /**
     * Store the snapshot client directly to preserve correct stack frame attribution.
     */
    public static function init($sdk): void
    {
        if ($sdk !== null) {
            self::$snapshotClient = $sdk->snapshotClient();
        }
    }

    /**
     * Fire a code monitoring snapshot. No-op when tracing is disabled.
     */
    public static function capture(string $name, array $data): void
    {
        if (self::$snapshotClient === null) {
            return;
        }
        self::$snapshotClient->checkAndCapture($name, $data);
    }
}
```

### Java

Create a `Breakpoints` utility class (e.g., `src/main/java/com/myapp/Breakpoints.java`):

```java
import dev.tracekit.SnapshotClient;
import dev.tracekit.Tracekit;

/**
 * Thin wrapper around SnapshotClient to preserve correct stack frame attribution.
 * The SDK uses stack inspection internally — calling through the SDK wrapper
 * adds an extra frame that shifts the reported call site.
 */
public final class Breakpoints {
    private static SnapshotClient snapshotClient;

    public static void init(Tracekit sdk) {
        if (sdk != null) {
            snapshotClient = sdk.snapshotClient();
        }
    }

    public static void capture(String name, Map<String, Object> data) {
        if (snapshotClient == null) return;
        snapshotClient.checkAndCapture(name, data);
    }
}
```

### .NET

Create a `Breakpoints` static helper (e.g., `Breakpoints.cs`):

```csharp
using TraceKit;

/// <summary>
/// Thin wrapper around SnapshotClient to preserve correct stack frame attribution.
/// The SDK uses stack inspection internally — calling through the SDK wrapper
/// adds an extra frame that shifts the reported call site.
/// </summary>
public static class Breakpoints
{
    private static ISnapshotClient? _snapshotClient;

    public static void Init(TracekitSdk sdk)
    {
        _snapshotClient = sdk.SnapshotClient();
    }

    public static void Capture(string name, Dictionary<string, object> data)
    {
        _snapshotClient?.CheckAndCapture(name, data);
    }
}
```

### Ruby

Create a `Breakpoints` module (e.g., `lib/breakpoints.rb`):

```ruby
module Breakpoints
  # Hold the snapshot client directly to preserve correct stack frame attribution.
  @snapshot_client = nil

  def self.init(sdk)
    @snapshot_client = sdk&.snapshot_client
  end

  def self.capture(name, data = {})
    return unless @snapshot_client
    @snapshot_client.check_and_capture(name, data)
  end
end
```

### Dashboard Breakpoints

You can also set breakpoints from the dashboard without any code changes. See Step 2 above for the dashboard workflow. Programmatic snapshots captured via the wrapper pattern above appear in the dashboard alongside dashboard-created breakpoints and can be filtered, compared, and linked to traces.

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
3. **Within a few seconds** (SSE) or **up to 30 seconds** (polling), check `https://app.tracekit.dev/code-monitoring` for a new snapshot
4. **Verify** the captured variables match expected state (request headers, local variables, etc.)
5. **Click the Trace ID** to confirm the snapshot is linked to a distributed trace
6. **Remove or deactivate** the test breakpoint once verified

If snapshots do not appear, see Troubleshooting below.

## Troubleshooting

### Snapshots not appearing

- **Check `enableCodeMonitoring: true`** is set in your SDK init config. Without this, the SDK does not poll for breakpoints.
- **Check the breakpoint is active** in the dashboard. Expired or paused breakpoints do not fire.
- **Check the kill switch** is not active for the service (Code Monitoring dashboard > service > Kill Switch toggle).
- **Check the circuit breaker** hasn't tripped -- look for "circuit breaker open" in SDK logs. It auto-recovers after 5 minutes.
- **Check the service is sending data** -- visit `https://app.tracekit.dev/traces` to confirm traces are flowing from this service.
- **Check the service name matches** -- the breakpoint targets a specific service. Ensure `serviceName` in your SDK init matches the service selected when creating the breakpoint.
- **Check network connectivity** -- the SDK must reach `https://app.tracekit.dev` to fetch breakpoint configurations and upload snapshots.
- **Check condition expressions** -- if a breakpoint has a condition, verify it's syntactically valid and the metadata matches.

### Too many snapshots

- **Add a condition** to narrow when the breakpoint fires (e.g., `user.id == "specific-user"`).
- **Lower the rate limit** (e.g., from 10/min to 2/min).
- **Set an expiration** so the breakpoint auto-removes after a set time.

### Missing variables in snapshots

- **Increase capture depth** if nested object values show as `[truncated]`. Use `captureConfig.captureDepth` to increase (default: unlimited, but breakpoint settings may limit).
- **Check PII scrubbing** -- variables matching the 13 built-in PII patterns are replaced with typed markers like `[REDACTED:email]`. This is expected behavior for sensitive data.
- **Check payload limits** -- if using `captureConfig.maxPayload`, large snapshots are truncated with a `_truncated` marker.
- **Check compiler optimizations** -- in Go, variables optimized away by the compiler may not be capturable. Build with `-gcflags='-N -l'` for debugging.

### High latency after enabling breakpoints

- **Reduce capture depth** to minimize serialization overhead.
- **Add conditions** to reduce how often breakpoints fire.
- **Avoid hot loops** -- do not set unconditional breakpoints inside tight loops or high-frequency functions.
- **Use logpoints** instead of full snapshots for lightweight tracing.

## Production Safety Features

All safety features are enabled by default with zero configuration:

| Feature | Default | Override |
|---|---|---|
| PII Scrubbing (13 patterns) | Enabled | `captureConfig.piiScrubbing: false` to disable |
| Crash Isolation | Always on | Cannot be disabled |
| Circuit Breaker (3 failures/60s, 5min cooldown) | Enabled | `captureConfig.circuitBreaker: { maxFailures, windowMs, cooldownMs }` |
| Remote Kill Switch | Available | Toggle in dashboard per service |
| Real-Time SSE | Auto-discovered | No config needed, falls back to polling |
| Capture Limits | Disabled (unlimited) | `captureConfig: { captureDepth, maxPayload, captureTimeout }` |

### Kill Switch

To immediately stop all code monitoring for a service:
1. Go to **Code Monitoring** in the dashboard
2. Select the service
3. Toggle the **Kill Switch**

All connected SDKs stop capturing within 1 second (SSE) or 60 seconds (polling).

### Circuit Breaker

If the TraceKit backend becomes unreachable:
1. After 3 HTTP 5xx/network failures within 60 seconds, code monitoring auto-disables
2. After 5 minutes, it automatically re-enables and retries
3. Only HTTP 5xx and network errors count -- 4xx errors are ignored

## Next Steps

Once code monitoring is working, consider:
- **Distributed Tracing** (`tracekit-distributed-tracing` skill) -- See breakpoint snapshots linked to full request traces across services
- **Alerts** (`tracekit-alerts` skill) -- Set up alerts when breakpoints fire frequently, indicating a recurring issue

## References

- Code monitoring docs: `https://app.tracekit.dev/docs/code-monitoring`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
