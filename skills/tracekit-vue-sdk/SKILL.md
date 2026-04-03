---
name: tracekit-vue-sdk
description: Sets up TraceKit APM in Vue.js applications with error handlers, router integration, and distributed tracing. Supports both Vue 2 (Options API) and Vue 3 (Composition API). Use when the user asks to add TraceKit, add observability, instrument a Vue app, or configure APM in a Vue project.
---

# TraceKit Vue SDK Setup

## Auth Bootstrap

Do not tell the user to go sign up, log in separately, or manually create an API key before setup. First check for existing TraceKit auth. If `~/.tracekitconfig` does not contain the production profile or `TRACEKIT_API_KEY` is missing, apply the `tracekit-auth` skill first. Use `./scripts/run-tracekit-auth.sh status` to check, then guide the user through the TraceKit email verification flow with `register` and `verify`. That flow signs the user into an existing account for that email or creates the account automatically, then saves the returned credentials for the rest of the setup.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to a Vue.js application
- Add observability or APM to a Vue app
- Instrument a Vue project with error tracking or distributed tracing
- Configure TraceKit API keys in a Vue project
- Set up Vue error handlers with automatic error reporting
- Add performance monitoring to a Vue app
- Debug production Vue apps with live breakpoints

**Not Vue?** If the user is using React, Angular, Next.js, Nuxt, or a plain JS/TS project without a framework, use the corresponding skill instead. If the user is using Nuxt, use the `tracekit-nuxt-sdk` skill -- it provides SSR-aware initialization.

## Non-Negotiable Rules

1. **Never hardcode API keys** in code. Always use environment variables or build-time injection (e.g., `import.meta.env.VITE_TRACEKIT_API_KEY`).
2. **Always include a verification step** confirming errors and traces appear in `https://app.tracekit.dev/traces`.
3. **Always enable code monitoring** (`enableCodeMonitoring: true`) — it is TraceKit's differentiator for live debugging.
4. **Always init TraceKit before mounting the app** — the plugin must be registered with `app.use()` before `app.mount()`.

## Detection

Before applying this skill, detect the project type:

1. **Check for `package.json`** — confirms this is a JavaScript/TypeScript project.
2. **Check for `vue`** in `dependencies` — confirms this is a Vue project.
3. **Check for Nuxt:** If `nuxt` is in dependencies, use the `tracekit-nuxt-sdk` skill instead.
4. **Detect Vue version:**
   - Check `main.ts` or `main.js` for `createApp` usage => **Vue 3** (use Composition API examples)
   - Check for `new Vue` usage => **Vue 2** (use Options API examples)
   - Check `vue` version in `package.json`: `^3.x` = Vue 3, `^2.x` = Vue 2
5. **Check for TypeScript:** `tsconfig.json` presence means use `.ts` snippets.
6. **Only ask the user** if Vue version cannot be determined.

## Step 1: Environment Setup

Set the `TRACEKIT_API_KEY` environment variable. This is the only required secret.

Add to your `.env` file:

```bash
VITE_TRACEKIT_API_KEY=ctxio_your_api_key_here
```

**Vue CLI projects** (Webpack-based):
```bash
VUE_APP_TRACEKIT_API_KEY=ctxio_your_api_key_here
```

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. Use `.env` files, deployment secret managers, or CI variables.

## Step 2: Install SDK

```bash
npm install @tracekit/vue
```

Or with Yarn:

```bash
yarn add @tracekit/vue
```

This installs the TraceKit Vue SDK which wraps `@tracekit/browser` with Vue-specific integrations: plugin pattern, error handler chaining, router breadcrumbs, and Composition API composables. You only need this one package.

## Step 3: Initialize TraceKit Plugin

Register the `TraceKitPlugin` with your Vue app. The plugin handles SDK initialization, error handler setup, and optional router breadcrumbs in a single `app.use()` call.

### Vue 3 (Composition API)

```typescript
// src/main.ts
import { createApp } from 'vue';
import { createRouter, createWebHistory } from 'vue-router';
import { TraceKitPlugin } from '@tracekit/vue';
import App from './App.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: () => import('./views/Home.vue') },
    { path: '/dashboard', component: () => import('./views/Dashboard.vue') },
  ],
});

const app = createApp(App);

// Register TraceKit BEFORE mounting
app.use(TraceKitPlugin, {
  apiKey: import.meta.env.VITE_TRACEKIT_API_KEY,
  serviceName: 'my-vue-app',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  release: import.meta.env.VITE_APP_VERSION || '0.0.0',
  environment: import.meta.env.MODE,
  router, // enables automatic route change breadcrumbs
});

app.use(router);
app.mount('#app');
```

### Vue 2 (Options API)

```javascript
// src/main.js
import Vue from 'vue';
import VueRouter from 'vue-router';
import { TraceKitPlugin } from '@tracekit/vue';
import App from './App.vue';

Vue.use(VueRouter);

const router = new VueRouter({
  mode: 'history',
  routes: [
    { path: '/', component: () => import('./views/Home.vue') },
    { path: '/dashboard', component: () => import('./views/Dashboard.vue') },
  ],
});

// Register TraceKit plugin
Vue.use(TraceKitPlugin, {
  apiKey: process.env.VUE_APP_TRACEKIT_API_KEY,
  serviceName: 'my-vue-app',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  release: process.env.VUE_APP_VERSION || '0.0.0',
  environment: process.env.NODE_ENV,
  router, // enables automatic route change breadcrumbs
});

new Vue({
  router,
  render: (h) => h(App),
}).$mount('#app');
```

**Key points:**
- `TraceKitPlugin` must be registered before `app.mount()` (Vue 3) or `new Vue()` (Vue 2)
- The plugin calls `init()` internally — do not call `init()` separately
- Passing `router` is optional but recommended — it enables automatic navigation breadcrumbs
- `serviceName` should match your app's logical name

## Step 4: Error Handler

The `TraceKitPlugin` automatically hooks into Vue's error handling lifecycle. It intercepts errors via `app.config.errorHandler` (Vue 3) or `Vue.config.errorHandler` (Vue 2) and reports them to TraceKit with component context.

**What gets captured automatically:**
- Template rendering errors
- Lifecycle hook errors (`mounted`, `updated`, etc.)
- Watcher errors
- Component event handler errors (in Vue 3)

**Error handler chaining:** If you already have a custom `errorHandler`, the plugin chains with it — your handler runs first, then TraceKit captures the error. No errors are silently swallowed.

```typescript
// Vue 3 — manual error handler setup (if NOT using the plugin)
import { createApp } from 'vue';
import { setupErrorHandler, captureException } from '@tracekit/vue';

const app = createApp(App);
setupErrorHandler(app); // hooks into app.config.errorHandler
app.mount('#app');
```

```javascript
// Vue 2 — manual error handler setup (if NOT using the plugin)
import Vue from 'vue';
import { setupErrorHandler } from '@tracekit/vue';

setupErrorHandler(Vue); // hooks into Vue.config.errorHandler
```

**Capturing errors in event handlers and async code:**

For errors outside Vue's error handling lifecycle (e.g., in `@click` handlers or `async` methods), use `captureException`:

### Composition API (Vue 3)

```vue
<script setup lang="ts">
import { captureException } from '@tracekit/vue';

async function handleSubmit() {
  try {
    await submitForm();
  } catch (err) {
    captureException(err as Error, { component: 'ContactForm' });
  }
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <!-- form fields -->
    <button type="submit">Send</button>
  </form>
</template>
```

### Options API (Vue 2)

```vue
<script>
export default {
  methods: {
    async handleSubmit() {
      try {
        await this.submitForm();
      } catch (err) {
        this.$tracekit.captureException(err, { component: 'ContactForm' });
      }
    },
  },
};
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <!-- form fields -->
    <button type="submit">Send</button>
  </form>
</template>
```

## Step 5: Router Integration

When you pass a `router` instance to `TraceKitPlugin`, the plugin automatically captures navigation breadcrumbs. Every route change records:
- Previous route path
- New route path
- Navigation type (push, replace, back/forward)

```typescript
// Vue 3 — router is passed in plugin options
app.use(TraceKitPlugin, {
  apiKey: import.meta.env.VITE_TRACEKIT_API_KEY,
  serviceName: 'my-vue-app',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  router, // <-- enables navigation breadcrumbs
});
```

**Manual setup** (if not using the plugin):

```typescript
import { createRouter, createWebHistory } from 'vue-router';
import { setupRouterBreadcrumbs } from '@tracekit/vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [/* ... */],
});

setupRouterBreadcrumbs(router);
```

**Parameterized routes:** By default, route parameters are included in breadcrumbs (e.g., `/users/123`). To use parameterized paths instead (e.g., `/users/:id`), pass `parameterizedRoutes: true` in plugin options.

## Step 6: Custom Performance Spans

### Composition API (Vue 3)

Use the `useTraceKitSpan` composable to measure operations within components:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useTraceKitSpan, captureException } from '@tracekit/vue';

const data = ref(null);
const { startSpan } = useTraceKitSpan();

onMounted(async () => {
  const span = startSpan('load-dashboard-data', {
    'component': 'Dashboard',
  });
  try {
    const response = await fetch('/api/dashboard');
    data.value = await response.json();
    span.end();
  } catch (err) {
    captureException(err as Error);
    span.end();
  }
});
</script>
```

### Options API (Vue 2)

Use the `$tracekit` instance property:

```vue
<script>
export default {
  data() {
    return { data: null };
  },
  async mounted() {
    const span = this.$tracekit.startSpan('load-dashboard-data', {
      component: 'Dashboard',
    });
    try {
      const response = await fetch('/api/dashboard');
      this.data = await response.json();
      span.end();
    } catch (err) {
      this.$tracekit.captureException(err);
      span.end();
    }
  },
};
</script>
```

## Step 7: Distributed Tracing

Configure `tracePropagationTargets` in the plugin options to attach trace headers to outbound fetch/XHR requests:

```typescript
app.use(TraceKitPlugin, {
  apiKey: import.meta.env.VITE_TRACEKIT_API_KEY,
  serviceName: 'my-vue-app',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  tracePropagationTargets: [
    'https://api.myapp.com',
    'https://auth.myapp.com',
    /^https:\/\/.*\.myapp\.com/,
  ],
  router,
});
```

**How it works:**
1. Fetch/XHR requests to matching URLs receive a `traceparent` header
2. Your backend SDK reads this header and links the backend span to the frontend trace
3. The full request lifecycle appears as a single trace in the TraceKit dashboard

**Important:** Your backend CORS configuration must accept the `traceparent` and `tracestate` headers.

## Step 8: Session Replay (Optional)

Enable session replay via the plugin config to record user sessions for visual debugging:

```typescript
import { replayIntegration } from '@tracekit/replay';

const replay = replayIntegration({
  sessionSampleRate: 0.1,
  errorSampleRate: 1.0,
  maskAllText: true,
  blockAllMedia: true,
});

app.use(TraceKitPlugin, {
  apiKey: import.meta.env.VITE_TRACEKIT_API_KEY,
  serviceName: 'my-vue-app',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  addons: [replay],
  router,
});
```

## Step 9: Source Maps (Optional)

Upload source maps so stack traces show original file names and line numbers.

Add to your `.env`:

```bash
TRACEKIT_AUTH_TOKEN=your_auth_token_here
```

After building:

```bash
tracekit sourcemaps upload --release=1.0.0 ./dist
```

**Build integration** — add to `package.json`:

```json
{
  "scripts": {
    "build": "vite build",
    "postbuild": "tracekit sourcemaps upload --release=$npm_package_version ./dist"
  }
}
```

Ensure the `release` value matches the `release` option in your plugin config.

## Step 10: Verification

After integrating, verify errors and traces are flowing:

1. **Start your application** with the API key env var set.
2. **Trigger a test error** — add this temporarily in a component:

   **Vue 3:**
   ```vue
   <script setup lang="ts">
   import { onMounted } from 'vue';
   import { captureException } from '@tracekit/vue';

   onMounted(() => {
     captureException(new Error('TraceKit Vue test error'));
   });
   </script>
   ```

   **Vue 2:**
   ```vue
   <script>
   export default {
     mounted() {
       this.$tracekit.captureException(new Error('TraceKit Vue test error'));
     },
   };
   </script>
   ```

3. **Open** `https://app.tracekit.dev/traces`.
4. **Confirm** the test error and your service name appear within 30-60 seconds.
5. **Remove the test code** once verified.

## Complete Working Example

### Vue 3 (Composition API)

```typescript
// src/main.ts
import { createApp } from 'vue';
import { createRouter, createWebHistory } from 'vue-router';
import { TraceKitPlugin } from '@tracekit/vue';
import { replayIntegration } from '@tracekit/replay';
import App from './App.vue';

// --- Routes ---
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: () => import('./views/Home.vue') },
    { path: '/dashboard', component: () => import('./views/Dashboard.vue') },
    { path: '/users/:id', component: () => import('./views/UserProfile.vue') },
  ],
});

// --- Session Replay ---
const replay = replayIntegration({
  sessionSampleRate: 0.1,
  errorSampleRate: 1.0,
  maskAllText: true,
  blockAllMedia: true,
});

// --- App Init ---
const app = createApp(App);

app.use(TraceKitPlugin, {
  apiKey: import.meta.env.VITE_TRACEKIT_API_KEY,
  serviceName: 'my-vue-app',
  release: import.meta.env.VITE_APP_VERSION || '0.0.0',
  environment: import.meta.env.MODE,
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  tracePropagationTargets: [
    'https://api.myapp.com',
    /^https:\/\/.*\.myapp\.com/,
  ],
  addons: [replay],
  router,
});

app.use(router);
app.mount('#app');
```

```vue
<!-- src/views/Dashboard.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useTraceKitSpan, captureException, setUser, setTag } from '@tracekit/vue';

const data = ref(null);
const error = ref<string | null>(null);
const { startSpan } = useTraceKitSpan();

onMounted(async () => {
  // Set user context
  setUser({ id: 'user-123', email: 'alice@example.com' });
  setTag('plan', 'pro');

  // Measure data load
  const span = startSpan('load-dashboard');
  try {
    const response = await fetch('/api/dashboard');
    data.value = await response.json();
    span.end();
  } catch (err) {
    captureException(err as Error, { component: 'Dashboard' });
    error.value = 'Failed to load dashboard';
    span.end();
  }
});
</script>

<template>
  <div>
    <h1>Dashboard</h1>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-else-if="data">{{ data }}</div>
    <div v-else>Loading...</div>
  </div>
</template>
```

### Vue 2 (Options API)

```javascript
// src/main.js
import Vue from 'vue';
import VueRouter from 'vue-router';
import { TraceKitPlugin } from '@tracekit/vue';
import App from './App.vue';

Vue.use(VueRouter);

const router = new VueRouter({
  mode: 'history',
  routes: [
    { path: '/', component: () => import('./views/Home.vue') },
    { path: '/dashboard', component: () => import('./views/Dashboard.vue') },
    { path: '/users/:id', component: () => import('./views/UserProfile.vue') },
  ],
});

Vue.use(TraceKitPlugin, {
  apiKey: process.env.VUE_APP_TRACEKIT_API_KEY,
  serviceName: 'my-vue-app',
  release: process.env.VUE_APP_VERSION || '0.0.0',
  environment: process.env.NODE_ENV,
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  tracePropagationTargets: [
    'https://api.myapp.com',
  ],
  router,
});

new Vue({
  router,
  render: (h) => h(App),
}).$mount('#app');
```

## Troubleshooting

### Errors not captured by Vue error handler

- **Plugin registration order:** `TraceKitPlugin` must be registered with `app.use()` BEFORE calling `app.mount()`. If registered after, the error handler is not installed before the app starts rendering.
- **Existing error handler:** If you have a custom `app.config.errorHandler`, the plugin chains with it. Ensure your handler does not swallow errors silently.
- **Async errors:** Errors in `async` methods called from `@click` handlers are NOT caught by Vue's `errorHandler` in Vue 2. Use `try/catch` with `captureException`. Vue 3 catches these automatically.

### Vue 2 vs Vue 3 init differences

| Feature | Vue 2 | Vue 3 |
|---------|-------|-------|
| Plugin registration | `Vue.use(TraceKitPlugin, opts)` | `app.use(TraceKitPlugin, opts)` |
| Error handler | `Vue.config.errorHandler` | `app.config.errorHandler` |
| Instance access | `this.$tracekit` | `useTraceKitSpan()` composable |
| Router | `new VueRouter()` | `createRouter()` |
| Env vars | `VUE_APP_*` (Webpack) | `VITE_*` (Vite) |

### Router events not tracking

- **Check router is passed to plugin:** Navigation breadcrumbs only work when `router` is included in plugin options.
- **Check router version:** Vue Router 3.x for Vue 2, Vue Router 4.x for Vue 3. Mismatched versions cause silent failures.
- **Manual setup:** If not using the plugin, call `setupRouterBreadcrumbs(router)` after creating the router.

### Distributed tracing not connecting

- **Check `tracePropagationTargets`:** URLs must match your backend endpoints. Verify `traceparent` headers appear in the browser Network tab.
- **Check CORS:** Your backend must accept `traceparent` and `tracestate` in `Access-Control-Allow-Headers`.
- **Check backend SDK:** The backend must be instrumented with a TraceKit SDK that reads `traceparent`.

### Source maps not resolving

- **Check release version:** The `release` in plugin config must match the `--release` flag during upload.
- **Vue CLI output:** Upload from `./dist`.
- **Vite output:** Upload from `./dist`.

## Next Steps

Once your Vue app is traced, consider:
- **Code Monitoring** — Set live breakpoints and capture snapshots in production without redeploying (already enabled via `enableCodeMonitoring: true`)
- **Session Replay** — Visual debugging with full session recordings (see `tracekit-session-replay` skill)
- **Source Maps** — Readable stack traces with original source code (see `tracekit-source-maps` skill)
- **Backend Tracing** — Add `@tracekit/node-apm` or another backend SDK for end-to-end distributed traces (see `tracekit-node-sdk`, `tracekit-go-sdk`, and other backend skills)
- **Browser SDK** — For advanced browser-level configuration, see the `tracekit-browser-sdk` skill

## References

- Vue SDK docs: `https://app.tracekit.dev/docs/frontend/frameworks/vue`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
- Quick start: `https://app.tracekit.dev/docs/quickstart`
