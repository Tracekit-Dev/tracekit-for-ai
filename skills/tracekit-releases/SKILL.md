---
name: tracekit-releases
description: Set up release tracking with TraceKit to monitor crash-free rates, deploy health, and regressions per release. Covers CLI release creation, deploy marking, and SDK configuration. Use when the user asks about releases, deployments, version tracking, or crash-free rates.
---

# TraceKit Release Tracking

## When To Use

Use this skill when the user asks to:
- Track releases or deployments in TraceKit
- Monitor crash-free rate per release
- Set up version tracking for errors
- Get deploy notifications
- Detect regressions per release
- Associate commits with a release
- Integrate release tracking into CI/CD

## Non-Negotiable Rules

1. **Never hardcode auth tokens** in code or CI configs. Always use `TRACEKIT_AUTH_TOKEN` env var.
2. **Always set `release` in SDK init** to match the CLI release version exactly.
3. **Always verify crash-free rate appears** on the dashboard after the first deploy.

## Prerequisites

- Any TraceKit SDK (backend or frontend) must be installed and sending data to the dashboard.
- `@tracekit/cli` must be installed (see Step 1 below).
- If no SDK is set up yet, complete the appropriate SDK skill first (e.g., `tracekit-node-sdk`, `tracekit-go-sdk`, `tracekit-react-sdk`).

## Detection

Before applying this skill, detect the project type:

1. **Check for any TraceKit package** in `package.json`, `go.mod`, `requirements.txt`, `composer.json`, `pom.xml`, `.csproj`, or `Gemfile`.
2. **If no SDK detected**, redirect to the `tracekit-apm-setup` skill to install an SDK first.
3. **Releases work with ALL SDKs** — no SDK-specific branching needed. Every SDK supports the `release` config key.

## Step 1: Install TraceKit CLI

```bash
npm install -D @tracekit/cli
```

This installs the TraceKit CLI as a dev dependency. It provides commands for creating releases, uploading source maps, and marking deploys.

## Step 2: Configure Auth Token

Set the `TRACEKIT_AUTH_TOKEN` environment variable. This is separate from `TRACEKIT_API_KEY` (used by SDKs) — the auth token is used by the CLI for release management.

```bash
export TRACEKIT_AUTH_TOKEN=your_auth_token_here
```

Where to get your auth token:
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Navigate to **Settings > Auth Tokens**
3. Click **Create Token**
4. Select the `releases:write` scope
5. Copy the token and store it securely

Do **not** commit auth tokens. Use `.env` files, CI secret managers, or deployment secret stores.

## Step 3: Set Release in SDK Init

Configure your SDK to tag all events with the release version. The `release` value in SDK init must match exactly what you pass to the CLI.

### Node.js

```javascript
const TraceKit = require('@tracekit/node');

TraceKit.init({
  apiKey: process.env.TRACEKIT_API_KEY,
  serviceName: 'my-node-service',
  release: process.env.npm_package_version, // e.g., "1.2.3"
});
```

### Go

```go
sdk, err := tracekit.NewSDK(&tracekit.Config{
    APIKey:      os.Getenv("TRACEKIT_API_KEY"),
    ServiceName: "my-go-service",
    Release:     version, // set from build flags or env var
})
```

### Python

```python
import os
import tracekit

tracekit.init(
    api_key=os.environ.get('TRACEKIT_API_KEY'),
    service_name='my-python-service',
    release=os.environ.get('APP_VERSION', 'dev'),
)
```

All TraceKit SDKs support a `release` config key. The value can be any string — semver versions, git SHAs, or build numbers all work.

## Step 4: Create a Release

Use the CLI to create a release in TraceKit. This registers the version so errors and transactions can be grouped by release.

```bash
npx tracekit-cli releases new $VERSION \
  --auth-token=$TRACEKIT_AUTH_TOKEN
```

Replace `$VERSION` with your release identifier (e.g., `1.2.3`, `abc123`).

## Step 5: Associate Commits (Optional)

Link commits to a release so you can see which code changes are included. This enables the "Suspect Commits" feature that highlights likely culprits for new errors.

```bash
npx tracekit-cli releases set-commits $VERSION \
  --auto \
  --auth-token=$TRACEKIT_AUTH_TOKEN
```

The `--auto` flag detects commits from your Git repository automatically. You must run this from within the Git repo.

## Step 6: Mark Deploy

Signal that a release has been deployed to an environment. This marks the deploy on the timeline and starts tracking crash-free rate from this point.

```bash
npx tracekit-cli releases deploys $VERSION new \
  --env=production \
  --auth-token=$TRACEKIT_AUTH_TOKEN
```

Supported environments: `production`, `staging`, `development`, or any custom string.

## Step 7: CI Pipeline Integration

Combine release creation, commit association, and deploy marking in your CI pipeline. Here is a GitHub Actions example:

```yaml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history for commit association

      - name: Install dependencies
        run: npm ci

      - name: Build and deploy
        run: npm run build && npm run deploy

      - name: Create release and mark deploy
        run: |
          npx tracekit-cli releases new ${{ github.sha }} \
            --auth-token=$TRACEKIT_AUTH_TOKEN
          npx tracekit-cli releases set-commits ${{ github.sha }} \
            --auto \
            --auth-token=$TRACEKIT_AUTH_TOKEN
          npx tracekit-cli releases deploys ${{ github.sha }} new \
            --env=production \
            --auth-token=$TRACEKIT_AUTH_TOKEN
        env:
          TRACEKIT_AUTH_TOKEN: ${{ secrets.TRACEKIT_AUTH_TOKEN }}
```

**Key points:**
- `fetch-depth: 0` ensures full Git history is available for commit association.
- The auth token is stored in GitHub Secrets, never hardcoded.
- Using `github.sha` as the version ties each release to a specific commit.

## Step 8: Verification

After setting up release tracking, verify it works end-to-end:

1. **Create a release and mark a deploy** using the CLI commands above.
2. **Navigate to** `https://app.tracekit.dev/releases`.
3. **Find your release** in the list — it should show the version, deploy time, and environment.
4. **Verify crash-free rate** — a percentage should appear once enough sessions are collected.
5. **Trigger an error** in your application — confirm it appears associated with the correct release.
6. **Check the deploy timeline** — your deployment should appear as a marker on the timeline view.

## Troubleshooting

### Release not appearing in dashboard

- **Check auth token scope:** Ensure the token has `releases:write` permission. Create a new token if needed.
- **Check CLI output:** Run the CLI command without piping to see error messages. A 401 means invalid token, 403 means insufficient scope.
- **Check version string:** The version must be a non-empty string. Empty or null versions are silently ignored.

### Events not linked to release

- **Check `release` in SDK init** matches the version passed to `tracekit-cli releases new` exactly. Case and whitespace matter.
- **Check SDK is initialized** before the events are captured — if `release` is set after init, earlier events will not be tagged.

### Crash-free rate not calculating

- **Minimum session count:** TraceKit needs at least 100 sessions in a release for the crash-free rate to reach statistical significance and display.
- **Sessions must be enabled:** Ensure your SDK has session tracking enabled (it is by default in browser SDKs; backend SDKs count transactions as sessions).

### Deploy not showing on timeline

- **Check environment name:** Ensure you passed `--env` to the deploy command. Without it, the deploy may not appear in filtered views.
- **Check timing:** Deploy markers appear on the timeline based on when the CLI command ran, not when code was pushed.

## Next Steps

Once release tracking is configured, consider:
- **Alerts** — Set up alert rules to get notified on release regressions (see `tracekit-alerts` skill)
- **Source Maps** — Upload source maps per release for readable frontend stack traces (see `tracekit-source-maps` skill)
- **Distributed Tracing** — Connect traces across services to see the full request lifecycle (see `tracekit-distributed-tracing` skill)

## References

- Release tracking docs: `https://app.tracekit.dev/docs/releases`
- CLI reference: `https://app.tracekit.dev/docs/cli`
- Dashboard: `https://app.tracekit.dev`
