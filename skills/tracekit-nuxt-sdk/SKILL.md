---
name: tracekit-nuxt-sdk
description: Sets up TraceKit APM in Nuxt applications with auto-plugin registration, error handlers, and distributed tracing.
---

# TraceKit Nuxt SDK Setup

## Auth Bootstrap

Do not tell the user to go sign up, log in separately, or manually create an API key before setup. First check for existing TraceKit auth. If `~/.tracekitconfig` does not contain the production profile or `TRACEKIT_API_KEY` is missing, apply the `tracekit-auth` skill first. Use `./scripts/run-tracekit-auth.sh status` to check, then guide the user through the TraceKit email verification flow with `register` and `verify`. That flow signs the user into an existing account for that email or creates the account automatically, then saves the returned credentials for the rest of the setup.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to a Nuxt application
- Add observability, error tracking, or APM to a Nuxt project
- Instrument a Nuxt app with distributed tracing
- Set up error monitoring in a Nuxt 3 application
- Configure TraceKit in a Nuxt project with composables
- Debug production Nuxt applications with live breakpoints

If the user has a vanilla JavaScript/TypeScript project without Nuxt, use the `tracekit-browser-sdk` skill instead. If the user has a plain Vue SPA (no Nuxt), use the `tracekit-vue-sdk` skill instead. If the user has Nuxt 2, refer them to the `tracekit-vue-sdk` skill for Vue 2 patterns.

## Non-Negotiable Rules

1. **Never hardcode API keys** in code. Always use `runtimeConfig` with `TRACEKIT_API_KEY` env var.
2. **Always include a verification step** confirming errors appear in `https://app.tracekit.dev`.
3. **Always enable code monitoring** (`enableCodeMonitoring: true`) -- it is TraceKit's differentiator.
4. **Always initialize TraceKit before app bootstrap** -- use the `.client.ts` plugin suffix to ensure browser-only init.

## Detection

Before applying this skill, detect the project type:

1. **Check `package.json`** for `nuxt` in dependencies -- confirms this is a Nuxt project.
2. **Check Nuxt version:** This skill targets Nuxt 3. For Nuxt 2, recommend the `tracekit-vue-sdk` skill with Vue 2 patterns.
3. **Confirm Nuxt 3** by checking for `nuxt.config.ts` (TypeScript) or `nuxt.config.js` at the project root.

## Step 1: Environment Setup

Set the `TRACEKIT_API_KEY` environment variable. Nuxt uses `.env` files and `runtimeConfig`.

Add to `.env`:

```bash
TRACEKIT_API_KEY=ctxio_your_api_key_here
```

Configure `runtimeConfig` in `nuxt.config.ts` to expose the key:

```typescript
// nuxt.config.ts
export default defineNuxtConfig({
  runtimeConfig: {
    // Server-side only
    tracekitApiKey: process.env.TRACEKIT_API_KEY || '',
    // Client-side (exposed to browser)
    public: {
      tracekitApiKey: process.env.TRACEKIT_API_KEY || '',
    },
  },
});
```

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. Use `.env` files, CI secrets, or deployment secret managers.

## Step 2: Install SDK

```bash
npm install @tracekit/nuxt
```

Or with Yarn:

```bash
yarn add @tracekit/nuxt
```

This installs the TraceKit Nuxt wrapper with the `@tracekit/browser` SDK, Nuxt plugin factory, composables, Vue error handler integration, and router breadcrumbs.

## Step 3: Module Registration

Register `@tracekit/nuxt` as a Nuxt module in `nuxt.config.ts`:

```typescript
// nuxt.config.ts
export default defineNuxtConfig({
  modules: ['@tracekit/nuxt'],

  tracekit: {
    // Module-level config (used by the Nuxt module for build-time setup)
    enabled: true,
  },

  runtimeConfig: {
    public: {
      tracekitApiKey: process.env.TRACEKIT_API_KEY || '',
    },
  },
});
```

The module handles auto-registration of the client plugin and provides the `useTraceKit()` composable as an auto-import.

## Step 4: Client Plugin

The module auto-registers a client plugin, but you can also create it manually for full control. The plugin file **must** end in `.client.ts` to ensure it only runs in the browser.

```typescript
// plugins/tracekit.client.ts
import { defineNuxtPlugin, useRuntimeConfig, useRouter } from '#imports';
import { createTraceKitPlugin, setupRouterBreadcrumbs } from '@tracekit/nuxt';

export default defineNuxtPlugin((nuxtApp) => {
  const config = useRuntimeConfig();

  // Initialize TraceKit (client-side only due to .client.ts suffix)
  const plugin = createTraceKitPlugin({
    apiKey: config.public.tracekitApiKey,
    release: '1.0.0',
    environment: process.env.NODE_ENV || 'production',
    endpoint: 'https://app.tracekit.dev/v1/traces',
    enableCodeMonitoring: true,
  });

  // Run the plugin initialization
  plugin(nuxtApp);

  // Set up router breadcrumbs
  const router = useRouter();
  setupRouterBreadcrumbs(router);
});
```

`createTraceKitPlugin()` creates a Nuxt plugin function that:
- Initializes the `@tracekit/browser` SDK
- Hooks into `vue:error` for component error capture
- Provides composable functions via `useNuxtApp().$tracekit`

Without the `.client.ts` suffix, Nuxt would also execute the plugin on the server where browser APIs are unavailable.

## Step 5: Error Handler

TraceKit hooks into Nuxt's error handling automatically via the `vue:error` hook in the client plugin. This captures:
- Vue component rendering errors
- Lifecycle hook errors
- Event handler errors

For additional error handling, integrate with Nuxt's `useError()` composable:

```vue
<!-- error.vue (Nuxt error page) -->
<script setup lang="ts">
import { captureException } from '@tracekit/nuxt';

const error = useError();

// Capture the error in TraceKit
if (error.value) {
  captureException(new Error(error.value.message), {
    statusCode: error.value.statusCode,
    statusMessage: error.value.statusMessage,
  });
}

const handleClear = () => clearError({ redirect: '/' });
</script>

<template>
  <div>
    <h1>{{ error?.statusCode }} - {{ error?.statusMessage }}</h1>
    <button @click="handleClear">Go Home</button>
  </div>
</template>
```

For inline error handling in components:

```vue
<script setup lang="ts">
import { useTraceKit } from '@tracekit/nuxt';

const { captureException, setUser } = useTraceKit();

async function loadData() {
  try {
    const data = await $fetch('/api/dashboard');
    // process data...
  } catch (err) {
    captureException(err as Error, { component: 'Dashboard' });
  }
}
</script>
```

## Step 6: Router Integration

Navigation breadcrumbs are captured automatically when `setupRouterBreadcrumbs()` is called in the plugin (Step 4). Nuxt uses Vue Router under the hood.

The integration uses `router.afterEach` to capture `NavigationEnd` events with from/to paths.

To disable parameterized routes (use actual URLs instead of route patterns):

```typescript
setupRouterBreadcrumbs(router, false);
```

If using the module registration (Step 3), router breadcrumbs are set up automatically.

## Step 7: Custom Performance Spans and Composable

The `useTraceKit()` composable provides TraceKit functions inside any Nuxt component:

```vue
<script setup lang="ts">
import { useTraceKit } from '@tracekit/nuxt';

const { captureException, captureMessage, setUser } = useTraceKit();

// Set user context after authentication
onMounted(async () => {
  const { data: user } = await useFetch('/api/auth/me');
  if (user.value) {
    setUser({ id: user.value.id, email: user.value.email });
  }
});

// Capture a custom message
function onCheckout() {
  captureMessage('User started checkout', 'info');
}
</script>
```

**Re-exported functions** available from `@tracekit/nuxt`:

```typescript
import {
  captureException,
  captureMessage,
  setUser,
  setTag,
  setExtra,
  addBreadcrumb,
  getClient,
} from '@tracekit/nuxt';
```

## Step 8: Distributed Tracing

TraceKit instruments `$fetch` and `useFetch` to propagate trace headers to your backend APIs. Configure `tracePropagationTargets` in the plugin init:

```typescript
createTraceKitPlugin({
  apiKey: config.public.tracekitApiKey,
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  tracePropagationTargets: [
    'https://api.example.com',
    /^\/api\//,  // Same-origin API routes
  ],
});
```

Outgoing `fetch` requests matching the targets automatically receive trace headers (`tracekit-trace-id`, `baggage`) for end-to-end trace correlation with backend services.

## Step 9: Session Replay (Optional)

Enable session replay in the client plugin config (browser only):

```typescript
createTraceKitPlugin({
  apiKey: config.public.tracekitApiKey,
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  replay: {
    enabled: true,
    sampleRate: 0.1,        // Record 10% of sessions
    errorSampleRate: 1.0,    // Record 100% of sessions with errors
    maskAllText: true,       // Mask sensitive text by default
    blockAllMedia: false,
  },
});
```

Session replay captures DOM mutations, network requests, and console logs in the browser. It does not affect server-side rendering.

## Step 10: Source Maps (Optional)

Upload source maps for readable stack traces in production errors.

After building with `nuxt build`, upload source maps:

```bash
tracekit sourcemaps upload \
  --api-key $TRACEKIT_API_KEY \
  --release 1.0.0 \
  --dist .output/public
```

Add this command to your CI/CD pipeline after `nuxt build`.

To enable source maps in the build:

```typescript
// nuxt.config.ts
export default defineNuxtConfig({
  sourcemap: {
    server: true,
    client: true,
  },
});
```

## Step 11: Verification

After integrating, verify errors are captured:

1. **Start your application** with `nuxt dev` or deploy to a test environment.
2. **Trigger a test error** -- add a temporary button that throws:
   ```vue
   <script setup lang="ts">
   function throwTestError() {
     throw new Error('TraceKit Nuxt test error');
   }
   </script>

   <template>
     <button @click="throwTestError">Test Error</button>
   </template>
   ```
3. **Open** `https://app.tracekit.dev`.
4. **Confirm** the test error appears within 30-60 seconds with component context.

If errors do not appear, see Troubleshooting below.

## Complete Working Example

All files needed for a full Nuxt 3 setup:

```bash
# .env
TRACEKIT_API_KEY=ctxio_your_api_key_here
```

```typescript
// nuxt.config.ts
export default defineNuxtConfig({
  modules: ['@tracekit/nuxt'],

  tracekit: {
    enabled: true,
  },

  runtimeConfig: {
    public: {
      tracekitApiKey: process.env.TRACEKIT_API_KEY || '',
    },
  },

  sourcemap: {
    server: true,
    client: true,
  },
});
```

```typescript
// plugins/tracekit.client.ts
import { defineNuxtPlugin, useRuntimeConfig, useRouter } from '#imports';
import { createTraceKitPlugin, setupRouterBreadcrumbs } from '@tracekit/nuxt';

export default defineNuxtPlugin((nuxtApp) => {
  const config = useRuntimeConfig();
  const router = useRouter();

  const plugin = createTraceKitPlugin({
    apiKey: config.public.tracekitApiKey,
    release: '1.0.0',
    environment: process.env.NODE_ENV || 'production',
    endpoint: 'https://app.tracekit.dev/v1/traces',
    enableCodeMonitoring: true,
    tracePropagationTargets: ['/api/'],
    replay: {
      enabled: true,
      sampleRate: 0.1,
      errorSampleRate: 1.0,
    },
  });

  plugin(nuxtApp);
  setupRouterBreadcrumbs(router);
});
```

```vue
<!-- error.vue -->
<script setup lang="ts">
import { captureException } from '@tracekit/nuxt';

const error = useError();

if (error.value) {
  captureException(new Error(error.value.message), {
    statusCode: error.value.statusCode,
  });
}

const handleClear = () => clearError({ redirect: '/' });
</script>

<template>
  <div>
    <h1>{{ error?.statusCode }} - {{ error?.statusMessage }}</h1>
    <p>An unexpected error occurred.</p>
    <button @click="handleClear">Go Home</button>
  </div>
</template>
```

```vue
<!-- pages/index.vue -->
<script setup lang="ts">
import { useTraceKit } from '@tracekit/nuxt';

const { setUser } = useTraceKit();

onMounted(async () => {
  const { data: user } = await useFetch('/api/auth/me');
  if (user.value) {
    setUser({ id: user.value.id, email: user.value.email });
  }
});
</script>

<template>
  <div>
    <h1>Welcome to My App</h1>
    <NuxtPage />
  </div>
</template>
```

## Troubleshooting

### SSR vs client init

- **Client-only:** The `.client.ts` plugin suffix ensures TraceKit only initializes in the browser. Without it, Nuxt runs the plugin on the server where `window`, `document`, and other browser APIs are unavailable.
- **If you see "window is not defined":** Check that your plugin file ends in `.client.ts`, not just `.ts`.

### Module not loaded

- Ensure `'@tracekit/nuxt'` is in the `modules` array (not `buildModules`).
- Check `nuxt.config.ts` for syntax errors.
- Run `nuxt prepare` to regenerate auto-imports after adding the module.

### Nitro server tracing

- The `@tracekit/nuxt` module focuses on client-side (browser) error capture. For Nitro server API routes, use the `tracekit-node-sdk` skill to instrument the server runtime.
- Server errors in Nuxt API routes (`server/api/*.ts`) are Node.js errors -- use Node.js patterns for capture.

### Hydration mismatches

- TraceKit captures hydration mismatch warnings as breadcrumbs when they occur in the browser.
- If you see hydration errors after adding TraceKit, they are pre-existing issues surfaced by the error handler -- not caused by TraceKit.

### Composable not available

- Ensure the `@tracekit/nuxt` module is registered in `nuxt.config.ts`.
- Run `nuxt prepare` to regenerate type definitions and auto-imports.
- `useTraceKit()` is only available in component setup functions and Nuxt lifecycle hooks.

### Errors not appearing in dashboard

- **Check API key:** Ensure `runtimeConfig.public.tracekitApiKey` is set. Print it: `console.log(useRuntimeConfig().public.tracekitApiKey)`.
- **Check outbound access:** Your app must reach `https://app.tracekit.dev/v1/traces`. Test with: `curl -X POST https://app.tracekit.dev/v1/traces` (expect 401).
- **Check plugin execution:** Add `console.log('TraceKit plugin loaded')` at the top of your `.client.ts` plugin to confirm it runs.

## Next Steps

Once your Nuxt app is traced, consider:
- **Browser SDK** -- For non-Nuxt pages, use the `tracekit-browser-sdk` skill
- **Vue SDK** -- For plain Vue SPAs, use the `tracekit-vue-sdk` skill
- **Session Replay** -- Record and replay user sessions with linked traces
- **Source Maps** -- Upload source maps for readable production stack traces
- **Backend SDKs** -- Connect frontend traces to backend services for full distributed tracing

## References

- Nuxt SDK docs: `https://app.tracekit.dev/docs/frontend/frameworks/nuxt`
- Browser SDK docs: `https://app.tracekit.dev/docs/frontend/browser-sdk`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
