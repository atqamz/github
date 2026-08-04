# AGENTS.md

Repo-specific rules. Global rules apply unless overridden here.

## Rules

- Read `README.md` first. It documents the baseline, the stack, and the five things deliberately left unmanaged (required status checks, required signatures, labels, archived repos, unmodelled rule types) - do not "fix" any of them without reading why.
- The Pulumi project is `github-config`, not `github`. Never rename it to match the repo: project config keys land in a namespace named after the project, and `github:` is the provider's own namespace.
- Adding or changing a repo is an edit to the `repos` table in `repos.go`, never a new resource block. A field belongs in `spec` only if it genuinely varies per repo; anything uniform goes in `apply`'s baseline in `main.go`.
- Before commit: `nix fmt`, then `go vet ./... && go build ./...` in the devshell.
- Never run `pulumi up` unattended. `pulumi preview` is the safe read; `up` mutates live repo settings across every personal repo at once.
- Removing a row from `repos` archives that repo (`archive_on_destroy`), and only after a deliberate `pulumi state unprotect`. Never unprotect to make a preview clean.
- The `adopt` config flag is for a first run against a repo that already exists. Leave it unset in the committed stack file.
- This repo is the source of truth for repo *settings*. Branch protection lives here; workflows, dependabot config and labels live in each repo.
