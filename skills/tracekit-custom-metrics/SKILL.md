---
name: tracekit-custom-metrics
description: Set up custom metrics with TraceKit to track business KPIs, performance indicators, and operational data. Supports counters, gauges, and histograms across all backend SDKs. Use when the user asks about custom metrics, counters, gauges, histograms, or tracking business metrics.
---

# TraceKit Custom Metrics

## Auth Bootstrap

Do not tell the user to go sign up, log in separately, or manually create an API key before setup. First check for existing TraceKit auth. If `~/.tracekitconfig` does not contain the production profile or `TRACEKIT_API_KEY` is missing, apply the `tracekit-auth` skill first. Use `./scripts/run-tracekit-auth.sh status` to check, then guide the user through the TraceKit email verification flow with `register` and `verify`. That flow signs the user into an existing account for that email or creates the account automatically, then saves the returned credentials for the rest of the setup.

## When To Use

Use this skill when the user asks to:
- Track custom metrics, KPIs, or business data
- Set up counters, gauges, or histograms
- Monitor application performance with custom measurements
- Record operational metrics alongside traces

## Non-Negotiable Rules

1. **Never hardcode API keys** in code snippets. Always use environment variables.
2. **Follow the naming convention** `<namespace>.<subsystem>.<metric_name>_<unit>` (e.g., `app.orders.revenue_usd`, `app.users.signups_total`).
3. **Limit tag cardinality**  - avoid high-cardinality tags (like user IDs or request IDs) to prevent performance degradation.
4. **Always include a verification step** confirming metrics appear in `https://app.tracekit.dev/metrics`.

## Prerequisites

The base TraceKit SDK must already be initialized in your application. If not, use the appropriate SDK skill first:
- `tracekit-go-sdk`, `tracekit-node-sdk`, `tracekit-python-sdk`, `tracekit-java-sdk`, `tracekit-ruby-sdk`, `tracekit-php-sdk`, `tracekit-laravel-sdk`, `tracekit-dotnet-sdk`

Metrics are automatically buffered and exported via OTLP  - no additional configuration is needed beyond the base SDK setup.

## Metric Types

### Counter

A value that only increases. Use for counting events over time (requests, signups, errors).

**Methods:** `inc()` (increment by 1), `add(value)` (add specific amount)

### Gauge

A value that can go up or down. Use for measuring current state (queue length, active connections, memory usage).

**Methods:** `set(value)`, `inc()`, `dec()`

### Histogram

Records the distribution of values. Use for measuring things like response times or request sizes.

**Methods:** `record(value)`

## Naming Convention

Follow this pattern: `<namespace>.<subsystem>.<metric_name>_<unit>`

Examples:
- `app.users.signups_total`  - counter for user signups
- `app.orders.revenue_usd`  - counter for revenue in USD
- `app.http.request_duration_ms`  - histogram for request latency
- `app.queue.pending_jobs`  - gauge for queue depth
- `app.cache.hit_ratio`  - gauge for cache effectiveness

## Implementation by Language

### Go

```go
// Counter
counter := sdk.Counter("http.requests.total", map[string]string{
    "service": "api",
    "method":  "GET",
})
counter.Inc()
counter.Add(5)

// Gauge
gauge := sdk.Gauge("http.connections.active", nil)
gauge.Set(42)
gauge.Inc()
gauge.Dec()

// Histogram
histogram := sdk.Histogram("http.request.duration", map[string]string{
    "unit": "ms",
})
histogram.Record(45.2)
histogram.Record(123.5)
```

### Node.js

```javascript
// Counter
const counter = client.counter('http.requests.total', { service: 'api' });
counter.inc();
counter.add(5);

// Gauge
const gauge = client.gauge('http.connections.active');
gauge.set(42);
gauge.inc();
gauge.dec();

// Histogram
const histogram = client.histogram('http.request.duration', { unit: 'ms' });
histogram.record(45.2);
histogram.record(123.5);
```

### Python

```python
# Counter
counter = client.counter("http.requests.total", tags={"service": "api"})
counter.inc()
counter.add(5)

# Gauge
gauge = client.gauge("http.connections.active")
gauge.set(42)
gauge.inc()
gauge.dec()

# Histogram
histogram = client.histogram("http.request.duration", tags={"unit": "ms"})
histogram.record(45.2)
histogram.record(123.5)
```

### Java

```java
// Counter
Counter counter = sdk.counter("http.requests.total",
    Map.of("service", "api"));
counter.inc();
counter.add(5.0);

// Gauge
Gauge gauge = sdk.gauge("http.connections.active", Map.of());
gauge.set(42.0);
gauge.inc();
gauge.dec();

// Histogram
Histogram histogram = sdk.histogram("http.request.duration",
    Map.of("unit", "ms"));
histogram.record(45.2);
histogram.record(123.5);
```

### Ruby

```ruby
sdk = Tracekit.sdk

# Counter
counter = sdk.counter("http.requests.total", {
  service: "api",
  endpoint: "/users"
})
counter.add(1)
counter.add(5)

# Gauge
gauge = sdk.gauge("memory.usage.bytes")
gauge.set(1024 * 1024 * 512)
gauge.inc
gauge.dec

# Histogram
histogram = sdk.histogram("http.request.duration", {
  unit: "ms"
})
histogram.record(123.45)
histogram.record(67.89)
```

### PHP

```php
use TraceKit\PHP\MetricsRegistry;

$metrics = new MetricsRegistry($endpoint, $apiKey, 'my-service');

// Counter
$counter = $metrics->counter('http.requests.total', [
    'service' => 'api',
    'method' => 'GET',
]);
$counter->inc();
$counter->add(5);

// Gauge
$gauge = $metrics->gauge('http.connections.active');
$gauge->set(42);
$gauge->inc();
$gauge->dec();

// Histogram
$histogram = $metrics->histogram('http.request.duration', [
    'unit' => 'ms',
]);
$histogram->record(45.2);
$histogram->record(123.5);
```

### .NET

```csharp
var sdk = TracekitSDK.Create(config);

// Counter
var counter = sdk.Counter("http.requests.total",
    new() { ["service"] = "api" });
counter.Inc();
counter.Add(5);

// Gauge
var gauge = sdk.Gauge("http.connections.active");
gauge.Set(42);
gauge.Inc();
gauge.Dec();

// Histogram
var histogram = sdk.Histogram("http.request.duration",
    new() { ["unit"] = "ms" });
histogram.Record(45.2);
histogram.Record(123.5);
```

## Tags Best Practices

Tags let you slice and filter metrics in the dashboard. Use them for dimensions you want to group by.

**Good tags** (low cardinality):
- `service`, `environment`, `region`, `method`, `status_code`, `endpoint`

**Bad tags** (high cardinality  - avoid):
- `user_id`, `request_id`, `session_id`, `timestamp`, `trace_id`

Example with tags:

```javascript
const counter = client.counter('app.orders.completed_total', {
  payment_method: 'stripe',
  plan: 'premium',
  region: 'us-east-1'
});
counter.inc();
```

## Verification

After adding custom metrics:

1. **Deploy or restart** your application.
2. **Trigger the code paths** that record metrics (e.g., process some orders, handle some requests).
3. **Open** `https://app.tracekit.dev/metrics` (Metrics Explorer).
4. **Search** for your metric name (e.g., `http.requests.total`).
5. **Confirm** data points appear within 60 seconds.

## Troubleshooting

### Metrics not appearing in dashboard

- **Check SDK is initialized:** Metrics require the base TraceKit SDK to be initialized with a valid API key.
- **Check metric names:** Names must follow the dot-separated convention. Avoid spaces or special characters.
- **Check tag cardinality:** If you have too many unique tag combinations, metrics may be dropped.

### Unexpected metric values

- **Counters only go up:** If you need a value that decreases, use a gauge instead.
- **Histograms record distributions:** Individual values are not stored  - only aggregated percentiles (p50, p90, p99).
- **Check units:** Ensure you are recording in consistent units (e.g., always milliseconds, not sometimes seconds).

## Next Steps

- **Alerts**  - Set up alerts on metric thresholds using the `tracekit-alerts` skill
- **Dashboards**  - Create custom dashboards combining metrics and traces at `https://app.tracekit.dev/dashboards`
- **Distributed Tracing**  - Correlate metrics with traces for full observability

## References

- Custom Metrics docs: `https://app.tracekit.dev/docs/metrics`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
