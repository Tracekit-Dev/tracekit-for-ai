---
name: tracekit-apm-setup
description: Entry point skill for setting up TraceKit APM. Detects the user's stack and routes to the appropriate SDK skill.
---

# TraceKit APM Setup

This is the entry point skill for TraceKit APM. It detects the user's technology stack and routes to the appropriate SDK-specific skill for detailed setup instructions.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to their project (without specifying a language/framework)
- Set up APM, observability, or error tracking (general request)
- Get started with TraceKit
- Choose the right TraceKit SDK for their project

## Detection

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

## Next Steps

After detecting the stack, route to the appropriate skill:

### Backend SDKs

- **Go** -- `tracekit-go-sdk` skill (Gin, Echo, net/http)
- **Node.js** -- `tracekit-node-sdk` skill (Express, Fastify, NestJS)
- **Python** -- `tracekit-python-sdk` skill (Django, Flask, FastAPI)
- **PHP** -- `tracekit-php-sdk` skill
- **Laravel** -- `tracekit-laravel-sdk` skill
- **Java** -- `tracekit-java-sdk` skill (Spring Boot, Micronaut)
- **.NET** -- `tracekit-dotnet-sdk` skill (ASP.NET Core)
- **Ruby** -- `tracekit-ruby-sdk` skill (Rails, Sinatra)

### Frontend SDKs

- **Browser (vanilla JS/TS)** -- `tracekit-browser-sdk` skill
- **React** -- `tracekit-react-sdk` skill (ErrorBoundary, React Router breadcrumbs)
- **Vue** -- `tracekit-vue-sdk` skill (Vue plugin, Vue Router breadcrumbs)
- **Angular** -- `tracekit-angular-sdk` skill (NgModule/standalone, DI ErrorHandler)
- **Next.js** -- `tracekit-nextjs-sdk` skill (multi-runtime, App Router/Pages Router)
- **Nuxt** -- `tracekit-nuxt-sdk` skill (Nuxt module, defineNuxtPlugin)

## References

- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
