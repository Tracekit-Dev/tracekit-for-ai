---
name: tracekit-source-maps
description: Set up source map uploading for TraceKit to get readable stack traces from minified JavaScript errors. Covers build plugin setup (webpack/vite), CLI upload, and CI pipeline integration. Use when the user asks about source maps, readable errors, or symbolication.
---

# TraceKit Source Maps

## When To Use

Use this skill when the user asks to:
- Upload source maps to TraceKit
- Get readable stack traces from minified JavaScript errors
- Symbolicate minified errors
- Debug production JS errors with original source code
- Set up source map integration in a CI/CD pipeline
- Fix "minified code" in error stack traces

**Source maps let you see original file names, line numbers, and column numbers** in stack traces instead of minified bundled code. Without source maps, production JavaScript errors show references like `main.a3f2b.js:1:4523` instead of `src/components/Dashboard.tsx:42:8`.

## Non-Negotiable Rules

1. **Never hardcode auth tokens** in code or CI config. Always use `TRACEKIT_AUTH_TOKEN` env var.
2. **Always verify symbolication** by triggering a real error and checking the stack trace shows original source.
3. **Source maps must match the deployed release version exactly** -- the `release` in SDK init must match the `--release` used during upload.
4. **Never serve source maps publicly in production** -- upload to TraceKit only, then delete local `.map` files before deploying.

## Prerequisites

Source maps require:
- A **frontend SDK** installed and initialized (`@tracekit/browser`, `@tracekit/react`, etc.)
- A **build step** that produces minified JavaScript (webpack, vite, rollup, esbuild, etc.)
- **TraceKit CLI** for uploading maps (installed in Step 1)

**If no frontend SDK is detected**, redirect to `tracekit-browser-sdk` skill (or the appropriate framework skill) first.

## Detection

Before applying this skill, detect the project setup:

1. **Check `package.json`** for a TraceKit frontend package:
   - `@tracekit/browser`, `@tracekit/react`, `@tracekit/vue`, `@tracekit/angular`, `@tracekit/nextjs`, `@tracekit/nuxt`
   - If none found, redirect to `tracekit-browser-sdk` skill
2. **Detect build tool** to determine plugin path:
   - `vite.config.*` => Vite plugin path (Step 3A)
   - `webpack.config.*` => Webpack plugin path (Step 3B)
   - Neither => CLI-only path (Step 3C)

## Step 1: Install TraceKit CLI

Install the TraceKit CLI using Homebrew or the install script:

```bash
# macOS / Linux (Homebrew)
brew install Tracekit-Dev/tap/tracekit

# Or use the install script
curl -fsSL https://raw.githubusercontent.com/Tracekit-Dev/cli/main/install.sh | sh
```

Verify the installation:

```bash
tracekit --version
```

The CLI is used for uploading source maps to TraceKit. See https://github.com/Tracekit-Dev/cli for manual downloads (Windows, ARM64, etc.).

## Step 2: Configure Auth Token

Source map uploads require an auth token (separate from the API key used in the SDK).

**Get your auth token:**
1. Log in to [TraceKit](https://app.tracekit.dev)
2. Go to **Settings > Auth Tokens**
3. Click **"Create Token"**
4. Select the **`releases:write`** scope
5. Copy the generated token

**Set the environment variable:**

```bash
export TRACEKIT_AUTH_TOKEN=your_auth_token_here
```

Add to your `.env` file for local development. For CI, add it as a repository secret (see Step 5).

Do **not** commit auth tokens. Use `.env` files, CI secrets, or deployment secret managers.

## Step 3: Build Plugin Setup

Choose the path matching your build tool. The build plugin automatically uploads source maps after each build.

### Path A: Vite

Install the Vite plugin:

```bash
npm install -D @tracekit/vite-plugin
```

Configure in `vite.config.ts`:

```javascript
// vite.config.ts
import { defineConfig } from 'vite';
import { traceKitSourceMaps } from '@tracekit/vite-plugin';

export default defineConfig({
  build: {
    sourcemap: true, // Required: generate source maps
  },
  plugins: [
    traceKitSourceMaps({
      authToken: process.env.TRACEKIT_AUTH_TOKEN,
      release: process.env.npm_package_version,
      cleanArtifacts: true, // Delete .map files after upload
    }),
  ],
});
```

### Path B: Webpack

Install the Webpack plugin:

```bash
npm install -D @tracekit/webpack-plugin
```

Configure in `webpack.config.js`:

```javascript
// webpack.config.js
const { TraceKitSourceMapPlugin } = require('@tracekit/webpack-plugin');

module.exports = {
  devtool: 'source-map', // Required: generate source maps
  plugins: [
    new TraceKitSourceMapPlugin({
      authToken: process.env.TRACEKIT_AUTH_TOKEN,
      release: process.env.npm_package_version,
      cleanArtifacts: true, // Delete .map files after upload
    }),
  ],
};
```

### Path C: CLI Only (Rollup, esbuild, or custom)

If you use a build tool without a dedicated plugin, upload source maps manually via the CLI after building:

1. **Ensure your build generates source maps** (consult your bundler docs).
2. **Upload after build:**

```bash
tracekit sourcemaps upload \
  --auth-token=$TRACEKIT_AUTH_TOKEN \
  --release=$(node -p "require('./package.json').version") \
  ./dist
```

3. **Delete source maps before deploying:**

```bash
find ./dist -name "*.map" -delete
```

## Step 4: Release Association

The `release` value must match between your SDK init and the source map upload. This is how TraceKit maps errors to the correct source maps.

```javascript
import { init } from '@tracekit/browser';

init({
  apiKey: import.meta.env.VITE_TRACEKIT_API_KEY,
  serviceName: 'my-frontend-app',
  endpoint: 'https://app.tracekit.dev/v1/traces',
  enableCodeMonitoring: true,
  release: import.meta.env.VITE_APP_VERSION || '0.0.0', // Must match upload --release
});
```

**Common release strategies:**
- **`package.json` version** -- `process.env.npm_package_version` (e.g., `1.2.3`)
- **Git SHA** -- `process.env.GITHUB_SHA` or output of `git rev-parse HEAD` (e.g., `a1b2c3d`)
- **Build timestamp** -- less common, harder to correlate

Pick one strategy and use it consistently in both SDK init and source map upload.

## Step 5: CI Pipeline Integration

Automate source map uploads in your deployment pipeline. Here is a complete GitHub Actions workflow:

```yaml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4  # https://github.com/actions/checkout
      - uses: actions/setup-node@v4  # https://github.com/actions/setup-node
        with:
          node-version: 20

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build

      - name: Upload source maps to TraceKit
        run: |
          tracekit sourcemaps upload \
            --auth-token=$TRACEKIT_AUTH_TOKEN \
            --release=${{ github.sha }} \
            ./dist
        env:
          TRACEKIT_AUTH_TOKEN: ${{ secrets.TRACEKIT_AUTH_TOKEN }}

      - name: Delete local source maps
        run: find ./dist -name "*.map" -delete

      - name: Deploy
        run: echo "Your deploy command here (e.g., aws s3 sync, vercel deploy, etc.)"
```

**Setup required:**
1. Go to your GitHub repo > **Settings > Secrets and variables > Actions**
2. Add a new secret: **`TRACEKIT_AUTH_TOKEN`** with your auth token value
3. Update the `release` value in your SDK init to use `process.env.GITHUB_SHA` (or match whatever you pass to `--release`)

**For other CI providers**, the pattern is the same:
1. Install dependencies
2. Build (with source maps enabled)
3. Upload source maps via CLI
4. Delete `.map` files
5. Deploy

## Step 6: Verification

Verify source maps are working end to end:

1. **Build your application** -- you should see "Source maps uploaded successfully" in the build output (if using a plugin) or after the CLI upload command
2. **Deploy the built application** (without `.map` files)
3. **Trigger a JavaScript error** -- add this temporarily:
   ```javascript
   import { captureException } from '@tracekit/browser';
   setTimeout(() => {
     captureException(new Error('Source map test error'));
   }, 2000);
   ```
4. **Visit** `https://app.tracekit.dev/issues` and find the error
5. **Check the stack trace** -- it should show:
   - Original file names (e.g., `src/components/Dashboard.tsx`)
   - Original line numbers and column numbers
   - Original function names
   - **NOT** minified references like `main.a3f2b.js:1:4523`
6. **Remove the test error code** once verified

## Troubleshooting

### Stack trace still shows minified code

- **Check release version matches** -- the `release` in `init()` must exactly match the `--release` used during upload. Print both to verify.
- **Check source maps were generated** -- verify `.map` files exist in your build output directory before upload. For Vite: `build: { sourcemap: true }`. For Webpack: `devtool: 'source-map'`.
- **Check upload succeeded** -- run `tracekit sourcemaps list --release=YOUR_RELEASE` to verify maps are uploaded.

### "Source map not found" warning in dashboard

- **Check upload included all `.map` files** -- the CLI uploads all `.map` files in the specified directory recursively.
- **Check file paths match** -- source map file names must correspond to the deployed JS file names.
- **Check the release exists** -- go to Dashboard > Releases and verify the release version appears.

### Auth token errors during upload

- **Verify token has `releases:write` scope** -- tokens without this scope cannot upload source maps.
- **Check `TRACEKIT_AUTH_TOKEN` is set** -- run `echo $TRACEKIT_AUTH_TOKEN` to verify (should not be empty).
- **Check token is not expired** -- auth tokens can have expiration dates. Generate a new one if needed.

### CI upload fails

- **Check `TRACEKIT_AUTH_TOKEN` secret is set** in your CI provider's secret management (GitHub: Settings > Secrets > Actions).
- **Check build output directory** -- ensure `./dist` (or your output dir) contains `.map` files before the upload step.
- **Check network access** -- CI runners must reach `https://app.tracekit.dev`.

## Next Steps

Once source maps are working, consider:
- **Releases** (`tracekit-releases` skill) -- Track deployments with version metadata and deploy notifications
- **Alerts** (`tracekit-alerts` skill) -- Get notified when new error types appear or error rates spike after a deploy

## References

- Source maps docs: `https://app.tracekit.dev/docs/frontend/source-maps`
- TraceKit CLI docs: `https://app.tracekit.dev/docs/cli`
- TraceKit docs root: `https://app.tracekit.dev/docs`
- Dashboard: `https://app.tracekit.dev`
