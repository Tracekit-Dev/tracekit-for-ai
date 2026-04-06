---
name: tracekit-session-replay
description: Set up TraceKit Session Replay to record and replay user sessions with linked distributed traces. Covers privacy settings, sampling rates, and GDPR compliance. Use when the user asks to record sessions, replay user interactions, or debug user-reported issues visually.
---

# TraceKit Session Replay

## Auth Bootstrap

Do not tell the user to go sign up, log in separately, or manually create an API key before setup. First check for existing TraceKit auth. If `~/.tracekitconfig` does not contain the production profile or `TRACEKIT_API_KEY` is missing, apply the `tracekit-auth` skill first. Use `./scripts/run-tracekit-auth.sh status` to check, then guide the user through the TraceKit email verification flow with `register` and `verify`. That flow signs the user into an existing account for that email or creates the account automatically, then saves the returned credentials for the rest of the setup.

## When To Use

Use this skill when the user asks to:
- Record user sessions for debugging
- Replay what a user did before an error
- Set up visual debugging or session recording
- See what users did on the page
- Debug user-reported issues visually
- Add session replay to a frontend app

**Session replay is a browser-only feature.** It records DOM mutations, user interactions, and network requests to reconstruct a video-like replay of user sessions. Replays are linked to distributed traces so you can see the full backend context for any user action.

## Non-Negotiable Rules

1. **Never hardcode API keys** in code. Always use environment variables or build-time injection.
2. **Text masking and input masking are always on and not configurable.** The SDK masks all text with same-length asterisk replacement and masks all input values by default. There is no option to disable this. You can only selectively unmask elements using `unmask` selectors or the `data-tracekit-unmask` attribute.
3. **Always include a verification step** confirming replays appear in `https://app.tracekit.dev/replays`.
4. **Always discuss GDPR/privacy implications** before enabling replay - session recordings capture user behavior and may be subject to data protection regulations.

## Prerequisites

Session replay requires a frontend SDK to be installed and initialized. Complete **one** of the following SDK skills first:

- `tracekit-browser-sdk` - Vanilla JavaScript/TypeScript
- `tracekit-react-sdk` - React applications
- `tracekit-vue-sdk` - Vue applications
- `tracekit-angular-sdk` - Angular applications
- `tracekit-nextjs-sdk` - Next.js applications
- `tracekit-nuxt-sdk` - Nuxt applications

**Session replay is a BROWSER-ONLY feature.** It does not apply to backend SDKs (Node.js, Go, Python, etc.). If the user is working on a backend project, this skill does not apply.

## Detection

Before applying this skill, verify a frontend TraceKit SDK is installed:

1. **Check `package.json`** for any TraceKit frontend package:
   - `@tracekit/browser` - vanilla JS/TS
   - `@tracekit/react` - React
   - `@tracekit/vue` - Vue
   - `@tracekit/angular` - Angular
   - `@tracekit/nextjs` - Next.js
   - `@tracekit/nuxt` - Nuxt
2. If **none found**, redirect to `tracekit-browser-sdk` skill (or the appropriate framework skill) to install the SDK first.
3. **Check if replay is already configured** - search for `@tracekit/replay` in `package.json` or `replayIntegration` in source files.

## Step 1: Install Replay Integration

```bash
npm install @tracekit/replay
```

The replay integration is a separate package that adds session recording capabilities to your existing TraceKit browser SDK.

## Step 2: Add Replay to SDK Init

Add the replay integration to your TraceKit initialization:

```javascript
import { init } from '@tracekit/browser';
import { replayIntegration } from '@tracekit/replay';

const replay = replayIntegration({
  // Sampling
  sessionSampleRate: 0.1,   // Record 10% of all sessions (default)
  errorSampleRate: 0.0,     // Capture rate for error-only buffer sessions (default)

  // Privacy - blockMedia defaults to true, text/input masking is always on
  blockMedia: true,          // Block images, videos, canvas, svg, iframe (default)
  inlineImages: false,       // Do not inline images as base64 (default)
  unmask: [],                // CSS selectors for elements safe to show unmasked
});

init({
  apiKey: import.meta.env.VITE_TRACEKIT_API_KEY,
  serviceName: 'my-frontend-app',
  endpoint: 'https://app.tracekit.dev',
  addons: [replay],
});
```

## Step 3: Privacy Configuration

This is the most critical section. Session replay records user interactions, so privacy must be understood before deploying to production.

### Text and Input Masking (Always On)

Text masking and input masking are **always enabled and not configurable**. The SDK:
- Masks all text content with **same-length asterisk replacement** (e.g., "Hello" becomes "*****")
- Masks all form input values (passwords, emails, search queries)
- Uses `maskTextSelector: '*'` and `maskAllInputs: true` internally - these cannot be turned off

To selectively show specific text content, use unmasking:

| Setting | Default | Description |
|---|---|---|
| `unmask: ['.public-content']` | `[]` | CSS selectors for elements safe to show unmasked. Use sparingly. |
| `data-tracekit-unmask` attribute | n/a | Add this HTML attribute to any element that should display unmasked text. |

**Example: Unmasking specific public content:**

```html
<!-- These elements will show their real text in replays -->
<h1 data-tracekit-unmask>Welcome to Our App</h1>
<nav class="public-nav" data-tracekit-unmask>
  <a href="/pricing">Pricing</a>
  <a href="/docs">Docs</a>
</nav>

<!-- Everything else is masked automatically -->
<p>This text appears as asterisks in replay</p>
```

```javascript
const replay = replayIntegration({
  unmask: ['.public-nav', '.marketing-hero'],  // CSS selectors to unmask
});
```

### Media Blocking

| Setting | Default | Description |
|---|---|---|
| `blockMedia: true` | **true** | Blocks images, videos, canvas, SVG, and iframe elements from recording. Replaced with placeholders. **Strongly recommended.** |
| `inlineImages: false` | **false** | When true, captures images as base64 data URIs. Increases payload size significantly. Only enable if you need to see images in replays. |

### GDPR Considerations

Session replay records user behavior and may be subject to GDPR, CCPA, or other data protection regulations:

1. **Consent required** - Display a consent banner before enabling session replay. Only start recording after the user consents.
2. **Data retention** - Configure retention in the dashboard (Settings > Data Retention). Default is 30 days.
3. **Right to deletion** - Users can request deletion of their session data. Use the TraceKit API to delete sessions by user ID.
4. **Data processing agreement** - Ensure your TraceKit DPA covers session replay data.
5. **Privacy policy** - Update your privacy policy to mention session recording.

**Conditional initialization based on consent:**

```javascript
import { init } from '@tracekit/browser';
import { replayIntegration } from '@tracekit/replay';

// Only add replay if user has consented
const addons = [];
if (userHasConsentedToRecording()) {
  addons.push(replayIntegration({
    blockMedia: true,
  }));
}

init({
  apiKey: import.meta.env.VITE_TRACEKIT_API_KEY,
  serviceName: 'my-frontend-app',
  endpoint: 'https://app.tracekit.dev',
  addons,
});
```

## Step 4: Sampling Configuration

Control how many sessions are recorded to manage storage costs and volume.

```javascript
const replay = replayIntegration({
  sessionSampleRate: 0.1,   // Record 10% of all sessions
  errorSampleRate: 0.5,     // Use 50% of remaining budget for error buffer capture
});
```

### Three-Mode Sampling System

The SDK uses a three-mode system based on a random roll at session start:

| Mode | Range | Behavior |
|---|---|---|
| `session` | `[0, sessionSampleRate)` | Full recording - all events forwarded immediately for upload. |
| `buffer` | `[sessionSampleRate, sessionSampleRate + errorSampleRate)` | Error capture - events stored in a 60-second ring buffer. If an error occurs, the buffer is flushed and mode switches to `session` for continued recording. |
| `off` | `[sessionSampleRate + errorSampleRate, 1.0]` | No recording - events are discarded. No recording pipeline is set up. |

**Important:** `sessionSampleRate + errorSampleRate` must not exceed 1.0. If it does, the SDK clamps `errorSampleRate` automatically.

| Setting | Default | Purpose |
|---|---|---|
| `sessionSampleRate` | `0.1` (10%) | Fraction of sessions that get full recording from the start. |
| `errorSampleRate` | `0.0` (0%) | Fraction of sessions that capture errors via ring buffer. Set higher to catch errors in sessions that were not selected for full recording. |

**How it works:**
- At session start, a random number determines the mode: `session`, `buffer`, or `off`
- In `buffer` mode, events are kept in a 60-second ring buffer
- If an error occurs during a `buffer` session, all buffered events are flushed and recording switches to full `session` mode
- In `session` mode, events are forwarded immediately to the compression and upload pipeline

**Cost implications:**
- Higher `sessionSampleRate` = more storage = higher bill
- Start with `0.1` (10%) for general sessions and increase `errorSampleRate` to catch errors
- Adjust based on traffic volume and budget

### Additional Replay Settings

| Setting | Default | Description |
|---|---|---|
| `idleTimeout` | `1800000` (30 min) | Idle timeout in ms before the session ends. After timeout, pending events are flushed, a new session ID is generated, and a new sampling decision is made. |
| `flushInterval` | `30000` (30s) | Interval in ms between automatic uploads of recorded events. |
| `maxBufferSize` | `24117248` (23MB) | Maximum buffer size in bytes for pending events before forced flush. |

## Step 5: Manual Control Methods

The `replayIntegration()` function returns an integration object with two manual control methods:

```javascript
const replay = replayIntegration({
  sessionSampleRate: 0.1,
});

// Force an immediate upload of pending events (useful before page transitions)
replay.flush();

// Get the current session ID (useful for linking to support tickets or logs)
const sessionId = replay.getSessionId();  // Returns '' if replay is not active
```

**Use cases for `flush()`:**
- Before a known page navigation to ensure events are sent
- At critical user actions (e.g., checkout completion)

**Use cases for `getSessionId()`:**
- Including session ID in support tickets
- Logging session ID alongside server-side traces
- Building custom "View Replay" links in internal tools

## Step 6: Linking Replays to Traces

When both session replay and distributed tracing are enabled, they are linked automatically. The SDK injects a `replay_id` tag on error events so the playback UI can link errors to their replay session. No additional configuration needed.

### View a Replay from a Trace

1. Open a trace in `https://app.tracekit.dev/traces`
2. If the trace originated from a browser session with replay, a **"View Replay"** button appears in the trace detail header
3. Click to jump to the replay at the exact moment the traced request occurred

### View a Trace from a Replay

1. Open a replay in `https://app.tracekit.dev/replays`
2. The replay timeline shows **event markers** for errors, network requests, and console logs
3. Click any event marker to see details, including the **Trace ID**
4. Click the Trace ID to jump to the full distributed trace waterfall

### Filter Replays

Use the replay list to find specific sessions:
- **By error type** - find all sessions where a specific error occurred
- **By user** - find all sessions for a specific user (requires `setUser()` in SDK)
- **By time range** - find sessions during an incident window
- **By URL** - find sessions that visited a specific page

## Step 7: Verification

Verify session replay is working:

1. **Start your application** with replay configured
2. **Set `sessionSampleRate: 1.0` during testing** to ensure every session is recorded
3. **Navigate through a few pages** and interact with the UI (clicks, form inputs, page transitions)
4. **Visit** `https://app.tracekit.dev/replays`
5. **Find your session** in the replay list (most recent, filtered by your user or time)
6. **Click to play** the replay
7. **Verify privacy settings:**
   - Text should appear as asterisks matching the original text length (e.g., "Hello" shows as "*****")
   - Images and videos should be blocked/replaced with placeholders (if `blockMedia: true`)
   - Form inputs should be masked
   - Any elements with `data-tracekit-unmask` or matching `unmask` selectors should show real text
8. **Click an error marker** (if any) to verify it links to a distributed trace

## Troubleshooting

### Replays not appearing

- **Check sampling rate** - `sessionSampleRate: 0.1` means only 10% of sessions are recorded. Set to `1.0` during testing.
- **Check replay integration is added** - verify `@tracekit/replay` is in `package.json` and `replayIntegration()` is passed to `addons`.
- **Check browser console** for errors from the TraceKit SDK. Enable `debug: true` in init config for verbose logging.
- **Check Content Security Policy** - CSP must allow connections to `https://app.tracekit.dev`.

### All text visible (not masked)

- Text and input masking is always on. If text is showing unmasked, check:
  - The `unmask` config option is not too broad (e.g., do not use `unmask: ['*']`)
  - Elements do not have the `data-tracekit-unmask` attribute applied too widely
  - **Rebuild and redeploy** - config changes require a new deployment to take effect

### Session ending unexpectedly

- **Check `idleTimeout`** - default is 30 minutes of inactivity. After timeout, the session ends, events are flushed, and a new session starts with a fresh sampling decision.
- **Check visibility handling** - recording pauses when the tab is hidden and resumes when visible. This is automatic.
- **Check if user navigated away** - replay stops when the user closes the tab or navigates to an external site.

### Large replay file sizes

- **Lower `sessionSampleRate`** to reduce total recorded sessions.
- **Keep `blockMedia: true`** (default) to exclude images and videos from recordings.
- **Keep `inlineImages: false`** (default) to avoid base64-encoded images in the payload.
- **Check `maxBufferSize`** - default is 23MB. The SDK flushes when this limit is reached.

## Complete Configuration Reference

```javascript
import { init } from '@tracekit/browser';
import { replayIntegration } from '@tracekit/replay';

const replay = replayIntegration({
  // Sampling
  sessionSampleRate: 0.1,     // 0.0-1.0, default: 0.1 (10% full recording)
  errorSampleRate: 0.0,       // 0.0-1.0, default: 0.0 (error buffer capture rate)

  // Privacy
  blockMedia: true,            // Block img/video/canvas/svg/iframe, default: true
  inlineImages: false,         // Inline images as base64, default: false
  unmask: [],                  // CSS selectors to unmask, default: []
  // Note: text masking and input masking are always on (not configurable)
  // Use data-tracekit-unmask attribute or unmask selectors to show specific text

  // Timing
  idleTimeout: 1800000,        // 30 min idle before session ends, default: 1800000
  flushInterval: 30000,        // Upload interval in ms, default: 30000
  maxBufferSize: 24117248,     // Max buffer size in bytes (23MB), default: 24117248
});

init({
  apiKey: import.meta.env.VITE_TRACEKIT_API_KEY,
  serviceName: 'my-frontend-app',
  endpoint: 'https://app.tracekit.dev',
  addons: [replay],
});

// Manual control
replay.flush();                // Force immediate upload
replay.getSessionId();         // Get current session ID ('' if inactive)
```

## Next Steps

Once session replay is working, consider:
- **Source Maps** (`tracekit-source-maps` skill) - See readable stack traces in replay error markers instead of minified code
- **Alerts** (`tracekit-alerts` skill) - Get notified when replay-captured errors spike

## References

- Session replay docs: `https://app.tracekit.dev/docs/frontend/session-replay`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
