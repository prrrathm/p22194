# Repository guidance

## Documentation work

For project, product, technical, or operational documentation, use the `documentation-workflow` skill. Read the linked source and any template, separate confirmed facts from proposed decisions, save in the canonical location, and verify the saved artifact. When the document leaves decisions or delivery actions open, create linked, human-owned tracker tasks for review and for each independently answerable decision/action.

## Repository shape

This root is a Go workspace and a collection of Git submodules. Each submodule has its own history and may contain more specific instructions; read a submodule's `AGENTS.md` and README before changing it. Keep a change in the repository that owns the code. Update the parent repository's recorded submodule revision only when that revision change is part of the requested work.

Known areas:

| Path | Repository or role |
| --- | --- |
| `api-gateway` | Go API gateway submodule |
| `mdb-schemas` | Shared Go schemas and data access submodule |
| `services/users` | Go users service submodule |
| `services/kitchen` | Go kitchen service submodule |
| `services/p22194-god-svc` | Go service submodule |
| `services/mail` | Go service in the root repository |
| `services/find-my-trip` | Python service submodule |
| `services/searxng` | External service submodule |
| `clients/*` | Client application submodules |
| `cloudflare-worker` | Cloudflare Worker code |

The root `go.work` currently includes `api-gateway`, `mdb-schemas`, `services/kitchen`, `services/mail`, `services/users`, and `services/p22194-god-svc`. Check the target module's own `go.mod` and local instructions before choosing Go commands.

## Working safely

- Check `git status` in the root and every affected submodule before editing. Preserve existing user changes and unrelated submodule revisions.
- Inspect only files needed for the task. `.env` files may contain credentials; use `.env.example` to understand configuration and never print or copy secret values into documentation, logs, or commits.
- `Taskfile.yml` loads the root `.env`. Its targets include `init`, `update-submodules`, `run-gateway`, and service/client run targets. `init` and `update-submodules` change submodule state; use them only when the requested task needs that change.
- Root services are not one deployable application. Follow each service's local build, test, and deployment configuration.
- Keep API and schema changes compatible with their callers. Update or document affected client and service contracts when a change crosses repository boundaries.

## Validation

For a requested validation, use the affected module's documented commands and report the exact checks that ran. The root `Taskfile.yml` contains development/run tasks, not a repository-wide test target. Do not claim a service or integration check passed based only on a successful build or a different module's tests.
