---
name: tracekit-php-sdk
description: Sets up TraceKit APM in vanilla PHP applications for automatic distributed tracing, error capture, and code monitoring. Use when the user asks to add TraceKit, add observability, instrument a PHP app, or configure APM in a PHP project that does not use Laravel.
---

# TraceKit PHP SDK Setup

> **Coming soon -- SDK in development.** The patterns below reflect the planned API. Package names and method signatures may change before GA release.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to a PHP application
- Add observability or APM to a vanilla PHP project
- Instrument a PHP service with distributed tracing
- Configure TraceKit API keys in a PHP project
- Debug production PHP services with live breakpoints
- Set up code monitoring in a PHP app

**Important:** If the project uses Laravel (check for `laravel/framework` in `composer.json`), use the `tracekit-laravel-sdk` skill instead -- it provides Laravel-specific service providers, config publishing, and facades.

## Non-Negotiable Rules

1. **Never hardcode API keys** in code. Always use `getenv('TRACEKIT_API_KEY')`.
2. **Always initialize TraceKit before handling requests** -- the `Tracekit\init()` call must happen at the top of your entry point.
3. **Always include a verification step** confirming traces appear in `https://app.tracekit.dev/traces`.
4. **Always enable code monitoring** (`enable_code_monitoring => true`) -- it is TraceKit's differentiator.
5. **Use env vars for all secrets** -- `.env` files, CI secrets, or deployment secret managers.

## Detection

Before applying this skill, detect the project type:

1. **Check for `composer.json`** -- confirms this is a PHP project.
2. **Confirm no Laravel** -- scan `composer.json` for `laravel/framework`. If found, use the `tracekit-laravel-sdk` skill.
3. **Check PHP version** -- requires PHP 8.1 or higher.
4. **Only ask the user** if `composer.json` is missing or if the framework cannot be determined.

## Step 1: Environment Setup

Set the `TRACEKIT_API_KEY` environment variable. This is the only required secret.

Add to your `.env` file or environment:

```bash
export TRACEKIT_API_KEY=ctxio_your_api_key_here
```

The OTLP endpoint is hardcoded in the SDK init -- no need to configure it separately.

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. Use `.env` files, deployment secret managers, or CI variables.

## Step 2: Install SDK

```bash
composer require tracekit/php-apm
```

This installs the TraceKit PHP SDK with built-in OpenTelemetry support, HTTP middleware, database query tracing, and code monitoring.

**Prerequisites:**
- PHP 8.1 or higher
- Composer package manager
- A TraceKit account ([create one free](https://app.tracekit.dev/register))

## Step 3: Initialize TraceKit

Add this to the **top** of your entry point (e.g., `index.php` or `public/index.php`), before any request handling:

```php
<?php

require_once __DIR__ . '/vendor/autoload.php';

// Initialize TraceKit -- MUST be before request handling
Tracekit\init([
    'api_key'                => getenv('TRACEKIT_API_KEY'),
    'service_name'           => 'my-php-service',
    'endpoint'               => 'https://app.tracekit.dev/v1/traces',
    'enable_code_monitoring' => true,
]);

// ... your application code below
```

**Key points:**
- `service_name` should match your service's logical name (e.g., `"api-gateway"`, `"user-service"`)
- `enable_code_monitoring => true` enables live breakpoints and snapshots in production
- The `init()` call must happen before any route or request handling

## Step 4: HTTP Request Tracing

Wrap your request handler with TraceKit middleware to automatically trace all HTTP requests:

```php
<?php

require_once __DIR__ . '/vendor/autoload.php';

Tracekit\init([
    'api_key'                => getenv('TRACEKIT_API_KEY'),
    'service_name'           => 'my-php-service',
    'endpoint'               => 'https://app.tracekit.dev/v1/traces',
    'enable_code_monitoring' => true,
]);

// Option A: Use TracekitMiddleware to wrap your request handler
$handler = new Tracekit\TracekitMiddleware(function ($request) {
    // Your request handling logic
    $path = $_SERVER['REQUEST_URI'];

    if ($path === '/api/users') {
        header('Content-Type: application/json');
        echo json_encode(['users' => ['alice', 'bob']]);
        return;
    }

    http_response_code(404);
    echo json_encode(['error' => 'Not found']);
});

$handler->handle();
```

If you prefer manual span management:

```php
// Option B: Manual span creation
$span = Tracekit\startSpan('handle-request', [
    'http.method' => $_SERVER['REQUEST_METHOD'],
    'http.url'    => $_SERVER['REQUEST_URI'],
]);

try {
    // Your request handling logic
    processRequest();
} finally {
    Tracekit\finishSpan($span);
}
```

## Step 5: Database Query Tracing

TraceKit automatically traces PDO queries when using the provided wrapper:

```php
use Tracekit\TracekitPDO;

// Wrap your PDO connection with TraceKit
$db = new TracekitPDO(
    'mysql:host=localhost;dbname=myapp',
    'username',
    'password'
);

// Queries are automatically traced
$stmt = $db->prepare('SELECT * FROM users WHERE id = ?');
$stmt->execute([$userId]);
$user = $stmt->fetch();
```

For existing PDO connections, wrap them:

```php
$existingPdo = new PDO('mysql:host=localhost;dbname=myapp', 'user', 'pass');
$tracedPdo = Tracekit\wrapPDO($existingPdo);
```

## Step 6: Error Capture

Capture exceptions explicitly where you handle them:

```php
try {
    $result = someRiskyOperation();
} catch (\Exception $e) {
    Tracekit\captureException($e);
    // Handle the error...
}
```

Set up a global exception handler to catch unhandled exceptions:

```php
set_exception_handler(function (\Throwable $e) {
    Tracekit\captureException($e);
    http_response_code(500);
    echo json_encode(['error' => 'Internal server error']);
});
```

For adding context to traces, use manual spans:

```php
$span = Tracekit\startSpan('process-order', [
    'order.id' => $orderId,
    'user.id'  => $userId,
]);

try {
    // Your business logic
    $order = processOrder($orderId);
} catch (\Exception $e) {
    Tracekit\captureException($e);
    throw $e;
} finally {
    Tracekit\finishSpan($span);
}
```

## Step 6b: Snapshot Capture (Code Monitoring)

For programmatic snapshots, **use the SnapshotClient directly** — do not call through the SDK wrapper. The SDK uses stack inspection internally to identify the call site. Adding extra layers shifts the frame and causes snapshots to report the wrong source location.

Create a `Breakpoints` helper (e.g., `src/Breakpoints.php`):

```php
<?php

namespace App;

class Breakpoints
{
    private static $snapshotClient = null;

    public static function init($sdk): void
    {
        if ($sdk !== null) {
            self::$snapshotClient = $sdk->snapshotClient();
        }
    }

    public static function capture(string $name, array $data): void
    {
        if (self::$snapshotClient === null) {
            return;
        }
        self::$snapshotClient->checkAndCapture($name, $data);
    }
}
```

Initialize after SDK setup:

```php
\App\Breakpoints::init($sdk);
```

Use at call sites:

```php
\App\Breakpoints::capture('payment-failed', ['order_id' => $orderId, 'error' => $e->getMessage()]);
```

See the `tracekit-code-monitoring` skill for the full pattern across all languages.

## Step 7: External HTTP Call Tracing

Trace outgoing HTTP requests with the TraceKit HTTP client wrapper:

```php
use Tracekit\TracekitHttpClient;

$client = new TracekitHttpClient();

// GET request -- automatically traced
$response = $client->get('https://api.example.com/users');

// POST request -- automatically traced
$response = $client->post('https://api.example.com/orders', [
    'json' => ['item' => 'widget', 'quantity' => 5],
]);
```

If using cURL directly, wrap calls manually:

```php
$span = Tracekit\startSpan('http-client', [
    'http.method' => 'GET',
    'http.url'    => 'https://api.example.com/data',
]);

$ch = curl_init('https://api.example.com/data');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

$span->setAttribute('http.status_code', $httpCode);
Tracekit\finishSpan($span);
```

## Step 8: Verification

After integrating, verify traces are flowing:

1. **Start your application** with `TRACEKIT_API_KEY` set in the environment.
2. **Hit your endpoints 3-5 times** -- e.g., `curl http://localhost:8080/api/users`.
3. **Open** `https://app.tracekit.dev/traces`.
4. **Confirm** new spans and your service name appear within 30-60 seconds.

If traces do not appear, see Troubleshooting below.

## Troubleshooting

### Traces not appearing in dashboard

- **Check `TRACEKIT_API_KEY`:** Ensure the env var is set in the runtime environment (not just in your shell). Verify: `echo getenv('TRACEKIT_API_KEY');`.
- **Check outbound access:** Your service must reach `https://app.tracekit.dev/v1/traces`. Verify with: `curl -X POST https://app.tracekit.dev/v1/traces` (expect 401 -- means the endpoint is reachable).
- **Check init order:** `Tracekit\init()` must be called **before** any request handling. If init happens after your router runs, requests are not traced.

### Init order wrong

Symptoms: Server starts fine but no traces appear despite traffic.

Fix: Move `Tracekit\init()` to the very top of your entry point (`index.php`), before `require_once` of your router or framework bootstrap.

### Missing environment variable

Symptoms: `Tracekit init failed` error on startup, or traces appear without an API key (rejected by backend).

Fix: Ensure `TRACEKIT_API_KEY` is exported in your shell, `.env` file, Docker Compose, or deployment config. For PHP-FPM, set it in the pool config (`env[TRACEKIT_API_KEY]`).

### Service name collisions

Symptoms: Traces appear under the wrong service in the dashboard.

Fix: Use a unique `service_name` per deployed service. Avoid generic names like `"app"` or `"server"`.

### PHP-FPM environment variables

Symptoms: `getenv('TRACEKIT_API_KEY')` returns `false` in PHP-FPM.

Fix: Add to your PHP-FPM pool config (`/etc/php/8.x/fpm/pool.d/www.conf`):

```ini
env[TRACEKIT_API_KEY] = ctxio_your_api_key_here
```

Or use a `.env` loader like `vlucas/phpdotenv`:

```php
$dotenv = Dotenv\Dotenv::createImmutable(__DIR__);
$dotenv->load();
```

## Next Steps

Once your PHP application is traced, consider:
- **Code Monitoring** -- Set live breakpoints and capture snapshots in production without redeploying (already enabled via `enable_code_monitoring => true`)
- **Distributed Tracing** -- Connect traces across multiple services for full request visibility
- **Frontend Observability** -- Add `@tracekit/browser` to your frontend for end-to-end trace correlation

## References

- PHP SDK docs: `https://app.tracekit.dev/docs/languages/php`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
