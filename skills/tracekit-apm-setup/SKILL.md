---
name: tracekit-apm-setup
description: Entry point skill for setting up TraceKit APM. Detects the user's stack and routes to the appropriate SDK skill.
---

# TraceKit APM Setup

## Auth Bootstrap

Do not tell the user to go sign up, log in separately, or manually create an API key before setup. First check for existing TraceKit auth. If `~/.tracekitconfig` does not contain the production profile or `TRACEKIT_API_KEY` is missing, apply the `tracekit-auth` skill first. Use `./scripts/run-tracekit-auth.sh status` to check, then guide the user through the TraceKit email verification flow with `register` and `verify`. That flow signs the user into an existing account for that email or creates the account automatically, then saves the returned credentials for the rest of the setup.

This is the entry point skill for TraceKit APM. It detects the user's technology stack, asks which features they want, and routes to the appropriate SDK and feature skills for setup.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to their project (without specifying a language/framework)
- Set up APM, observability, or error tracking (general request)
- Get started with TraceKit
- Choose the right TraceKit SDK for their project

## Step 1: Detection

Before stack detection, ensure TraceKit auth exists for production setup flows. If the user is not connected yet, apply the `tracekit-auth` skill first. Do not redirect them to a separate signup or API-key-generation flow. The auth skill should handle the email verification flow, connect the existing account for that email or create it automatically, save the production credentials, and then continue directly with SDK setup.

Scan the project to determine the technology stack:

1. **Check for `package.json`** -- Node.js/JavaScript ecosystem
   - `next` in dependencies => Use `tracekit-nextjs-sdk` skill
   - `nuxt` in dependencies => Use `tracekit-nuxt-sdk` skill
   - `@angular/core` in dependencies => Use `tracekit-angular-sdk` skill
   - `react` in dependencies (without `next`) => Use `tracekit-react-sdk` skill
   - `vue` in dependencies (without `nuxt`) => Use `tracekit-vue-sdk` skill
   - `express`, `fastify`, or `@nestjs/core` => Use `tracekit-node-sdk` skill
   - None of the above => Use `tracekit-browser-sdk` skill (vanilla JS/TS)
2. **Check for `go.mod`** -- Use `tracekit-go-sdk` skill
3. **Check for `requirements.txt` or `pyproject.toml`** -- Use `tracekit-python-sdk` skill
4. **Check for `composer.json`** -- PHP ecosystem
   - `laravel/framework` in require => Use `tracekit-laravel-sdk` skill
   - Otherwise => Use `tracekit-php-sdk` skill
5. **Check for `pom.xml` or `build.gradle`** -- Use `tracekit-java-sdk` skill
6. **Check for `*.csproj` or `*.sln`** -- Use `tracekit-dotnet-sdk` skill
7. **Check for `Gemfile`** -- Use `tracekit-ruby-sdk` skill

## Step 2: Ask Which Features to Set Up

After detecting the stack, **ask the user which TraceKit features they want to enable**. Do not assume — present the options and let them choose. Users may want just error tracking, or the full suite.

Present the following feature menu:

> **Which TraceKit features would you like to set up?** (you can pick multiple)
>
> 1. **Error Tracking** — Automatic error capture with stack traces and context *(included in base SDK setup)*
> 2. **Distributed Tracing** — Request tracing across services with performance monitoring *(included in base SDK setup)*
> 3. **Code Monitoring** — Live breakpoints and snapshots for production debugging without redeploying *(included in base SDK setup, enable with `enable_code_monitoring=True`)*
> 4. **Custom Metrics** — Track business KPIs, counters, gauges, and histograms alongside traces
> 5. **Session Replay** *(frontend only)* — Record and replay user sessions linked to traces
> 6. **Source Maps** *(frontend only)* — Readable stack traces from minified JavaScript
> 7. **Release Tracking** — Monitor crash-free rates, deploy health, and regressions per release
> 8. **Alerts** — Set up notifications for errors, performance degradation, and availability
> 9. **All of the above** — Full observability suite

**Default behavior:** If the user says "just set it up" or doesn't specify, set up options 1-3 (Error Tracking, Distributed Tracing, and Code Monitoring) via the base SDK skill. Mention the other features are available and can be added later.

## Step 3: Route to Skills

Based on the detected stack and selected features, apply skills in this order:

### 1. Base SDK Skill (always apply first)

Route to the SDK skill matching the detected stack. This sets up error tracking, distributed tracing, and code monitoring.

#### Backend SDKs

- **Go** — `tracekit-go-sdk` skill (Gin, Echo, net/http)
- **Node.js** — `tracekit-node-sdk` skill (Express, Fastify, NestJS)
- **Python** — `tracekit-python-sdk` skill (Django, Flask, FastAPI)
- **PHP** — `tracekit-php-sdk` skill
- **Laravel** — `tracekit-laravel-sdk` skill
- **Java** — `tracekit-java-sdk` skill (Spring Boot, Micronaut)
- **.NET** — `tracekit-dotnet-sdk` skill (ASP.NET Core)
- **Ruby** — `tracekit-ruby-sdk` skill (Rails, Sinatra)

#### Frontend SDKs

- **Browser (vanilla JS/TS)** — `tracekit-browser-sdk` skill
- **React** — `tracekit-react-sdk` skill (ErrorBoundary, React Router breadcrumbs)
- **Vue** — `tracekit-vue-sdk` skill (Vue plugin, Vue Router breadcrumbs)
- **Angular** — `tracekit-angular-sdk` skill (NgModule/standalone, DI ErrorHandler)
- **Next.js** — `tracekit-nextjs-sdk` skill (multi-runtime, App Router/Pages Router)
- **Nuxt** — `tracekit-nuxt-sdk` skill (Nuxt module, defineNuxtPlugin)

### 2. Feature Skills (apply based on user selection)

After the base SDK is set up, apply additional feature skills as selected:

- **Session Replay** — `tracekit-session-replay` skill *(frontend projects only)*
- **Source Maps** — `tracekit-source-maps` skill *(frontend projects only)*
- **Release Tracking** — `tracekit-releases` skill
- **Alerts** — `tracekit-alerts` skill
- **Custom Metrics** — `tracekit-custom-metrics` skill
- **Distributed Tracing (multi-service)** — `tracekit-distributed-tracing` skill *(when connecting frontend + backend or multiple services)*
- **Code Monitoring (advanced)** — `tracekit-code-monitoring` skill *(for programmatic snapshots beyond the base setup)*

## References

- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
