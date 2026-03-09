---
name: tracekit-python-sdk
description: Sets up TraceKit APM in Python applications for automatic distributed tracing, error capture, and code monitoring. Supports Django, Flask, and FastAPI frameworks. Use when the user asks to add TraceKit, add observability, instrument a Python app, or configure APM in a Python project.
---

# TraceKit Python SDK Setup

> **Coming soon -- SDK in development.** The TraceKit Python SDK (`tracekit-apm`) is not yet published. The setup patterns below reflect the planned API. This skill will be updated when the SDK ships.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to a Python application
- Add observability or APM to a Python service
- Instrument a Django, Flask, or FastAPI app with distributed tracing
- Configure TraceKit API keys in a Python project
- Debug production Python services with live breakpoints
- Set up code monitoring in a Python app

## Non-Negotiable Rules

1. **Never hardcode API keys** in code. Always use `os.getenv("TRACEKIT_API_KEY")` or `os.environ["TRACEKIT_API_KEY"]`.
2. **Always call `tracekit.init()` before creating the app** — initialization must happen before framework setup so auto-instrumentation patches are applied.
3. **Always include a verification step** confirming traces appear in `https://app.tracekit.dev/traces`.
4. **Always enable code monitoring** (`enable_code_monitoring=True`) — it is TraceKit's differentiator.

## Detection

Before applying this skill, detect the project type:

1. **Check for Python project files** — `requirements.txt`, `pyproject.toml`, or `Pipfile` confirms this is a Python project.
2. **Detect framework** by scanning dependencies or imports:
   - `django` in dependencies/imports => Django framework (use Django branch)
   - `flask` in dependencies/imports => Flask framework (use Flask branch)
   - `fastapi` in dependencies/imports => FastAPI framework (use FastAPI branch)
3. **Only ask the user** if multiple frameworks are detected or if no project file is found.

## Step 1: Environment Setup

Set the `TRACEKIT_API_KEY` environment variable. This is the only required secret.

Add to your `.env` file or environment:

```bash
export TRACEKIT_API_KEY=ctxio_your_api_key_here
```

The OTLP endpoint is hardcoded in the SDK init — no need to configure it separately.

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. Use `.env` files, deployment secret managers, or CI variables.

## Step 2: Install SDK

```bash
pip install tracekit-apm
```

Framework-specific extras are available:

```bash
pip install tracekit-apm[flask]    # For Flask
pip install tracekit-apm[django]   # For Django
pip install tracekit-apm[fastapi]  # For FastAPI
pip install tracekit-apm[all]      # All frameworks
```

This installs the TraceKit Python SDK with built-in OpenTelemetry support, framework middleware, and code monitoring.

## Step 3: Initialize TraceKit

Add initialization at the top of your application entry point, **before** creating the framework app:

```python
import tracekit
import os

# Initialize TraceKit — MUST be before app creation
client = tracekit.init(
    api_key=os.getenv("TRACEKIT_API_KEY"),
    service_name="my-python-service",
    endpoint="https://app.tracekit.dev/v1/traces",
    enable_code_monitoring=True,
)
```

**Key points:**
- `service_name` should match your service's logical name (e.g., `"api-gateway"`, `"user-service"`)
- `enable_code_monitoring=True` enables live breakpoints and snapshots in production
- Call `tracekit.init()` before `Flask(__name__)`, `FastAPI()`, or Django's `MIDDLEWARE` setup

## Step 4: Framework Integration

Choose the branch matching your framework. Apply **one** of the following.

### Branch A: Flask

```python
# app.py
from flask import Flask, request, jsonify
import tracekit
from tracekit.middleware.flask import init_flask_app
import os

# Initialize TraceKit (before app creation!)
client = tracekit.init(
    api_key=os.getenv("TRACEKIT_API_KEY"),
    service_name="my-flask-app",
    endpoint="https://app.tracekit.dev/v1/traces",
    enable_code_monitoring=True,
)

# Create Flask app
app = Flask(__name__)

# Add TraceKit middleware (auto-traces all routes!)
init_flask_app(app, client)

@app.route("/api/users/<int:user_id>")
def get_user(user_id):
    return jsonify({"id": user_id, "name": "Alice"})

if __name__ == "__main__":
    app.run(port=5000)
```

**Order matters:** `tracekit.init()` then `Flask(__name__)` then `init_flask_app(app, client)` then route definitions.

### Branch B: Django

Add TraceKit initialization to your Django settings and middleware.

**1. Initialize in `settings.py`:**

```python
# settings.py
import tracekit
import os

# Initialize TraceKit at Django startup
client = tracekit.init(
    api_key=os.getenv("TRACEKIT_API_KEY"),
    service_name="my-django-app",
    endpoint="https://app.tracekit.dev/v1/traces",
    enable_code_monitoring=True,
)
```

**2. Add middleware to `MIDDLEWARE`:**

```python
# settings.py
MIDDLEWARE = [
    'tracekit.middleware.django.TracekitDjangoMiddleware',
    # ... other middleware (keep TraceKit first for complete request tracing)
    'django.middleware.security.SecurityMiddleware',
    'django.contrib.sessions.middleware.SessionMiddleware',
    'django.middleware.common.CommonMiddleware',
    # ...
]
```

**Order matters:** Place `TracekitDjangoMiddleware` at the top of `MIDDLEWARE` to trace the full request lifecycle.

### Branch C: FastAPI

```python
# main.py
from fastapi import FastAPI
import tracekit
from tracekit.middleware.fastapi import init_fastapi_app
import os

# Initialize TraceKit (before app creation!)
client = tracekit.init(
    api_key=os.getenv("TRACEKIT_API_KEY"),
    service_name="my-fastapi-app",
    endpoint="https://app.tracekit.dev/v1/traces",
    enable_code_monitoring=True,
)

# Create FastAPI app
app = FastAPI()

# Add TraceKit middleware (auto-traces all routes!)
init_fastapi_app(app, client)

@app.get("/api/users/{user_id}")
async def get_user(user_id: int):
    return {"id": user_id, "name": "Alice"}
```

**Order matters:** `tracekit.init()` then `FastAPI()` then `init_fastapi_app(app, client)` then route definitions.

## Step 5: Error Capture

Capture errors explicitly in except blocks:

```python
try:
    result = some_operation()
except Exception as e:
    tracekit.capture_exception(e)
    # handle the error...
```

## Step 5b: Snapshot Capture (Code Monitoring)

For programmatic snapshots, **use the snapshot client directly** — do not call through the SDK wrapper. The SDK uses stack inspection internally to identify the call site. Adding extra layers shifts the frame and causes snapshots to report the wrong source location.

Create a thin wrapper module (e.g., `app/breakpoints.py`):

```python
_snapshot_client = None


def init(sdk):
    """Store the snapshot client. No-op when sdk is None."""
    global _snapshot_client
    if sdk is not None:
        _snapshot_client = sdk.snapshot_client()


def capture(name: str, data: dict):
    """Fire a code monitoring snapshot. No-op when tracing is disabled."""
    if _snapshot_client is None:
        return
    _snapshot_client.check_and_capture(name, data)
```

Initialize after SDK setup:

```python
from app.breakpoints import init as init_breakpoints
init_breakpoints(sdk)
```

Use at call sites:

```python
from app.breakpoints import capture

capture("payment-failed", {"order_id": order.id, "error": str(e)})
```

See the `tracekit-code-monitoring` skill for the full pattern across all languages.

## Step 6: Verification

After integrating, verify traces are flowing:

1. **Start your application** with `TRACEKIT_API_KEY` set in the environment.
2. **Hit your endpoints 3-5 times** — e.g., `curl http://localhost:5000/api/users/1` (Flask) or `curl http://localhost:8000/api/users/1` (FastAPI/Django).
3. **Open** `https://app.tracekit.dev/traces`.
4. **Confirm** new spans and your service name appear within 30-60 seconds.

If traces do not appear, see Troubleshooting below.

## Troubleshooting

### Traces not appearing in dashboard

- **Check `TRACEKIT_API_KEY`:** Ensure the env var is set in the runtime environment. Print it: `print(os.getenv("TRACEKIT_API_KEY"))`.
- **Check outbound access:** Your service must reach `https://app.tracekit.dev/v1/traces`. Verify with: `curl -X POST https://app.tracekit.dev/v1/traces` (expect 401 — means the endpoint is reachable).
- **Check init order:** `tracekit.init()` must be called **before** creating the Flask/FastAPI app or before Django processes the first request. If init happens too late, auto-instrumentation patches miss early imports.

### Init order wrong

Symptoms: Server starts fine but no traces appear despite traffic.

Fix: Move `tracekit.init()` to the very top of your entry module, before importing route modules or creating the app object.

```python
# DO: Init before app creation
tracekit.init()
app = Flask(__name__)

# DON'T: Import route modules before tracekit.init()
```

### Missing environment variable

Symptoms: `None` API key warning on startup, or traces are rejected by the backend.

Fix: Ensure `TRACEKIT_API_KEY` is set in your `.env` file (loaded via `python-dotenv`), virtualenv activation script, Docker Compose, or deployment config.

### Django middleware order

Symptoms: Partial traces or missing request context.

Fix: Place `'tracekit.middleware.django.TracekitDjangoMiddleware'` as the **first** entry in `MIDDLEWARE`.

### Service name collisions

Symptoms: Traces appear under the wrong service in the dashboard.

Fix: Use a unique `service_name` per deployed service. Avoid generic names like `"app"` or `"server"`.

## Next Steps

Once your Python app is traced, consider:
- **Code Monitoring** — Set live breakpoints and capture snapshots in production without redeploying (already enabled via `enable_code_monitoring=True`)
- **Distributed Tracing** — Connect traces across multiple services for full request visibility
- **Frontend Observability** — Add `@tracekit/browser` to your frontend for end-to-end trace correlation

## References

- Python SDK docs: `https://app.tracekit.dev/docs/languages/python`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
