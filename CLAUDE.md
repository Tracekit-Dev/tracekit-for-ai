# TraceKit AI Skills Repo

This repo contains structured SKILL.md files that teach AI coding assistants how to set up and configure TraceKit APM. Each skill provides step-by-step instructions with working code snippets that can be applied directly to a user's project.

## Repo Structure

- `skills/` — Per-SDK and per-feature skill directories
- `SKILL_TREE.md` — Master index of all available skills
- `.claude-plugin/` — Claude Code plugin manifest
- `.cursor-plugin/` — Cursor plugin configuration

## Skill Format

Skills follow the Agent Skills open standard (agentskills.io):

- **Frontmatter:** YAML with `name` (required) and `description` (required)
- **Body:** Markdown following the detect -> configure -> verify pattern
- **Code blocks:** Complete, copy-paste-ready snippets for each supported framework

## Naming Conventions

- SDK skills: `tracekit-{sdk}-sdk` (e.g., `tracekit-go-sdk`, `tracekit-node-sdk`, `tracekit-react-sdk`)
- Feature skills: `tracekit-{feature}` (e.g., `tracekit-code-monitoring`, `tracekit-session-replay`)
- All skills live under `skills/{skill-name}/SKILL.md`

## Non-Negotiable Rules

1. **Never hardcode API keys** in code snippets. Always use `TRACEKIT_API_KEY` env var.
2. **Always include a dashboard verification step** confirming data appears in `https://app.tracekit.dev`.
3. **Always reference TraceKit docs** for advanced configuration beyond what the skill covers.
4. **Use env vars for all secrets** — `.env` files, CI secrets, or deployment secret managers.

## Key URLs

- Dashboard: `https://app.tracekit.dev`
- Docs root: `https://app.tracekit.dev/docs`
- Frontend docs: `https://app.tracekit.dev/docs/frontend`
- Backend docs (Node): `https://app.tracekit.dev/docs/languages/nodejs`
- Backend docs (Go): `https://app.tracekit.dev/docs/languages/go`

## Working in This Repo

When adding or modifying skills:

1. Each skill is a standalone directory under `skills/`
2. Update `SKILL_TREE.md` when adding new skills
3. Test that YAML frontmatter parses correctly
4. Ensure all code snippets are complete and runnable
5. Cross-reference related skills where relevant
