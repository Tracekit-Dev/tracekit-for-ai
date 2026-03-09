---
name: tracekit-laravel-sdk
description: Sets up TraceKit APM in Laravel applications with service providers, config publishing, and facades for automatic distributed tracing, error capture, and code monitoring. Use when the user asks to add TraceKit, add observability, or configure APM in a Laravel project.
---

# TraceKit Laravel SDK Setup
## When To Use

Use this skill when the user asks to:
- Add TraceKit to a Laravel application
- Add observability or APM to a Laravel project
- Instrument a Laravel service with distributed tracing
- Configure TraceKit API keys in a Laravel project
- Debug production Laravel services with live breakpoints
- Set up code monitoring in a Laravel app
- Trace Laravel queue jobs, database queries, or HTTP requests

**Important:** This skill is for Laravel projects only. If the project is vanilla PHP (no `laravel/framework` in `composer.json`), use the `tracekit-php-sdk` skill instead.

## Non-Negotiable Rules

1. **Never hardcode API keys** in code or config files. Always use `env('TRACEKIT_API_KEY')`.
2. **Always publish the config file** so the user has a clear place to customize settings.
3. **Always include a verification step** confirming traces appear in `https://app.tracekit.dev/traces`.
4. **Always enable code monitoring** (`enable_code_monitoring => true`) -- it is TraceKit's differentiator.
5. **Use `.env` for all secrets** -- Laravel's built-in env support handles this cleanly.

## Detection

Before applying this skill, detect the project type:

1. **Check for `composer.json`** -- confirms this is a PHP project.
2. **Check for Laravel** -- scan `composer.json` for `laravel/framework` in `require` or `require-dev`.
3. **Confirm Laravel markers** -- check for `artisan` file and `config/app.php`.
4. **Detect Laravel version** -- check `composer.lock` for `laravel/framework` version (10.x, 11.x, or 12.x supported).
5. **Only ask the user** if `composer.json` is missing or framework detection is ambiguous.

## Step 1: Environment Setup

Add the `TRACEKIT_API_KEY` to your `.env` file:

```bash
TRACEKIT_API_KEY=ctxio_your_api_key_here
```

The OTLP endpoint is configured in the published config file -- no need to set it as an env var.

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. Laravel's `.env` file is already in `.gitignore` by default.

## Step 2: Install SDK

```bash
composer require tracekit/laravel-apm
```

This installs the TraceKit Laravel package with:
- Auto-discovered service provider (no manual registration needed)
- Publishable config file
- HTTP middleware for request tracing
- Database query listener for Eloquent/query builder tracing
- Queue job tracing
- Exception handler integration
- Facade for manual instrumentation

**Prerequisites:**
- PHP 8.1 or higher
- Laravel 10.x, 11.x, or 12.x
- Composer package manager
- A TraceKit account ([create one free](https://app.tracekit.dev/register))

## Step 3: Publish Configuration

Publish the TraceKit config file:

```bash
php artisan vendor:publish --tag=tracekit-config
```

This creates `config/tracekit.php`:

```php
<?php

return [
    /*
    |--------------------------------------------------------------------------
    | TraceKit API Key
    |--------------------------------------------------------------------------
    |
    | Your TraceKit API key for authentication. Get one at:
    | https://app.tracekit.dev/api-keys
    |
    */
    'api_key' => env('TRACEKIT_API_KEY'),

    /*
    |--------------------------------------------------------------------------
    | Service Name
    |--------------------------------------------------------------------------
    |
    | A unique name identifying this service in the TraceKit dashboard.
    | Defaults to your APP_NAME from .env.
    |
    */
    'service_name' => env('TRACEKIT_SERVICE_NAME', env('APP_NAME', 'laravel')),

    /*
    |--------------------------------------------------------------------------
    | TraceKit Endpoint
    |--------------------------------------------------------------------------
    |
    | The OTLP endpoint where traces are sent.
    |
    */
    'endpoint' => 'https://app.tracekit.dev/v1/traces',

    /*
    |--------------------------------------------------------------------------
    | Code Monitoring
    |--------------------------------------------------------------------------
    |
    | Enable live breakpoints and snapshots for production debugging.
    | This is TraceKit's key differentiator -- leave enabled.
    |
    */
    'enable_code_monitoring' => true,

    /*
    |--------------------------------------------------------------------------
    | Auto-Tracing Options
    |--------------------------------------------------------------------------
    |
    | Control which components are automatically traced.
    |
    */
    'tracing' => [
        'requests'  => true,  // HTTP request/response tracing
        'database'  => true,  // Eloquent and query builder tracing
        'queue'     => true,  // Queue job tracing
        'http'      => true,  // Outgoing HTTP client tracing
        'cache'     => true,  // Cache operation tracing
    ],
];
```

## Step 4: Register Middleware

The TraceKit service provider auto-registers global middleware for request tracing. However, if you need to add it manually or apply it to specific route groups:

**Laravel 11.x / 12.x** (`bootstrap/app.php`):

```php
->withMiddleware(function (Middleware $middleware) {
    $middleware->prepend(\Tracekit\Laravel\Http\Middleware\TraceRequests::class);
})
```

**Laravel 10.x** (`app/Http/Kernel.php`):

```php
protected $middleware = [
    \Tracekit\Laravel\Http\Middleware\TraceRequests::class,
    // ... other middleware
];
```

The middleware automatically captures:
- HTTP method, route, and status code
- Request duration
- Client IP address
- Route parameters (sanitized)

## Step 5: Exception Handler Integration

Integrate TraceKit with Laravel's exception handler to automatically capture all unhandled exceptions.

**Laravel 11.x / 12.x** (`bootstrap/app.php`):

```php
->withExceptions(function (Exceptions $exceptions) {
    $exceptions->reportable(function (\Throwable $e) {
        \Tracekit\Laravel\Facades\Tracekit::captureException($e);
    });
})
```

**Laravel 10.x** (`app/Exceptions/Handler.php`):

```php
use Tracekit\Laravel\Facades\Tracekit;

public function register(): void
{
    $this->reportable(function (\Throwable $e) {
        Tracekit::captureException($e);
    });
}
```

## Step 6: Using the Facade

The `Tracekit` facade provides convenient access for manual instrumentation:

```php
use Tracekit\Laravel\Facades\Tracekit;

// Capture an exception
try {
    $result = riskyOperation();
} catch (\Exception $e) {
    Tracekit::captureException($e);
    throw $e;
}

// Create a custom span
$span = Tracekit::startSpan('process-order', [
    'order.id' => $orderId,
    'user.id'  => auth()->id(),
]);

try {
    $order = $this->processOrder($orderId);
} finally {
    Tracekit::finishSpan($span);
}

// Add attributes to the current span
Tracekit::setSpanAttribute('payment.method', 'stripe');
Tracekit::setSpanAttribute('order.total', $order->total);
```

## Step 6b: Snapshot Capture (Code Monitoring)

For programmatic snapshots, **use the SnapshotClient directly** — do not call through the SDK wrapper or facade. The SDK uses stack inspection internally to identify the call site. Adding extra layers shifts the frame and causes snapshots to report the wrong source location.

Create a `Breakpoints` helper (e.g., `app/Support/Breakpoints.php`):

```php
<?php

namespace App\Support;

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

Initialize in a service provider or `AppServiceProvider::boot()`:

```php
use App\Support\Breakpoints;
use Tracekit\Laravel\Facades\Tracekit;

public function boot(): void
{
    Breakpoints::init(Tracekit::sdk());
}
```

Use at call sites:

```php
use App\Support\Breakpoints;

Breakpoints::capture('payment-failed', ['order_id' => $orderId, 'error' => $e->getMessage()]);
```

See the `tracekit-code-monitoring` skill for the full pattern across all languages.

## Step 7: Database Query Tracing

Database queries are automatically traced when `tracing.database` is `true` in `config/tracekit.php`. Each query generates a span with:

- SQL statement (parameterized -- no sensitive data)
- Connection name
- Query duration
- Binding count

No additional setup is required. Eloquent models, query builder, and raw queries are all traced:

```php
// All of these are automatically traced
$users = User::where('active', true)->get();
$count = DB::table('orders')->count();
$results = DB::select('SELECT * FROM products WHERE price > ?', [100]);
```

## Step 8: Queue Job Tracing

Queue jobs are automatically traced when `tracing.queue` is `true`. Each job generates a span with:

- Job class name
- Queue name
- Attempt number
- Job duration

```php
// This job will be automatically traced
ProcessOrderJob::dispatch($order);
```

Trace context is propagated from the dispatching request to the queue worker, enabling end-to-end distributed tracing across sync and async boundaries.

## Step 9: Outgoing HTTP Client Tracing

Laravel's HTTP client calls are automatically traced when `tracing.http` is `true`:

```php
use Illuminate\Support\Facades\Http;

// Automatically traced with method, URL, status code, and duration
$response = Http::get('https://api.example.com/users');

$response = Http::post('https://api.example.com/orders', [
    'item' => 'widget',
    'quantity' => 5,
]);
```

Trace context headers are automatically injected into outgoing requests for distributed tracing across services.

## Step 10: Verification

After integrating, verify traces are flowing:

1. **Start your application** -- `php artisan serve` (ensure `.env` has `TRACEKIT_API_KEY`).
2. **Hit your endpoints 3-5 times** -- e.g., `curl http://localhost:8000/api/users`.
3. **Open** `https://app.tracekit.dev/traces`.
4. **Confirm** new spans and your service name appear within 30-60 seconds.

If traces do not appear, see Troubleshooting below.

## Troubleshooting

### Traces not appearing in dashboard

- **Check `.env`:** Ensure `TRACEKIT_API_KEY=ctxio_...` is set in `.env` (not `.env.example`).
- **Clear config cache:** Run `php artisan config:clear` after changing `.env`.
- **Check outbound access:** Your service must reach `https://app.tracekit.dev/v1/traces`. Verify with: `curl -X POST https://app.tracekit.dev/v1/traces` (expect 401 -- means the endpoint is reachable).

### Service provider not loaded

Symptoms: `Class 'Tracekit\Laravel\Facades\Tracekit' not found` error.

Fix: Ensure the package is installed (`composer show tracekit/laravel-apm`). If auto-discovery is disabled, manually register in `config/app.php`:

```php
'providers' => [
    // ...
    Tracekit\Laravel\TracekitServiceProvider::class,
],

'aliases' => [
    // ...
    'Tracekit' => Tracekit\Laravel\Facades\Tracekit::class,
],
```

### Config not published

Symptoms: `config('tracekit.api_key')` returns `null`.

Fix: Run `php artisan vendor:publish --tag=tracekit-config` and verify `config/tracekit.php` exists.

### Queue jobs not traced

Symptoms: HTTP requests show traces but queue jobs do not.

Fix: Ensure the queue worker is started **after** the TraceKit service provider boots. Restart workers after installing the package: `php artisan queue:restart`.

### Service name collisions

Symptoms: Traces appear under the wrong service in the dashboard.

Fix: Set a unique `TRACEKIT_SERVICE_NAME` in `.env` for each deployed service. Avoid using the default `APP_NAME` if multiple Laravel apps share the same name.

## Next Steps

Once your Laravel application is traced, consider:
- **Code Monitoring** -- Set live breakpoints and capture snapshots in production without redeploying (already enabled via `enable_code_monitoring => true`)
- **Distributed Tracing** -- Connect traces across multiple services for full request visibility
- **Frontend Observability** -- Add `@tracekit/browser` to your frontend for end-to-end trace correlation

## References

- Laravel SDK docs: `https://app.tracekit.dev/docs/languages/laravel`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
