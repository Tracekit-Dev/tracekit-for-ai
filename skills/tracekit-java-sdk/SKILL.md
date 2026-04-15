---
name: tracekit-java-sdk
description: Sets up TraceKit APM in Java applications for automatic distributed tracing, error capture, and code monitoring. Supports Spring Boot and Micronaut frameworks with Maven and Gradle build systems. Includes LLM instrumentation via OkHttp interceptor for OpenAI and Anthropic API call monitoring. Use when the user asks to add TraceKit, add observability, instrument a Java service, or configure APM in a Java project.
---

# TraceKit Java SDK Setup
## Auth Bootstrap

Do not tell the user to go sign up, log in separately, or manually create an API key before setup. First check for existing TraceKit auth. If `~/.tracekitconfig` does not contain the production profile or `TRACEKIT_API_KEY` is missing, apply the `tracekit-auth` skill first. Use `./scripts/run-tracekit-auth.sh status` to check, then guide the user through the TraceKit email verification flow with `register` and `verify`. That flow signs the user into an existing account for that email or creates the account automatically, then saves the returned credentials for the rest of the setup.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to a Java service
- Add observability or APM to a Java application
- Instrument a Java service with distributed tracing
- Configure TraceKit API keys in a Java project
- Debug production Java services with live breakpoints
- Set up code monitoring in a Java app
- Add tracing to a Spring Boot or Micronaut application
- Monitor OpenAI or Anthropic API calls in a Java service
- Add LLM observability to a Java application

## Non-Negotiable Rules

1. **Never hardcode API keys** in code or config files. Always use `System.getenv("TRACEKIT_API_KEY")` or externalized config.
2. **Always initialize TraceKit before starting the HTTP server** -- the SDK must be configured before routes are registered.
3. **Always include a verification step** confirming traces appear in `https://app.tracekit.dev/traces`.
4. **Always enable code monitoring** (`enableCodeMonitoring: true`) -- it is TraceKit's differentiator.
5. **Use env vars or externalized config for all secrets** -- never commit API keys to source control.

## Detection

Before applying this skill, detect the project type:

1. **Check for `pom.xml`** -- confirms Maven build system.
2. **Check for `build.gradle` or `build.gradle.kts`** -- confirms Gradle build system.
3. **Detect framework** by scanning dependencies:
   - `spring-boot-starter-web` or `org.springframework.boot` => Spring Boot (use Spring Boot branch)
   - `io.micronaut` => Micronaut (use Micronaut branch)
   - Neither => vanilla Java (use vanilla Java branch)
4. **Only ask the user** if multiple frameworks are detected or build files are missing.

## Step 1: Environment Setup

Set the `TRACEKIT_API_KEY` environment variable. This is the only required secret.

Add to your environment or `.env` file:

```bash
export TRACEKIT_API_KEY=ctxio_your_api_key_here
```

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. Use environment variables, secret managers, or CI/CD secrets.

## Step 2: Install SDK

Choose your build system and framework.

### Maven -- Spring Boot

Add to `pom.xml`:

```xml
<dependency>
    <groupId>dev.tracekit</groupId>
    <artifactId>tracekit-spring-boot-starter</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Maven -- Vanilla Java / Micronaut

Add to `pom.xml`:

```xml
<dependency>
    <groupId>dev.tracekit</groupId>
    <artifactId>tracekit-core</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Gradle -- Spring Boot

Add to `build.gradle`:

```groovy
implementation 'dev.tracekit:tracekit-spring-boot-starter:1.0.0'
```

Or `build.gradle.kts`:

```kotlin
implementation("dev.tracekit:tracekit-spring-boot-starter:1.0.0")
```

### Gradle -- Vanilla Java / Micronaut

Add to `build.gradle`:

```groovy
implementation 'dev.tracekit:tracekit-core:1.0.0'
```

Or `build.gradle.kts`:

```kotlin
implementation("dev.tracekit:tracekit-core:1.0.0")
```

**Prerequisites:**
- Java 8 or higher (Java 11+ recommended)
- Maven or Gradle build system
- A TraceKit account ([create one free](https://app.tracekit.dev/register))

## Step 3: Framework Integration

Choose the branch matching your framework. Apply **one** of the following.

### Branch A: Spring Boot (Recommended)

**Configuration via `application.yml`:**

```yaml
tracekit:
  api-key: ${TRACEKIT_API_KEY}
  service-name: my-spring-service
  endpoint: https://app.tracekit.dev/v1/traces
  enable-code-monitoring: true
```

Or via `application.properties`:

```properties
tracekit.api-key=${TRACEKIT_API_KEY}
tracekit.service-name=my-spring-service
tracekit.endpoint=https://app.tracekit.dev/v1/traces
tracekit.enable-code-monitoring=true
```

The Spring Boot starter provides auto-configuration -- no `@Bean` definitions needed. It automatically:
- Registers a servlet filter for HTTP request tracing
- Traces JDBC database queries
- Traces outgoing `RestTemplate` and `WebClient` calls
- Captures unhandled exceptions
- Propagates trace context to downstream services

**Manual bean registration** (only if auto-configuration is disabled):

```java
import dev.tracekit.TracekitAutoConfiguration;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Import;

@Configuration
@Import(TracekitAutoConfiguration.class)
public class TracekitConfig {
    // Auto-configuration handles everything
}
```

**Custom filter registration** (only if you need specific ordering):

```java
import dev.tracekit.spring.TracekitFilter;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class TracekitFilterConfig {

    @Bean
    public FilterRegistrationBean<TracekitFilter> tracekitFilter() {
        FilterRegistrationBean<TracekitFilter> registration = new FilterRegistrationBean<>();
        registration.setFilter(new TracekitFilter());
        registration.addUrlPatterns("/*");
        registration.setOrder(1); // Run early to capture full request
        return registration;
    }
}
```

### Branch B: Micronaut

**Configuration via `application.yml`:**

```yaml
tracekit:
  api-key: ${TRACEKIT_API_KEY}
  service-name: my-micronaut-service
  endpoint: https://app.tracekit.dev/v1/traces
  enable-code-monitoring: true
```

**Register the TraceKit bean and filter:**

```java
import dev.tracekit.Tracekit;
import dev.tracekit.TracekitConfig;
import io.micronaut.context.annotation.Factory;
import jakarta.inject.Singleton;

@Factory
public class TracekitFactory {

    @Singleton
    public Tracekit tracekit() {
        return new Tracekit(TracekitConfig.builder()
            .apiKey(System.getenv("TRACEKIT_API_KEY"))
            .serviceName("my-micronaut-service")
            .endpoint("https://app.tracekit.dev/v1/traces")
            .enableCodeMonitoring(true)
            .build());
    }
}
```

**HTTP filter for request tracing:**

```java
import dev.tracekit.Tracekit;
import io.micronaut.http.HttpRequest;
import io.micronaut.http.MutableHttpResponse;
import io.micronaut.http.annotation.Filter;
import io.micronaut.http.filter.HttpServerFilter;
import io.micronaut.http.filter.ServerFilterChain;
import jakarta.inject.Inject;
import org.reactivestreams.Publisher;

@Filter("/**")
public class TracekitHttpFilter implements HttpServerFilter {

    @Inject
    private Tracekit tracekit;

    @Override
    public Publisher<MutableHttpResponse<?>> doFilter(HttpRequest<?> request, ServerFilterChain chain) {
        return tracekit.traceRequest(request, chain);
    }
}
```

### Branch C: Vanilla Java

**Initialize TraceKit in your `main()` method:**

```java
import dev.tracekit.Tracekit;
import dev.tracekit.TracekitConfig;

public class Main {
    public static void main(String[] args) {
        // Initialize TraceKit -- MUST be before server start
        Tracekit tracekit = new Tracekit(TracekitConfig.builder()
            .apiKey(System.getenv("TRACEKIT_API_KEY"))
            .serviceName("my-java-service")
            .endpoint("https://app.tracekit.dev/v1/traces")
            .enableCodeMonitoring(true)
            .build());

        // Register shutdown hook to flush pending traces
        Runtime.getRuntime().addShutdownHook(new Thread(tracekit::shutdown));

        // ... start your HTTP server with tracekit filter
    }
}
```

## Step 4: Error Capture

Capture exceptions explicitly where you handle them:

```java
import dev.tracekit.Tracekit;

try {
    Object result = someOperation();
} catch (Exception e) {
    Tracekit.captureException(e);
    // Handle the error...
}
```

For adding context to traces, use manual spans:

```java
import dev.tracekit.Span;
import dev.tracekit.Tracekit;

Span span = Tracekit.startSpan("process-order");
span.setAttribute("order.id", orderId);
span.setAttribute("user.id", userId);

try {
    Order order = processOrder(orderId);
} catch (Exception e) {
    Tracekit.captureException(e);
    throw e;
} finally {
    span.end();
}
```

## Step 4b: Snapshot Capture (Code Monitoring)

For programmatic snapshots, **use the SnapshotClient directly**  - do not call through the SDK wrapper. The SDK uses stack inspection internally to identify the call site. Adding extra layers shifts the frame and causes snapshots to report the wrong source location.

Create a `Breakpoints` utility class:

```java
import dev.tracekit.SnapshotClient;
import dev.tracekit.Tracekit;
import java.util.Map;

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

Initialize after SDK setup:

```java
Breakpoints.init(tracekit);
```

Use at call sites:

```java
Breakpoints.capture("payment-failed", Map.of("orderId", orderId, "error", e.getMessage()));
```

See the `tracekit-code-monitoring` skill for the full pattern across all languages.

### Code Monitoring v25.0 Features

The latest SDK release adds these code monitoring capabilities (configured via the dashboard, no code changes needed):

- **Logpoint mode** -- capture expressions only without full variable snapshots, reducing overhead
- **Per-breakpoint limits** -- individual max captures (default: 100) and rate limits per breakpoint
- **Dynamic stack traces** -- configurable stack depth per breakpoint (1-50 frames)
- **Idle auto-expiry** -- breakpoints auto-expire after inactivity (default: 24h), pinnable to prevent expiry
- **Conditional expressions** -- server-side evaluated conditions with full operator support (`==`, `!=`, `>`, `<`, `&&`, `||`)

These features are available when `enableCodeMonitoring` is set to `true`. No SDK code changes required -- all configuration is done through the TraceKit dashboard.

For full details, see the `tracekit-code-monitoring` skill.

## Step 5: JDBC Database Tracing

For Spring Boot, JDBC tracing is automatic via the auto-configuration.

For Micronaut or vanilla Java, wrap your `DataSource`:

```java
import dev.tracekit.jdbc.TracekitDataSource;
import javax.sql.DataSource;

DataSource tracedDataSource = new TracekitDataSource(originalDataSource);
```

This traces all SQL queries with:
- SQL statement (parameterized -- no sensitive data)
- Database system and name
- Query duration

## Step 6: Outgoing HTTP Call Tracing

**Spring Boot (RestTemplate):**

```java
import dev.tracekit.spring.TracekitRestTemplateInterceptor;
import org.springframework.web.client.RestTemplate;

@Bean
public RestTemplate restTemplate() {
    RestTemplate restTemplate = new RestTemplate();
    restTemplate.getInterceptors().add(new TracekitRestTemplateInterceptor());
    return restTemplate;
}
```

**Spring Boot (WebClient):**

```java
import dev.tracekit.spring.TracekitWebClientFilter;
import org.springframework.web.reactive.function.client.WebClient;

@Bean
public WebClient webClient() {
    return WebClient.builder()
        .filter(new TracekitWebClientFilter())
        .build();
}
```

**Vanilla Java (manual):**

```java
Span span = Tracekit.startSpan("http-client");
span.setAttribute("http.method", "GET");
span.setAttribute("http.url", "https://api.example.com/data");

try {
    HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
    span.setAttribute("http.status_code", response.statusCode());
} finally {
    span.end();
}
```

## Step 7: Verification

After integrating, verify traces are flowing:

1. **Start your application** with `TRACEKIT_API_KEY` set in the environment.
2. **Hit your endpoints 3-5 times** -- e.g., `curl http://localhost:8080/api/users`.
3. **Open** `https://app.tracekit.dev/traces`.
4. **Confirm** new spans and your service name appear within 30-60 seconds.

If traces do not appear, see Troubleshooting below.

## Troubleshooting

### Traces not appearing in dashboard

- **Check `TRACEKIT_API_KEY`:** Ensure the env var is set in the runtime environment. Verify: `System.out.println(System.getenv("TRACEKIT_API_KEY"))`.
- **Check outbound access:** Your service must reach `https://app.tracekit.dev/v1/traces`. Verify with: `curl -X POST https://app.tracekit.dev/v1/traces` (expect 401 -- means the endpoint is reachable).
- **Check init order:** TraceKit must be initialized **before** starting the HTTP server.

### Spring Boot auto-configuration not loading

Symptoms: No tracing despite correct `application.yml` config.

Fix: Ensure `tracekit-spring-boot-starter` is on the classpath. Check `@SpringBootApplication` scan includes the TraceKit package. Run with `--debug` to see auto-configuration report.

### Micronaut bean not found

Symptoms: `No bean of type [dev.tracekit.Tracekit] exists` error.

Fix: Ensure the `TracekitFactory` is in a package scanned by Micronaut. Check that `tracekit-core` is on the classpath.

### Missing environment variable

Symptoms: `NullPointerException` on startup or traces rejected by backend.

Fix: Ensure `TRACEKIT_API_KEY` is exported in your shell, `.env` file, Docker Compose, or deployment config. For Spring Boot, use `${TRACEKIT_API_KEY}` placeholder syntax in `application.yml`.

### Service name collisions

Symptoms: Traces appear under the wrong service in the dashboard.

Fix: Use a unique `service-name` per deployed service. Avoid generic names like `"app"` or `"service"`.

## LLM Instrumentation (Manual Setup)

TraceKit can instrument OpenAI and Anthropic API calls made via OkHttp. This requires manual interceptor setup.

### When To Use

Add this when the user:
- Uses OpenAI or Anthropic APIs in their Java service
- Wants to monitor LLM cost, tokens, and latency
- Asks about AI observability in Java

### Setup

Add the TraceKit LLM interceptor to the OkHttpClient used for LLM API calls:

```java
import dev.tracekit.llm.TracekitLlmInterceptor;
import dev.tracekit.llm.LlmConfig;
import okhttp3.OkHttpClient;

// Default config (content capture off)
OkHttpClient llmClient = new OkHttpClient.Builder()
    .addInterceptor(new TracekitLlmInterceptor())
    .build();

// With custom config
OkHttpClient llmClient = new OkHttpClient.Builder()
    .addInterceptor(new TracekitLlmInterceptor(LlmConfig.builder()
        .captureContent(true)  // Enable prompt/completion capture
        .build()))
    .build();
```

Pass this client to your OpenAI or Anthropic SDK's HTTP client configuration.

### Environment Variable

Set `TRACEKIT_LLM_CAPTURE_CONTENT=true` to enable prompt/completion capture without code changes.

### Captured Attributes

LLM spans include: `gen_ai.system`, `gen_ai.request.model`, `gen_ai.usage.input_tokens`, `gen_ai.usage.output_tokens`, `gen_ai.response.finish_reasons`. Streaming responses produce a single span with accumulated token counts.

### Verification

After adding the interceptor, make an LLM API call and verify the span appears in the TraceKit dashboard under **LLM Observability** (`/ai/llm`).

## Next Steps

Once your Java service is traced, consider:
- **Code Monitoring** -- Set live breakpoints and capture snapshots in production without redeploying (already enabled via `enableCodeMonitoring: true`)
- **Distributed Tracing** -- Connect traces across multiple services for full request visibility
- **Frontend Observability** -- Add `@tracekit/browser` to your frontend for end-to-end trace correlation

## References

- Java SDK docs: `https://app.tracekit.dev/docs/languages/java`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
