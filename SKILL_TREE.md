# TraceKit Skill Tree

Master index of all TraceKit AI skills. Each skill teaches an AI coding assistant how to set up a specific SDK or feature in a user's project.

## How to Use

Install skills into your AI coding assistant:

```bash
# Using npx (works with 38+ AI tools)
npx skills add tracekit-dev/tracekit-for-ai

# Or curl for manual install
curl -sSL https://raw.githubusercontent.com/tracekit-dev/tracekit-for-ai/main/skills/<skill-name>/SKILL.md
```

All skills follow the **detect -> configure -> verify** pattern: detect the user's stack, apply the right configuration, and verify data appears in the TraceKit dashboard.

Authentication should be bootstrapped through `skills/tracekit-auth/` before asking users to manually create API keys.

---

## Backend SDK Setup

| Skill | Directory | Description | Status |
|-------|-----------|-------------|--------|
| TraceKit Go SDK | `skills/tracekit-go-sdk/` | Set up TraceKit APM in Go services (Gin, Echo, net/http) | Available |
| TraceKit Node SDK | `skills/tracekit-node-sdk/` | Set up TraceKit APM in Node.js apps (Express, Fastify, NestJS) | Available |
| TraceKit Python SDK | `skills/tracekit-python-sdk/` | Set up TraceKit APM in Python apps (Django, Flask, FastAPI) | Available |
| TraceKit PHP SDK | `skills/tracekit-php-sdk/` | Set up TraceKit APM in PHP applications | Available |
| TraceKit Laravel SDK | `skills/tracekit-laravel-sdk/` | Set up TraceKit APM in Laravel applications | Available |
| TraceKit Java SDK | `skills/tracekit-java-sdk/` | Set up TraceKit APM in Java apps (Spring Boot, Micronaut) | Available |
| TraceKit .NET SDK | `skills/tracekit-dotnet-sdk/` | Set up TraceKit APM in .NET applications | Available |
| TraceKit Ruby SDK | `skills/tracekit-ruby-sdk/` | Set up TraceKit APM in Ruby apps (Rails, Sinatra) | Available |

## Frontend SDK Setup

| Skill | Directory | Description | Status |
|-------|-----------|-------------|--------|
| TraceKit Browser SDK | `skills/tracekit-browser-sdk/` | Set up TraceKit APM in vanilla JavaScript/TypeScript apps | Available |
| TraceKit React SDK | `skills/tracekit-react-sdk/` | Set up TraceKit APM in React applications | Available |
| TraceKit Vue SDK | `skills/tracekit-vue-sdk/` | Set up TraceKit APM in Vue.js applications | Available |
| TraceKit Angular SDK | `skills/tracekit-angular-sdk/` | Set up TraceKit APM in Angular applications | Available |
| TraceKit Next.js SDK | `skills/tracekit-nextjs-sdk/` | Set up TraceKit APM in Next.js applications | Available |
| TraceKit Nuxt SDK | `skills/tracekit-nuxt-sdk/` | Set up TraceKit APM in Nuxt applications | Available |

## Feature Setup

| Skill | Directory | Description | Status |
|-------|-----------|-------------|--------|
| TraceKit Auth | `skills/tracekit-auth/` | Connect a user to TraceKit with email verification and save the production profile for future skills and MCP use | Available |
| TraceKit Code Monitoring | `skills/tracekit-code-monitoring/` | Enable live breakpoints and snapshots for production debugging | Available |
| TraceKit Session Replay | `skills/tracekit-session-replay/` | Record and replay user sessions with linked traces | Available |
| TraceKit Source Maps | `skills/tracekit-source-maps/` | Upload source maps for readable stack traces in production | Available |
| TraceKit Releases | `skills/tracekit-releases/` | Configure release tracking and deploy notifications | Available |
| TraceKit Alerts | `skills/tracekit-alerts/` | Set up alerting rules for errors, performance, and availability | Available |
| TraceKit Custom Metrics | `skills/tracekit-custom-metrics/` | Track business KPIs, counters, gauges, and histograms alongside traces | Available |
| TraceKit Distributed Tracing | `skills/tracekit-distributed-tracing/` | Connect frontend and backend traces across services | Available |
