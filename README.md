# TraceKit for AI

AI skills that teach your coding assistant how to set up TraceKit -- live breakpoints, distributed tracing, session replay, and error monitoring.

[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Skills](https://img.shields.io/badge/skills-20-green.svg)](skills/)

## What is this?

Structured skill files that AI coding assistants read to guide you through setting up TraceKit APM in any project. Each skill contains step-by-step instructions with working code snippets your AI assistant can apply directly.

Supports Claude Code, Cursor, and 38+ other AI tools via `npx skills add`. Skills follow a detect, configure, verify pattern -- your assistant identifies your stack, sets up TraceKit, and confirms data is flowing to your dashboard.

## Quick Start

### npx (works with 38+ AI tools)

```bash
npx skills add tracekit-dev/tracekit-for-ai
```

### Claude Code

```
/install-plugin https://github.com/tracekit-dev/tracekit-for-ai
```

### Cursor

Add `tracekit-dev/tracekit-for-ai` as a plugin source in Cursor settings.

## Usage

After installing, ask your AI assistant:

- "Set up TraceKit tracing in my Go project"
- "Add TraceKit error monitoring to my React app"
- "Enable session replay for my Next.js app"

Your assistant reads the relevant skill and walks you through setup.

## Available Skills

### Backend SDK Setup

| Skill | Directory | Description |
|-------|-----------|-------------|
| TraceKit Go SDK | `skills/tracekit-go-sdk/` | Set up distributed tracing and error monitoring in Go services |
| TraceKit Node.js SDK | `skills/tracekit-node-sdk/` | Set up tracing in Node.js/Express/Fastify/NestJS apps |
| TraceKit Python SDK | `skills/tracekit-python-sdk/` | Set up tracing in Python/Django/Flask/FastAPI apps |
| TraceKit PHP SDK | `skills/tracekit-php-sdk/` | Set up tracing in PHP applications |
| TraceKit Laravel SDK | `skills/tracekit-laravel-sdk/` | Set up tracing in Laravel applications |
| TraceKit Java SDK | `skills/tracekit-java-sdk/` | Set up tracing in Java/Spring Boot applications |
| TraceKit .NET SDK | `skills/tracekit-dotnet-sdk/` | Set up tracing in .NET/ASP.NET Core applications |
| TraceKit Ruby SDK | `skills/tracekit-ruby-sdk/` | Set up tracing in Ruby/Rails applications |

### Frontend SDK Setup

| Skill | Directory | Description |
|-------|-----------|-------------|
| TraceKit Browser SDK | `skills/tracekit-browser-sdk/` | Set up error monitoring and tracing in browser apps |
| TraceKit React | `skills/tracekit-react-sdk/` | Set up React error boundaries and component tracing |
| TraceKit Vue | `skills/tracekit-vue-sdk/` | Set up Vue error handler and navigation tracing |
| TraceKit Angular | `skills/tracekit-angular-sdk/` | Set up Angular ErrorHandler and route tracing |
| TraceKit Next.js | `skills/tracekit-nextjs-sdk/` | Set up Next.js multi-runtime tracing (server + client + edge) |
| TraceKit Nuxt | `skills/tracekit-nuxt-sdk/` | Set up Nuxt plugin and server middleware tracing |

### Feature Setup

| Skill | Directory | Description |
|-------|-----------|-------------|
| Code Monitoring | `skills/tracekit-code-monitoring/` | Enable live breakpoints and snapshots |
| Session Replay | `skills/tracekit-session-replay/` | Record and replay user sessions with privacy controls |
| Source Maps | `skills/tracekit-source-maps/` | Upload source maps for readable stack traces |
| Release Tracking | `skills/tracekit-releases/` | Track releases and deploy markers |
| Alerts | `skills/tracekit-alerts/` | Configure alert rules and notification channels |
| Distributed Tracing | `skills/tracekit-distributed-tracing/` | Set up frontend-to-backend trace propagation |

## How Skills Work

Each skill follows a three-step pattern:

1. **Detect** -- identifies your language, framework, and package manager from project files
2. **Configure** -- provides step-by-step setup with working code snippets your assistant applies directly
3. **Verify** -- confirms TraceKit is initialized and sending data to your dashboard

Skills include framework-specific variants, non-negotiable rules (never hardcode API keys, always use env vars), and links to TraceKit docs for advanced configuration.

## Links

- [TraceKit Dashboard](https://app.tracekit.dev)
- [Documentation](https://app.tracekit.dev/docs)
- [Quick Start Guide](https://app.tracekit.dev/docs/quickstart)

## License

Apache-2.0 -- see [LICENSE](LICENSE)
