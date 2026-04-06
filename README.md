# TraceKit for AI

AI skills that teach your coding assistant how to set up TraceKit -- live breakpoints, distributed tracing, session replay, and error monitoring.

[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Skills](https://img.shields.io/badge/skills-20-green.svg)](skills/)

## What is this?

Structured skill files that AI coding assistants read to guide you through setting up TraceKit APM in any project. Each skill contains step-by-step instructions with working code snippets your AI assistant can apply directly.

Supports Claude Code, Cursor, and 38+ other AI tools via `npx skills add`. Skills follow a detect, configure, verify pattern -- your assistant identifies your stack, sets up TraceKit, and confirms data is flowing to your dashboard.

## Compatibility

This repo is the shared source for TraceKit AI integrations. The `skills/` content is reused across assistants, but each tool reads a different plugin or marketplace format.

| Tool | Skills install | MCP install | Uses shared `skills/` | Metadata source |
|------|----------------|-------------|-----------------------|-----------------|
| npx-compatible tools | `npx skills add tracekit-dev/tracekit-for-ai` | Tool-dependent | Yes | repo skills layout |
| Claude Code | `/install-plugin https://github.com/tracekit-dev/tracekit-for-ai` | Load [`.mcp.json`](./.mcp.json) or run `claude mcp add --scope project tracekit ./scripts/run-tracekit-mcp.sh` | Yes | `.claude-plugin/plugin.json` and `.mcp.json` |
| Cursor | Add `tracekit-dev/tracekit-for-ai` as a plugin source | Load [`.cursor/mcp.json`](./.cursor/mcp.json) | Yes | `.cursor-plugin/plugin.json` and `.cursor/mcp.json` |
| Codex | Load [`.agents/plugins/marketplace.json`](./.agents/plugins/marketplace.json) and install `TraceKit` | Run `codex mcp add tracekit -- /absolute/path/to/tracekit-for-ai/scripts/run-tracekit-mcp.sh` | Yes | `.agents/plugins/marketplace.json`, `plugins/tracekit/.codex-plugin/plugin.json`, and `scripts/run-tracekit-mcp.sh` |

Short version: one TraceKit repo, one shared skill set, multiple assistant-specific install paths.

## Quick Start

### npx (works with 38+ AI tools)

```bash
npx skills add tracekit-dev/tracekit-for-ai
```

To install all 23 skills at once (skips the interactive picker):

```bash
npx skills add tracekit-dev/tracekit-for-ai --all
```

This installs the shared skills. MCP availability depends on whether that tool also supports loading this repo's MCP config.

### Claude Code

```
/install-plugin https://github.com/tracekit-dev/tracekit-for-ai
```

For MCP in Claude Code, either:

```bash
claude mcp add --scope project tracekit ./scripts/run-tracekit-mcp.sh
```

or open this repo's [`.mcp.json`](./.mcp.json) as the project MCP config.

### Cursor

Add `tracekit-dev/tracekit-for-ai` as a plugin source in Cursor settings.

For MCP in Cursor, load [`.cursor/mcp.json`](./.cursor/mcp.json) in the project.

### Codex

Clone this repo and point Codex at [`.agents/plugins/marketplace.json`](./.agents/plugins/marketplace.json), then install the `TraceKit` plugin from that local marketplace.

For MCP in Codex, add the local launcher script as a global MCP server using an absolute path:

```bash
codex mcp add tracekit -- /absolute/path/to/tracekit-for-ai/scripts/run-tracekit-mcp.sh
```

`tracekit` is just the server name shown inside Codex. Use an absolute path here because Codex stores the MCP server globally, so relative paths can break when Codex starts from another directory.

On first use, the bundled launcher scripts automatically download the correct `tracekit-agent` binary from the latest GitHub release into [`bin/`](./bin), so end users do not need Go or Python installed.

## Authentication

Users should not be sent off to sign up or manually create an API key first if the assistant can walk them through the TraceKit email verification flow.

- Auth skill: [`skills/tracekit-auth/SKILL.md`](./skills/tracekit-auth/SKILL.md)
- Auth launcher: [`scripts/run-tracekit-auth.sh`](./scripts/run-tracekit-auth.sh)
- Unified agent binary entrypoint: [`scripts/run-tracekit-agent.sh`](./scripts/run-tracekit-agent.sh)

The helper supports:

- `./scripts/run-tracekit-auth.sh status`
- `./scripts/run-tracekit-auth.sh register --email <email>`
- `./scripts/run-tracekit-auth.sh verify --session-id <session_id> --code <code>`

If the local agent binary is missing, the launcher downloads the correct release binary automatically before running the auth flow.

Successful verification signs the user into the existing account for that email or creates it automatically, then writes the production profile to `~/.tracekitconfig` so both the MCP server and future TraceKit skills can reuse the same credentials.

## MCP Server

This repo now includes a local TraceKit MCP server for agents that support MCP tool servers.

- Shared MCP config: [`.mcp.json`](./.mcp.json)
- Cursor MCP config: [`.cursor/mcp.json`](./.cursor/mcp.json)
- Runner script: [`scripts/run-tracekit-mcp.sh`](./scripts/run-tracekit-mcp.sh)
- Unified agent source: [`agent/tracekit-agent`](./agent/tracekit-agent)
- Release binaries directory: [`bin/`](./bin)

The server currently exposes read-focused tools:

- `tracekit_status`
- `tracekit_dashboard`
- `tracekit_services`
- `tracekit_service_detail`
- `tracekit_traces`
- `tracekit_alert_rules`
- `tracekit_triage_inbox`

### MCP Auth

The MCP server uses the same TraceKit credentials model as the CLI:

1. Preferred: the assistant uses `tracekit-auth` or `./scripts/run-tracekit-auth.sh` to write `~/.tracekitconfig`
2. Also supported: `tracekit login` and stored credentials in `~/.tracekitconfig`
3. Optional override: `TRACEKIT_API_KEY`, `TRACEKIT_USER_ID`, and `TRACEKIT_ENDPOINT`

### MCP Requirements

- For released binaries in `bin/`: no language runtime required
- If no local binary is present, the launcher self-installs the correct one from the latest GitHub release
- For development fallback when no local binary exists: Go 1.22 or newer
- valid TraceKit credentials

### How To Test

1. Start the auth flow:

```bash
./scripts/run-tracekit-auth.sh register --email you@example.com
```

Then verify with the emailed code:

```bash
./scripts/run-tracekit-auth.sh verify --session-id <session_id> --code <code>
```

2. Smoke test the MCP server startup:

```bash
./scripts/run-tracekit-mcp.sh
```

It will wait for MCP stdio messages. Press `Ctrl+C` to stop it.

3. Send a manual MCP initialize + tool call:

```bash
cd tracekit-for-ai

init='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"manual-test","version":"0.0.0"}}}'
call='{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"tracekit_status","arguments":{}}}'

{
  printf 'Content-Length: %s\r\n\r\n%s' ${#init} "$init"
  printf 'Content-Length: %s\r\n\r\n%s' ${#call} "$call"
} | ./scripts/run-tracekit-mcp.sh
```

4. Test in Claude Code by loading [`.mcp.json`](./.mcp.json) or running:

```bash
claude mcp add --scope project tracekit ./scripts/run-tracekit-mcp.sh
```

5. Test in Cursor by loading [`.cursor/mcp.json`](./.cursor/mcp.json) for the project.

6. Test with Codex by:

```bash
codex mcp add tracekit -- /absolute/path/to/tracekit-for-ai/scripts/run-tracekit-mcp.sh
```

Then load [`.agents/plugins/marketplace.json`](./.agents/plugins/marketplace.json), install `TraceKit`, and check that the MCP server is available.

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
