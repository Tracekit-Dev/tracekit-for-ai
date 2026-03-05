---
name: tracekit-dotnet-sdk
description: Sets up TraceKit APM in .NET applications for automatic distributed tracing, error capture, and code monitoring. Supports ASP.NET Core with dependency injection and middleware patterns. Use when the user asks to add TraceKit, add observability, instrument a .NET service, or configure APM in a C# project.
---

# TraceKit .NET SDK Setup

> **Coming soon -- SDK in development.** The patterns below reflect the planned API. Package names and method signatures may change before GA release.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to a .NET application
- Add observability or APM to an ASP.NET Core project
- Instrument a C# service with distributed tracing
- Configure TraceKit API keys in a .NET project
- Debug production .NET services with live breakpoints
- Set up code monitoring in a .NET app

## Non-Negotiable Rules

1. **Never hardcode API keys** in code or `appsettings.json`. Always use `Environment.GetEnvironmentVariable("TRACEKIT_API_KEY")` or User Secrets.
2. **Always register TraceKit middleware before `MapControllers()`** -- `app.UseTraceKit()` must be in the pipeline before endpoint routing.
3. **Always include a verification step** confirming traces appear in `https://app.tracekit.dev/traces`.
4. **Always enable code monitoring** (`EnableCodeMonitoring = true`) -- it is TraceKit's differentiator.
5. **Use env vars or User Secrets for all secrets** -- never commit API keys to source control.

## Detection

Before applying this skill, detect the project type:

1. **Check for `*.csproj`** -- confirms this is a .NET project.
2. **Check for `*.sln`** -- confirms a .NET solution.
3. **Confirm ASP.NET Core** -- scan `.csproj` for `Microsoft.AspNetCore` or `Microsoft.NET.Sdk.Web` SDK attribute.
4. **Check .NET version** -- requires .NET 8.0 or higher.
5. **Only ask the user** if `.csproj` is missing or framework cannot be determined.

## Step 1: Environment Setup

Set the `TRACEKIT_API_KEY` environment variable. This is the only required secret.

Add to your environment:

```bash
export TRACEKIT_API_KEY=ctxio_your_api_key_here
```

Or use .NET User Secrets for local development:

```bash
dotnet user-secrets set "TraceKit:ApiKey" "ctxio_your_api_key_here"
```

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. Use environment variables, User Secrets, Azure Key Vault, or AWS Secrets Manager.

## Step 2: Install SDK

### For ASP.NET Core (recommended)

```bash
dotnet add package TraceKit.AspNetCore
```

### For vanilla .NET (console apps, workers)

```bash
dotnet add package TraceKit.Core
```

**Prerequisites:**
- .NET 8.0 or higher
- ASP.NET Core 8.0+ (for web applications)
- A TraceKit account ([create one free](https://app.tracekit.dev/register))

## Step 3: Configure in Program.cs

Add TraceKit to your ASP.NET Core application's service collection and middleware pipeline:

```csharp
using TraceKit.AspNetCore;

var builder = WebApplication.CreateBuilder(args);

// Add TraceKit services to the DI container
builder.Services.AddTraceKit(options =>
{
    options.ApiKey = Environment.GetEnvironmentVariable("TRACEKIT_API_KEY");
    options.ServiceName = "my-dotnet-service";
    options.Endpoint = "https://app.tracekit.dev/v1/traces";
    options.EnableCodeMonitoring = true;
});

// Add your other services
builder.Services.AddControllers();

var app = builder.Build();

// Add TraceKit middleware -- MUST be before MapControllers
app.UseTraceKit();

app.UseRouting();
app.UseAuthorization();
app.MapControllers();

app.Run();
```

**Key points:**
- `AddTraceKit()` registers all required services in the DI container
- `UseTraceKit()` adds the tracing middleware to the HTTP pipeline
- `UseTraceKit()` **must** be called before `MapControllers()` and `UseRouting()` to capture the full request lifecycle

## Step 4: Configuration via appsettings.json

As an alternative to inline configuration, use `appsettings.json`:

```json
{
  "TraceKit": {
    "ApiKey": "",
    "ServiceName": "my-dotnet-service",
    "Endpoint": "https://app.tracekit.dev/v1/traces",
    "EnableCodeMonitoring": true,
    "Tracing": {
      "Requests": true,
      "Database": true,
      "HttpClient": true,
      "Exceptions": true
    }
  }
}
```

Then bind in `Program.cs`:

```csharp
builder.Services.AddTraceKit(builder.Configuration.GetSection("TraceKit"));
```

**Important:** Do not put the actual API key in `appsettings.json`. Override it with an environment variable:

```bash
export TraceKit__ApiKey=ctxio_your_api_key_here
```

.NET's configuration system automatically merges environment variables with `appsettings.json` using the `__` separator.

## Step 5: Error Capture

TraceKit automatically captures unhandled exceptions via the middleware. For explicit error capture:

```csharp
using TraceKit;

try
{
    var result = await SomeRiskyOperationAsync();
}
catch (Exception ex)
{
    TracekitSdk.CaptureException(ex);
    throw; // Re-throw or handle
}
```

**Global exception handler middleware** (for custom error responses):

```csharp
app.UseExceptionHandler(exceptionHandlerApp =>
{
    exceptionHandlerApp.Run(async context =>
    {
        var exception = context.Features.Get<IExceptionHandlerFeature>()?.Error;
        if (exception != null)
        {
            TracekitSdk.CaptureException(exception);
        }

        context.Response.StatusCode = 500;
        await context.Response.WriteAsJsonAsync(new { error = "Internal server error" });
    });
});

// Place BEFORE UseTraceKit for proper ordering
app.UseTraceKit();
```

For adding context to traces, use manual spans:

```csharp
using TraceKit;

using var span = TracekitSdk.StartSpan("process-order");
span.SetAttribute("order.id", orderId);
span.SetAttribute("user.id", userId);

try
{
    var order = await ProcessOrderAsync(orderId);
}
catch (Exception ex)
{
    TracekitSdk.CaptureException(ex);
    throw;
}
```

## Step 6: Database Tracing

TraceKit automatically traces Entity Framework Core and ADO.NET queries when configured. Add EF Core tracing:

```csharp
builder.Services.AddDbContext<AppDbContext>(options =>
{
    options.UseSqlServer(connectionString);
    options.AddTraceKitInterceptor(); // Auto-trace all EF Core queries
});
```

For raw ADO.NET, wrap your connection:

```csharp
using TraceKit.Data;

var tracedConnection = new TracekitDbConnection(originalConnection);
```

Traced queries include:
- SQL statement (parameterized -- no sensitive data)
- Database system and name
- Query duration
- Connection details

## Step 7: HttpClient Tracing

Trace outgoing HTTP calls by adding the TraceKit handler to `HttpClient`:

```csharp
// Via DI (recommended)
builder.Services.AddHttpClient("external-api")
    .AddTraceKitHandler(); // Auto-trace all outgoing requests

// Usage in a controller or service
public class MyService
{
    private readonly HttpClient _httpClient;

    public MyService(IHttpClientFactory httpClientFactory)
    {
        _httpClient = httpClientFactory.CreateClient("external-api");
    }

    public async Task<string> GetDataAsync()
    {
        // This call is automatically traced
        var response = await _httpClient.GetAsync("https://api.example.com/data");
        return await response.Content.ReadAsStringAsync();
    }
}
```

Trace context headers are automatically injected into outgoing requests for distributed tracing across services.

## Step 8: Minimal API Support

TraceKit works with .NET Minimal APIs:

```csharp
var builder = WebApplication.CreateBuilder(args);

builder.Services.AddTraceKit(options =>
{
    options.ApiKey = Environment.GetEnvironmentVariable("TRACEKIT_API_KEY");
    options.ServiceName = "my-minimal-api";
    options.Endpoint = "https://app.tracekit.dev/v1/traces";
    options.EnableCodeMonitoring = true;
});

var app = builder.Build();

app.UseTraceKit();

app.MapGet("/api/users", () =>
{
    return Results.Ok(new[] { "alice", "bob" });
});

app.MapPost("/api/orders", (OrderRequest request) =>
{
    using var span = TracekitSdk.StartSpan("create-order");
    span.SetAttribute("order.item", request.Item);

    // Process order...
    return Results.Created($"/api/orders/{orderId}", order);
});

app.Run();
```

## Step 9: Verification

After integrating, verify traces are flowing:

1. **Start your application** with `TRACEKIT_API_KEY` set: `TRACEKIT_API_KEY=ctxio_... dotnet run`.
2. **Hit your endpoints 3-5 times** -- e.g., `curl http://localhost:5000/api/users`.
3. **Open** `https://app.tracekit.dev/traces`.
4. **Confirm** new spans and your service name appear within 30-60 seconds.

If traces do not appear, see Troubleshooting below.

## Troubleshooting

### Traces not appearing in dashboard

- **Check `TRACEKIT_API_KEY`:** Ensure the env var is set in the runtime environment. Verify: `Console.WriteLine(Environment.GetEnvironmentVariable("TRACEKIT_API_KEY"))`.
- **Check outbound access:** Your service must reach `https://app.tracekit.dev/v1/traces`. Verify with: `curl -X POST https://app.tracekit.dev/v1/traces` (expect 401 -- means the endpoint is reachable).
- **Check middleware order:** `app.UseTraceKit()` must be called **before** `app.MapControllers()`.

### Middleware order wrong

Symptoms: Server starts fine but no traces appear despite traffic.

Fix: Ensure `app.UseTraceKit()` is called before `app.UseRouting()` and `app.MapControllers()`:

```csharp
app.UseTraceKit();    // First
app.UseRouting();     // Second
app.UseAuthorization();
app.MapControllers(); // Last
```

### Missing environment variable

Symptoms: `ApiKey is null` warning on startup, or traces rejected by backend.

Fix: Ensure `TRACEKIT_API_KEY` is set. For local development, use User Secrets:

```bash
dotnet user-secrets set "TraceKit:ApiKey" "ctxio_your_key"
```

For production, use your platform's secret management (Azure Key Vault, AWS Secrets Manager, etc.).

### Service name collisions

Symptoms: Traces appear under the wrong service in the dashboard.

Fix: Use a unique `ServiceName` per deployed service. Avoid generic names like `"app"` or `"api"`.

### Entity Framework queries not traced

Symptoms: HTTP requests show traces but database queries do not.

Fix: Ensure `.AddTraceKitInterceptor()` is called on your `DbContextOptions`. If using multiple contexts, add it to each one.

## Next Steps

Once your .NET application is traced, consider:
- **Code Monitoring** -- Set live breakpoints and capture snapshots in production without redeploying (already enabled via `EnableCodeMonitoring = true`)
- **Distributed Tracing** -- Connect traces across multiple services for full request visibility
- **Frontend Observability** -- Add `@tracekit/browser` to your frontend for end-to-end trace correlation

## References

- .NET SDK docs: `https://app.tracekit.dev/docs/languages/dotnet`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
