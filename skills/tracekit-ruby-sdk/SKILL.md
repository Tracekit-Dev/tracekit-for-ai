---
name: tracekit-ruby-sdk
description: Sets up TraceKit APM in Ruby applications for automatic distributed tracing, error capture, and code monitoring. Supports Rails and Sinatra frameworks. Use when the user asks to add TraceKit, add observability, instrument a Ruby app, or configure APM in a Ruby/Rails project.
---

# TraceKit Ruby SDK Setup

> **Coming soon -- SDK in development.** The TraceKit Ruby SDK (`tracekit` gem) is not yet published. The setup patterns below reflect the planned API. This skill will be updated when the SDK ships.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to a Ruby application
- Add observability or APM to a Rails or Sinatra app
- Instrument a Ruby service with distributed tracing
- Configure TraceKit API keys in a Ruby project
- Debug production Ruby services with live breakpoints
- Set up code monitoring in a Ruby app

## Non-Negotiable Rules

1. **Never hardcode API keys** in code. Always use `ENV["TRACEKIT_API_KEY"]`.
2. **For Rails:** Rely on the Railtie for auto-initialization — just set environment variables. Do **not** call `Tracekit.configure` manually in a Rails app (conflicts with Railtie).
3. **For Sinatra/Rack:** Call `Tracekit.configure` before defining routes and add `use Tracekit::Middleware`.
4. **Always include a verification step** confirming traces appear in `https://app.tracekit.dev/traces`.
5. **Always enable code monitoring** (`enable_code_monitoring: true`) — it is TraceKit's differentiator.

## Detection

Before applying this skill, detect the project type:

1. **Check for `Gemfile`** — confirms this is a Ruby project.
2. **Detect framework** by scanning `Gemfile` for gems:
   - `gem 'rails'` or `gem "rails"` => Rails framework (use Rails branch)
   - `gem 'sinatra'` or `gem "sinatra"` => Sinatra framework (use Sinatra branch)
3. **Check for Rails directory structure:** `config/application.rb`, `config/initializers/` directory.
4. **Only ask the user** if neither framework is detected or if `Gemfile` is missing.

## Step 1: Environment Setup

Set the `TRACEKIT_API_KEY` environment variable. This is the only required secret.

Add to your `.env` file or environment:

```bash
export TRACEKIT_API_KEY=ctxio_your_api_key_here
```

For Rails, you can also set these optional environment variables:

```bash
TRACEKIT_SERVICE_NAME=my-rails-app
TRACEKIT_ENVIRONMENT=production
TRACEKIT_CODE_MONITORING=true
```

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. Use `.env` files, deployment secret managers, or CI variables.

## Step 2: Install SDK

Add to your `Gemfile`:

```ruby
gem 'tracekit'
```

Then install:

```bash
bundle install
```

This installs the TraceKit Ruby SDK with built-in OpenTelemetry support, Rack middleware, and code monitoring.

## Step 3: Initialize TraceKit

Choose the branch matching your framework. Rails and Sinatra have different initialization patterns.

### Branch A: Rails

Rails uses a Railtie-based auto-initialization. Just set environment variables -- no manual init code needed.

**1. Set environment variables** (in `.env` or your deployment config):

```bash
TRACEKIT_API_KEY=ctxio_your_api_key_here
TRACEKIT_SERVICE_NAME=my-rails-app
TRACEKIT_ENVIRONMENT=production
TRACEKIT_CODE_MONITORING=true
```

The TraceKit Railtie automatically:
- Loads configuration from ENV variables
- Initializes OpenTelemetry with OTLP exporters
- Adds Rack middleware for request instrumentation
- Sets up graceful shutdown

**2. (Optional) For advanced configuration**, create an initializer:

```ruby
# config/initializers/tracekit.rb
Tracekit::SDK.configure do |c|
  c.api_key              = ENV['TRACEKIT_API_KEY']
  c.service_name         = "my-rails-app"
  c.endpoint             = "https://app.tracekit.dev/v1/traces"
  c.enable_code_monitoring = true
end
```

> **Note:** Only use the initializer if you need configuration beyond what environment variables provide. The Railtie handles initialization automatically from ENV vars.

**That's it for Rails!** Your controllers, ActiveRecord queries, and HTTP calls are automatically traced.

### Branch B: Sinatra

Sinatra requires manual configuration:

```ruby
# app.rb
require 'sinatra'
require 'tracekit'

# Configure TraceKit
Tracekit.configure do |config|
  config.api_key = ENV['TRACEKIT_API_KEY']
  config.service_name = 'my-sinatra-app'
  config.endpoint = 'https://app.tracekit.dev/v1/traces'
  config.enable_code_monitoring = true
end

# Add TraceKit middleware
use Tracekit::Middleware

# Routes are automatically traced!
get '/api/users' do
  content_type :json
  { users: ['alice', 'bob'] }.to_json
end
```

**Order matters:** `Tracekit.configure` then `use Tracekit::Middleware` then route definitions.

For modular Sinatra apps:

```ruby
# config.ru
require './app'
require 'tracekit'

Tracekit.configure do |config|
  config.api_key = ENV['TRACEKIT_API_KEY']
  config.service_name = 'my-sinatra-app'
  config.endpoint = 'https://app.tracekit.dev/v1/traces'
  config.enable_code_monitoring = true
end

use Tracekit::Middleware
run MyApp
```

At application shutdown, flush pending traces:

```ruby
at_exit { Tracekit.shutdown }
```

## Step 4: Framework-Specific Usage

### Rails Controllers

Your controllers are automatically traced. Use the SDK for custom metrics and snapshots:

```ruby
class UsersController < ApplicationController
  def index
    # This action is automatically traced!
    @users = User.all
    render json: @users
  end

  def create
    sdk = Tracekit.sdk

    # Track metrics
    sdk.counter("user.created").add(1)

    # Capture snapshot with context
    sdk.capture_snapshot("user-create", {
      email: params[:email],
      name: params[:name]
    })

    @user = User.create!(user_params)
    render json: @user, status: :created
  end
end
```

### Sinatra Routes

Routes are automatically traced when `Tracekit::Middleware` is used:

```ruby
get '/api/users' do
  sdk = Tracekit.sdk

  # Track metrics
  sdk.counter("api.users.requests").add(1)

  content_type :json
  User.all.to_json
end
```

## Step 5: Error Capture

Capture errors explicitly in rescue blocks:

```ruby
begin
  result = some_operation
rescue => e
  Tracekit.capture_exception(e)
  # handle the error...
end
```

For adding context to traces, use snapshots:

```ruby
sdk = Tracekit.sdk

sdk.capture_snapshot("process-order", {
  order_id: order.id,
  user_id: current_user.id,
  total: order.total
})
```

## Step 6: Verification

After integrating, verify traces are flowing:

1. **Start your application** with `TRACEKIT_API_KEY` set in the environment.
   - Rails: `TRACEKIT_API_KEY=ctxio_... rails server`
   - Sinatra: `TRACEKIT_API_KEY=ctxio_... ruby app.rb`
2. **Hit your endpoints 3-5 times** — e.g., `curl http://localhost:3000/api/users`.
3. **Open** `https://app.tracekit.dev/traces`.
4. **Confirm** new spans and your service name appear within 30-60 seconds.

If traces do not appear, see Troubleshooting below.

## Troubleshooting

### Traces not appearing in dashboard

- **Check `TRACEKIT_API_KEY`:** Ensure the env var is set in the runtime environment. Print it: `puts ENV["TRACEKIT_API_KEY"]`.
- **Check outbound access:** Your service must reach `https://app.tracekit.dev/v1/traces`. Verify with: `curl -X POST https://app.tracekit.dev/v1/traces` (expect 401 — means the endpoint is reachable).
- **For Rails:** Verify the Railtie loaded by checking logs for TraceKit initialization messages at startup.

### Rails: Railtie not loading

Symptoms: No TraceKit initialization messages in Rails boot log.

Fix: Ensure `gem 'tracekit'` is in your `Gemfile` (not just in a group like `:development`). Run `bundle install`. The Railtie auto-loads when the gem is available.

### Sinatra: Middleware not added

Symptoms: Sinatra app starts but no traces appear.

Fix: Ensure `use Tracekit::Middleware` is called after `Tracekit.configure` and before route definitions. For modular apps, add it in `config.ru`.

### Rails: Double initialization

Symptoms: Duplicate traces or initialization warnings.

Fix: If using the Railtie (env var config), do **not** also call `Tracekit.configure` manually. Choose one approach.

### Missing environment variable

Symptoms: `nil` API key on startup, or traces are rejected by the backend.

Fix: Ensure `TRACEKIT_API_KEY` is set in your `.env` file (loaded via `dotenv-rails` gem), Docker Compose, or deployment config.

### Service name collisions

Symptoms: Traces appear under the wrong service in the dashboard.

Fix: Use a unique service name per deployed service. Set `TRACEKIT_SERVICE_NAME` env var or `config.service_name` in the configure block.

## Next Steps

Once your Ruby app is traced, consider:
- **Code Monitoring** — Set live breakpoints and capture snapshots in production without redeploying (already enabled via `enable_code_monitoring: true`)
- **Distributed Tracing** — Connect traces across multiple services for full request visibility
- **Frontend Observability** — Add `@tracekit/browser` to your frontend for end-to-end trace correlation

## References

- Ruby SDK docs: `https://app.tracekit.dev/docs/languages/ruby`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
