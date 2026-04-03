---
name: tracekit-auth
description: Sets up TraceKit authentication for AI-assisted setup flows. Use when TraceKit credentials are missing, when a user needs to connect their account by email verification, or before TraceKit MCP and SDK setup skills that require access to the production TraceKit account.
---

# TraceKit Auth

Use this skill to connect the user to TraceKit through email verification without requiring them to manually sign up, log in separately, or create API keys first.

## When To Use

Use this skill when:
- `~/.tracekitconfig` is missing or does not contain `https://app.tracekit.dev`
- `TRACEKIT_API_KEY` is not set
- a TraceKit SDK, feature, or MCP flow needs authenticated production access
- the user asks to connect, sign in, sign up, authenticate, or set up TraceKit credentials

## Workflow

1. Check existing auth first:
   - Run `./scripts/run-tracekit-auth.sh status`
   - If `configured` is `true`, reuse the existing production profile and do not re-register unless the user asks
2. If auth is missing, ask the user for their email address
3. Start the email verification flow:
   - Run `./scripts/run-tracekit-auth.sh register --email <email> --service-name <service-name>`
   - Capture the returned `session_id`
4. Ask the user for the 6-digit verification code from email
5. Complete verification and persist auth:
   - Run `./scripts/run-tracekit-auth.sh verify --session-id <session_id> --code <code> --service-name <service-name>`
6. Confirm success by rerunning `./scripts/run-tracekit-auth.sh status`

This flow signs the user into the existing TraceKit account for that email when one exists, or creates the account automatically when it does not.

## Defaults

- Always use production endpoint: `https://app.tracekit.dev`
- Default tag: `prod`
- Default service name: `tracekit-ai`
- Prefer writing `~/.tracekitconfig` so both MCP and future TraceKit skills can reuse the same credentials

## Important Rules

1. Do not ask the user to go sign up, log in in a separate step, or manually create an API key when the email verification flow is available
2. Never print the full API key back to the user after verification
3. Reuse the existing saved production profile when present
4. If verification succeeds, continue with the originally requested TraceKit task instead of stopping at auth
