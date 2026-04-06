---
name: tracekit-node-sdk
description: Sets up TraceKit APM in Node.js applications for automatic distributed tracing, error capture, and code monitoring. Supports Express, Fastify, and NestJS frameworks. Use when the user asks to add TraceKit, add observability, instrument a Node.js app, or configure APM in a Node.js/TypeScript project.
---

# TraceKit Node.js SDK Setup

## Auth Bootstrap

Do not tell the user to go sign up, log in separately, or manually create an API key before setup. First check for existing TraceKit auth. If `~/.tracekitconfig` does not contain the production profile or `TRACEKIT_API_KEY` is missing, apply the `tracekit-auth` skill first. Use `./scripts/run-tracekit-auth.sh status` to check, then guide the user through the TraceKit email verification flow with `register` and `verify`. That flow signs the user into an existing account for that email or creates the account automatically, then saves the returned credentials for the rest of the setup.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to a Node.js or TypeScript application
- Add observability or APM to a Node.js service
- Instrument an Express, Fastify, or NestJS app with distributed tracing
- Configure TraceKit API keys in a Node.js project
- Debug production Node.js services with live breakpoints
- Set up code monitoring in a Node.js app

## Non-Negotiable Rules

1. **Never hardcode API keys** in code. Always use `process.env.TRACEKIT_API_KEY`.
2. **Always call `tracekit.init()` before registering routes**  - middleware and route handlers must come after initialization.
3. **Always include a verification step** confirming traces appear in `https://app.tracekit.dev/traces`.
4. **Always enable code monitoring** (`enableCodeMonitoring: true`)  - it is TraceKit's differentiator.

## Detection

Before applying this skill, detect the project type:

1. **Check for `package.json`**  - confirms this is a Node.js project.
2. **Detect framework** by scanning `package.json` dependencies:
   - `"express"` in dependencies => Express framework (use Express branch)
   - `"fastify"` in dependencies => Fastify framework (use Fastify branch)
   - `"@nestjs/core"` in dependencies => NestJS framework (use NestJS branch)
3. **Check for TypeScript:** Look for `tsconfig.json` or `"typescript"` in devDependencies. Use TypeScript snippets if present.
4. **Only ask the user** if multiple frameworks are detected or if `package.json` is missing.

## Step 1: Environment Setup

Set the `TRACEKIT_API_KEY` environment variable. This is the only required secret.

Add to your `.env` file:

```bash
TRACEKIT_API_KEY=ctxio_your_api_key_here
```

The OTLP endpoint is hardcoded in the SDK init  - no need to configure it separately.

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. Use `.env` files, deployment secret managers, or CI variables.

## Step 2: Install SDK

```bash
npm install @tracekit/node-apm
```

Or with Yarn:

```bash
yarn add @tracekit/node-apm
```

This installs the TraceKit Node.js APM package with built-in OpenTelemetry support, framework middleware, and code monitoring.

## Step 3: Initialize TraceKit

Add initialization to your application entry point, **before** any route or middleware registration.

### TypeScript

```typescript
import * as tracekit from '@tracekit/node-apm';

// Initialize TraceKit  - MUST be before routes
tracekit.init({
  apiKey: process.env.TRACEKIT_API_KEY!,
  serviceName: 'my-node-service',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
});
```

### JavaScript (CommonJS)

```javascript
const tracekit = require('@tracekit/node-apm');

// Initialize TraceKit  - MUST be before routes
tracekit.init({
  apiKey: process.env.TRACEKIT_API_KEY,
  serviceName: 'my-node-service',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
});
```

**Key points:**
- `serviceName` should match your service's logical name (e.g., `"api-gateway"`, `"user-service"`)
- `enableCodeMonitoring: true` enables live breakpoints and snapshots in production
- Call `tracekit.init()` at the very top of your entry file, before importing route modules when possible

## Step 4: Framework Integration

Choose the branch matching your framework. Apply **one** of the following.

### Branch A: Express

```typescript
import express from 'express';
import * as tracekit from '@tracekit/node-apm';

// Initialize TraceKit (before routes!)
tracekit.init({
  apiKey: process.env.TRACEKIT_API_KEY!,
  serviceName: 'my-express-app',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
});

const app = express();

// Add TraceKit middleware (before routes!)
app.use(tracekit.middleware());

// Your routes  - automatically traced
app.get('/api/users', (req, res) => {
  res.json({ users: ['alice', 'bob'] });
});

app.listen(3000, () => {
  console.log('Server running on port 3000');
});
```

**Order matters:** `tracekit.init()` then `tracekit.middleware()` then route definitions.

### Branch B: Fastify

```typescript
import Fastify from 'fastify';
import * as tracekit from '@tracekit/node-apm';

// Initialize TraceKit (before routes!)
tracekit.init({
  apiKey: process.env.TRACEKIT_API_KEY!,
  serviceName: 'my-fastify-app',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
});

const fastify = Fastify();

// Register TraceKit plugin (before routes!)
fastify.register(tracekit.fastifyPlugin());

// Your routes  - automatically traced
fastify.get('/api/users', async (request, reply) => {
  return { users: ['alice', 'bob'] };
});

fastify.listen({ port: 3000 });
```

**Order matters:** `tracekit.init()` then `fastify.register(tracekit.fastifyPlugin())` then route definitions.

### Branch C: NestJS

NestJS uses a module-based approach. Import `TracekitModule` in your root `AppModule`:

```typescript
// app.module.ts
import { Module } from '@nestjs/common';
import { TracekitModule } from '@tracekit/node-apm/nestjs';
import { UsersModule } from './users/users.module';

@Module({
  imports: [
    TracekitModule.forRoot({
      apiKey: process.env.TRACEKIT_API_KEY!,
      serviceName: 'my-nestjs-app',
      endpoint: 'https://app.tracekit.dev/v1/traces',
      enableCodeMonitoring: true,
    }),
    UsersModule,
  ],
})
export class AppModule {}
```

The `TracekitModule` automatically registers a global interceptor that traces all HTTP requests. No additional middleware needed.

**Async configuration** (using `ConfigService`):

```typescript
// app.module.ts
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { TracekitModule } from '@tracekit/node-apm/nestjs';

@Module({
  imports: [
    ConfigModule.forRoot(),
    TracekitModule.forRootAsync({
      inject: [ConfigService],
      useFactory: (config: ConfigService) => ({
        apiKey: config.get('TRACEKIT_API_KEY')!,
        serviceName: config.get('APP_NAME', 'my-app'),
        endpoint: 'https://app.tracekit.dev/v1/traces',
        enableCodeMonitoring: true,
      }),
    }),
  ],
})
export class AppModule {}
```

## Step 5: Error Capture

Capture errors explicitly in catch blocks:

```typescript
import { getClient } from '@tracekit/node-apm';

try {
  await someOperation();
} catch (err) {
  const client = getClient();
  client.captureException(err as Error);
  // handle the error...
}
```

For adding context to traces, use manual spans:

```typescript
import { getClient } from '@tracekit/node-apm';

app.post('/api/orders', async (req, res) => {
  const client = getClient();

  const span = client.startSpan('process-order', null, {
    'order.id': req.body.orderId,
    'user.id': req.user?.id,
  });

  try {
    const result = await processOrder(req.body);
    span.end();
    res.json(result);
  } catch (err) {
    client.captureException(err as Error);
    span.end();
    res.status(500).json({ error: 'Processing failed' });
  }
});
```

## Step 5b: Snapshot Capture (Code Monitoring)

For programmatic snapshots, **use the SnapshotClient directly**  - do not call through the SDK wrapper. The SDK uses stack inspection internally to identify the call site. Adding extra layers shifts the frame and causes snapshots to report the wrong source location.

Create a thin wrapper module (e.g., `src/lib/breakpoints.ts`):

```typescript
import * as tracekit from '@tracekit/node-apm';

let snapshotClient: tracekit.SnapshotClient | null = null;

export function init(sdk: tracekit.SDK): void {
  snapshotClient = sdk.snapshotClient();
}

export function capture(name: string, data: Record<string, unknown>): void {
  if (!snapshotClient) return;
  snapshotClient.checkAndCapture(name, data);
}
```

Initialize after SDK setup:

```typescript
import * as breakpoints from './lib/breakpoints';
breakpoints.init(sdk);
```

Use at call sites:

```typescript
import { capture } from './lib/breakpoints';

capture('payment-failed', { orderId: order.id, error: String(err) });
```

See the `tracekit-code-monitoring` skill for the full pattern across all languages.

## Step 6: Verification

After integrating, verify traces are flowing:

1. **Start your application** with `TRACEKIT_API_KEY` set in the environment.
2. **Hit your endpoints 3-5 times**  - e.g., `curl http://localhost:3000/api/users`.
3. **Open** `https://app.tracekit.dev/traces`.
4. **Confirm** new spans and your service name appear within 30-60 seconds.

If traces do not appear, see Troubleshooting below.

## Troubleshooting

### Traces not appearing in dashboard

- **Check `TRACEKIT_API_KEY`:** Ensure the env var is set in the runtime environment. Print it: `console.log(process.env.TRACEKIT_API_KEY)`.
- **Check outbound access:** Your service must reach `https://app.tracekit.dev/v1/traces`. Verify with: `curl -X POST https://app.tracekit.dev/v1/traces` (expect 401  - means the endpoint is reachable).
- **Check init order:** `tracekit.init()` must be called **before** registering routes and middleware. If init happens after routes, requests are not traced.

### Init order wrong

Symptoms: Server starts fine but no traces appear despite traffic.

Fix: Move `tracekit.init()` to the very top of your entry file (`src/index.ts`, `src/server.ts`, or `main.ts`), before importing route modules or creating the Express/Fastify app.

### Missing environment variable

Symptoms: `undefined` API key warning on startup, or traces are rejected by the backend.

Fix: Ensure `TRACEKIT_API_KEY` is set in your `.env` file and loaded (e.g., via `dotenv`), Docker Compose, or deployment config.

### NestJS module not registered

Symptoms: NestJS app starts but no traces appear.

Fix: Ensure `TracekitModule.forRoot(...)` is in the `imports` array of your root `AppModule`, not a feature module.

### Service name collisions

Symptoms: Traces appear under the wrong service in the dashboard.

Fix: Use a unique `serviceName` per deployed service. Avoid generic names like `"app"` or `"server"`.

## Next Steps

Once your Node.js app is traced, consider:
- **Code Monitoring**  - Set live breakpoints and capture snapshots in production without redeploying (already enabled via `enableCodeMonitoring: true`)
- **Distributed Tracing**  - Connect traces across multiple services for full request visibility
- **Frontend Observability**  - Add `@tracekit/browser` to your frontend for end-to-end trace correlation

## References

- Node.js SDK docs: `https://app.tracekit.dev/docs/languages/nodejs`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
