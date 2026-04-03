---
name: tracekit-angular-sdk
description: Sets up TraceKit APM in Angular applications with dependency injection, error handlers, router integration, and distributed tracing. Supports both NgModule and standalone component architectures.
---

# TraceKit Angular SDK Setup

## Auth Bootstrap

Do not tell the user to go sign up, log in separately, or manually create an API key before setup. First check for existing TraceKit auth. If `~/.tracekitconfig` does not contain the production profile or `TRACEKIT_API_KEY` is missing, apply the `tracekit-auth` skill first. Use `./scripts/run-tracekit-auth.sh status` to check, then guide the user through the TraceKit email verification flow with `register` and `verify`. That flow signs the user into an existing account for that email or creates the account automatically, then saves the returned credentials for the rest of the setup.

## When To Use

Use this skill when the user asks to:
- Add TraceKit to an Angular application
- Add observability, error tracking, or APM to an Angular project
- Instrument an Angular app with distributed tracing
- Set up error monitoring in an Angular SPA
- Configure TraceKit in an Angular project with NgModule or standalone components
- Debug production Angular applications with live breakpoints

If the user has a vanilla JavaScript/TypeScript project without Angular, use the `tracekit-browser-sdk` skill instead.

## Non-Negotiable Rules

1. **Never hardcode API keys** in code. Always use Angular environment files with `TRACEKIT_API_KEY` env var injection.
2. **Always include a verification step** confirming errors appear in `https://app.tracekit.dev`.
3. **Always enable code monitoring** (`enableCodeMonitoring: true`) -- it is TraceKit's differentiator.
4. **Always initialize TraceKit before app bootstrap** -- the SDK must load before the application starts to capture all errors.

## Detection

Before applying this skill, detect the project type:

1. **Check `package.json`** for `@angular/core` in dependencies -- confirms this is an Angular project.
2. **Detect architecture** by scanning `src/main.ts`:
   - `bootstrapModule(AppModule)` => NgModule architecture (use NgModule branch)
   - `bootstrapApplication(AppComponent)` => Standalone component architecture (use Standalone branch)
3. **Only ask the user** if `main.ts` is missing or uses an unrecognized bootstrap pattern.

## Step 1: Environment Setup

Set the `TRACEKIT_API_KEY` environment variable. Angular uses environment files for configuration.

Add to `src/environments/environment.ts` (development):

```typescript
export const environment = {
  production: false,
  tracekitApiKey: 'ctxio_your_dev_api_key_here',
};
```

Add to `src/environments/environment.prod.ts` (production):

```typescript
export const environment = {
  production: true,
  tracekitApiKey: process.env['TRACEKIT_API_KEY'] || '',
};
```

Where to get your API key:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **API Keys** page
3. Generate a new key (starts with `ctxio_`)

Do **not** commit real API keys. For production builds, inject `TRACEKIT_API_KEY` via your CI/CD pipeline or deployment environment.

## Step 2: Install SDK

```bash
npm install @tracekit/angular
```

Or with Yarn:

```bash
yarn add @tracekit/angular
```

This installs the TraceKit Angular wrapper with the `@tracekit/browser` SDK, ErrorHandler integration, router breadcrumbs, and HttpClient interceptor for distributed tracing.

## Step 3: Initialize TraceKit

Choose the branch matching your Angular architecture. Apply **one** of the following.

### Branch A: Standalone Components (Angular 15+)

In `src/main.ts`:

```typescript
import { bootstrapApplication } from '@angular/platform-browser';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideTraceKit, provideTraceKitRouter } from '@tracekit/angular';
import { AppComponent } from './app/app.component';
import { routes } from './app/app.routes';
import { environment } from './environments/environment';

bootstrapApplication(AppComponent, {
  providers: [
    ...provideTraceKit({
      apiKey: environment.tracekitApiKey,
      release: '1.0.0',
      environment: environment.production ? 'production' : 'development',
      endpoint: 'https://app.tracekit.dev/v1/traces',
      enableCodeMonitoring: true,
    }),
    provideRouter(routes),
    ...provideTraceKitRouter(),
    provideHttpClient(),
  ],
});
```

`provideTraceKit()` initializes the SDK and returns a `Provider[]` array that replaces Angular's default `ErrorHandler` with `TraceKitErrorHandler`. Use the spread operator to merge into your providers array.

`provideTraceKitRouter()` sets up router breadcrumbs via `APP_INITIALIZER`. The router is lazily resolved from Angular's injector to avoid circular dependency issues.

### Branch B: NgModule

In `src/app/app.module.ts`:

```typescript
import { NgModule, ErrorHandler } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule, Routes } from '@angular/router';
import { HttpClientModule } from '@angular/common/http';
import { TraceKitModule } from '@tracekit/angular';
import { AppComponent } from './app.component';
import { HomeComponent } from './home.component';
import { UsersComponent } from './users.component';
import { environment } from '../environments/environment';

const routes: Routes = [
  { path: '', component: HomeComponent },
  { path: 'users/:id', component: UsersComponent },
];

@NgModule({
  declarations: [AppComponent, HomeComponent, UsersComponent],
  imports: [
    BrowserModule,
    HttpClientModule,
    RouterModule.forRoot(routes),
    TraceKitModule.forRoot({
      apiKey: environment.tracekitApiKey,
      release: '1.0.0',
      environment: environment.production ? 'production' : 'development',
      endpoint: 'https://app.tracekit.dev/v1/traces',
      enableCodeMonitoring: true,
    }),
  ],
  bootstrap: [AppComponent],
})
export class AppModule {}
```

`TraceKitModule.forRoot()` returns a module with providers that initializes the SDK, replaces the default `ErrorHandler`, and sets up router breadcrumbs. Do not mix `TraceKitModule.forRoot()` with `provideTraceKit()` in the same application.

## Step 4: Error Handler

TraceKit replaces Angular's built-in `ErrorHandler` with `TraceKitErrorHandler`. This is done automatically by both `provideTraceKit()` (standalone) and `TraceKitModule.forRoot()` (NgModule).

The `TraceKitErrorHandler` class:
- Captures all unhandled errors via `captureException()`
- Extracts the original error from Angular wrappers (`error.originalError` for template errors, `error.rejection` for unhandled promise rejections)
- Preserves Angular's default `console.error` output
- Is registered via `{ provide: ErrorHandler, useClass: TraceKitErrorHandler }` -- no `@Injectable()` decorator needed

**Manual registration** (only if you need to customize provider ordering):

Standalone:
```typescript
import { ErrorHandler } from '@angular/core';
import { TraceKitErrorHandler } from '@tracekit/angular';

bootstrapApplication(AppComponent, {
  providers: [
    { provide: ErrorHandler, useClass: TraceKitErrorHandler },
  ],
});
```

NgModule:
```typescript
@NgModule({
  providers: [
    { provide: ErrorHandler, useClass: TraceKitErrorHandler },
  ],
})
export class AppModule {}
```

## Step 5: Router Integration

Navigation breadcrumbs are captured automatically via Angular Router events.

**Standalone:** Add `...provideTraceKitRouter()` to your providers array (shown in Step 3).

**NgModule:** Router integration is included in `TraceKitModule.forRoot()` automatically.

The router integration subscribes to `router.events` and captures `NavigationEnd` events as breadcrumbs with from/to paths.

To disable parameterized routes (use actual URLs instead of route patterns):

```typescript
// Standalone
...provideTraceKitRouter(false)

// NgModule
TraceKitModule.forRoot({
  ...config,
  parameterizedRoutes: false,
})
```

## Step 6: Custom Error Capture and Performance Spans

Import TraceKit functions directly from `@tracekit/angular` -- no dependency injection needed for the SDK functions:

```typescript
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { captureException, setUser } from '@tracekit/angular';

@Component({
  selector: 'app-user',
  template: '<h1>{{ user?.name }}</h1>',
})
export class UserComponent implements OnInit {
  user: any;

  constructor(private route: ActivatedRoute) {}

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    this.loadUser(id!);
  }

  async loadUser(id: string) {
    try {
      const res = await fetch('/api/users/' + id);
      this.user = await res.json();
      setUser({ id: this.user.id, email: this.user.email });
    } catch (err) {
      captureException(err as Error, { userId: id });
    }
  }
}
```

**Re-exported functions** available from `@tracekit/angular`:

```typescript
import {
  captureException,
  captureMessage,
  setUser,
  setTag,
  setExtra,
  addBreadcrumb,
  getClient,
} from '@tracekit/angular';
```

## Step 7: Distributed Tracing

Add the TraceKit HttpClient interceptor to propagate trace headers to your backend APIs:

**Standalone:**

```typescript
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { traceKitInterceptor } from '@tracekit/angular';

bootstrapApplication(AppComponent, {
  providers: [
    ...provideTraceKit({
      apiKey: environment.tracekitApiKey,
      endpoint: 'https://app.tracekit.dev/v1/traces',
      enableCodeMonitoring: true,
      tracePropagationTargets: ['https://api.example.com', /^\/api\//],
    }),
    provideHttpClient(withInterceptors([traceKitInterceptor])),
  ],
});
```

**NgModule:**

```typescript
import { HTTP_INTERCEPTORS } from '@angular/common/http';
import { TraceKitHttpInterceptor } from '@tracekit/angular';

@NgModule({
  providers: [
    {
      provide: HTTP_INTERCEPTORS,
      useClass: TraceKitHttpInterceptor,
      multi: true,
    },
  ],
})
export class AppModule {}
```

The `tracePropagationTargets` config option controls which outgoing requests receive trace headers. Set it to match your API domains.

## Step 8: Session Replay (Optional)

Enable session replay to record and replay user sessions linked to error traces:

```typescript
...provideTraceKit({
  apiKey: environment.tracekitApiKey,
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  replay: {
    enabled: true,
    sampleRate: 0.1,        // Record 10% of sessions
    errorSampleRate: 1.0,    // Record 100% of sessions with errors
    maskAllText: true,       // Mask sensitive text by default
    blockAllMedia: false,
  },
})
```

Session replay runs client-side only and captures DOM mutations, network requests, and console logs.

## Step 9: Source Maps (Optional)

Upload source maps for readable stack traces in production errors.

Add to your `angular.json` build configuration:

```json
{
  "projects": {
    "my-app": {
      "architect": {
        "build": {
          "configurations": {
            "production": {
              "sourceMap": true
            }
          }
        }
      }
    }
  }
}
```

After building, upload source maps:

```bash
tracekit sourcemaps upload \
  --api-key $TRACEKIT_API_KEY \
  --release 1.0.0 \
  --dist ./dist/my-app/browser
```

Add this command to your CI/CD pipeline after `ng build --configuration production`.

## Step 10: Verification

After integrating, verify errors are captured:

1. **Start your application** with `ng serve` or deploy to a test environment.
2. **Trigger a test error** -- add a temporary button that throws:
   ```typescript
   throwTestError() {
     throw new Error('TraceKit Angular test error');
   }
   ```
3. **Open** `https://app.tracekit.dev`.
4. **Confirm** the test error appears within 30-60 seconds with component stack trace.

If errors do not appear, see Troubleshooting below.

## Complete Working Example

### Standalone Application (Angular 15+)

```typescript
// src/main.ts
import { bootstrapApplication } from '@angular/platform-browser';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import {
  provideTraceKit,
  provideTraceKitRouter,
  traceKitInterceptor,
} from '@tracekit/angular';
import { AppComponent } from './app/app.component';
import { environment } from './environments/environment';

const routes = [
  { path: '', loadComponent: () => import('./app/home.component') },
  { path: 'users/:id', loadComponent: () => import('./app/user.component') },
];

bootstrapApplication(AppComponent, {
  providers: [
    ...provideTraceKit({
      apiKey: environment.tracekitApiKey,
      release: '1.0.0',
      environment: environment.production ? 'production' : 'development',
      endpoint: 'https://app.tracekit.dev/v1/traces',
      enableCodeMonitoring: true,
      tracePropagationTargets: ['https://api.example.com', /^\/api\//],
      replay: {
        enabled: true,
        sampleRate: 0.1,
        errorSampleRate: 1.0,
      },
    }),
    provideRouter(routes),
    ...provideTraceKitRouter(),
    provideHttpClient(withInterceptors([traceKitInterceptor])),
  ],
});
```

### NgModule Application

```typescript
// src/app/app.module.ts
import { NgModule, ErrorHandler } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule, Routes } from '@angular/router';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { TraceKitModule, TraceKitHttpInterceptor } from '@tracekit/angular';
import { AppComponent } from './app.component';
import { HomeComponent } from './home.component';
import { UsersComponent } from './users.component';
import { environment } from '../environments/environment';

const routes: Routes = [
  { path: '', component: HomeComponent },
  { path: 'users/:id', component: UsersComponent },
];

@NgModule({
  declarations: [AppComponent, HomeComponent, UsersComponent],
  imports: [
    BrowserModule,
    HttpClientModule,
    RouterModule.forRoot(routes),
    TraceKitModule.forRoot({
      apiKey: environment.tracekitApiKey,
      release: '1.0.0',
      environment: environment.production ? 'production' : 'development',
      endpoint: 'https://app.tracekit.dev/v1/traces',
      enableCodeMonitoring: true,
      tracePropagationTargets: ['https://api.example.com', /^\/api\//],
      replay: {
        enabled: true,
        sampleRate: 0.1,
        errorSampleRate: 1.0,
      },
    }),
  ],
  providers: [
    {
      provide: HTTP_INTERCEPTORS,
      useClass: TraceKitHttpInterceptor,
      multi: true,
    },
  ],
  bootstrap: [AppComponent],
})
export class AppModule {}
```

## Troubleshooting

### Errors not captured

- **Check ErrorHandler provider ordering:** If another library overrides `ErrorHandler` after TraceKit, TraceKit's handler will not receive errors. Ensure `provideTraceKit()` or `TraceKitModule.forRoot()` is listed before other provider overrides.
- **Check API key:** Ensure `environment.tracekitApiKey` is set. Print it: `console.log(environment.tracekitApiKey)`.
- **Check outbound access:** Your app must reach `https://app.tracekit.dev/v1/traces`. Test with: `curl -X POST https://app.tracekit.dev/v1/traces` (expect 401 -- means endpoint is reachable).

### NgModule vs standalone confusion

- **Detect via `src/main.ts`:** Look for `bootstrapModule(AppModule)` (NgModule) vs `bootstrapApplication(AppComponent)` (standalone).
- **Do not mix:** Never use both `TraceKitModule.forRoot()` and `provideTraceKit()` in the same app.

### Zone.js interactions

- TraceKit hooks into Angular's error handling via `ErrorHandler`, which is zone-aware. The SDK does not patch Zone.js directly.
- If using `zoneless` mode (Angular 18+), TraceKit still works via the `ErrorHandler` DI token.

### AOT compilation issues

- `TraceKitErrorHandler` is a plain class without `@Injectable()` decorator, so it works with AOT compilation out of the box.
- If you see "No provider for ErrorHandler" errors, ensure the TraceKit providers are in the root module/application config.

### HttpClient interceptor not adding trace headers

- Verify `tracePropagationTargets` includes your API domain.
- Standalone: ensure `withInterceptors([traceKitInterceptor])` is passed to `provideHttpClient()`.
- NgModule: ensure `HTTP_INTERCEPTORS` provider is registered with `multi: true`.

## Next Steps

Once your Angular app is traced, consider:
- **Browser SDK** -- For non-Angular pages in the same project, use the `tracekit-browser-sdk` skill
- **Session Replay** -- Record and replay user sessions with linked traces
- **Source Maps** -- Upload source maps for readable production stack traces
- **Backend SDKs** -- Connect frontend traces to backend services for full distributed tracing

## References

- Angular SDK docs: `https://app.tracekit.dev/docs/frontend/frameworks/angular`
- Browser SDK docs: `https://app.tracekit.dev/docs/frontend/browser-sdk`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
