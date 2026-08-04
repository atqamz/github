# github

Config-as-code for my personal GitHub repos, so repo settings are reviewable in a diff instead of clicked into a web form.

Pulumi + Go, state in Pulumi Cloud (`atqamz-gmail-com/github-config/prod`).

## Why

The settings that matter are the ones nobody remembers to set.
Before this repo, 14 of 16 repos had no ruleset at all (their default branch was force-pushable and deletable), all 17 allowed rebase merges despite a standing never-`--rebase` rule, and 14 left merged branches undeleted.
One table in `repos.go` fixed all of it in a single run, and keeps it fixed.

## Layout

- `repos.go` - the repo table. Adding a repo is one struct literal; a field belongs in `spec` only if it genuinely varies per repo, never as a restatement of the baseline.
- `main.go` - the baseline every repo gets, plus its alert switch and default-branch ruleset.
- `Pulumi.yaml`, `Pulumi.prod.yaml` - project and single stack (`prod`). Provider owner is `atqamz`.
- `flake.nix` - devshell with `pulumi`, the Go language plugin, `go`, `gopls`, `gh`.

The Pulumi project is named `github-config`, not `github`.
Project config lands in a namespace named after the project, so a project called `github` would put every key in `github:` - which is also the GitHub provider's own config namespace, and the provider rejects keys it does not recognise.
Renaming the project once avoids that collision for every future config key, not just the one that surfaced it.

## Baseline

Every managed repo gets:

| setting | value | why |
| --- | --- | --- |
| `allow_rebase_merge` | false | rebase merges rewrite committer dates and break signature chains |
| `allow_squash_merge` | true | for branches whose intermediate commits are noise |
| `allow_merge_commit` | true | the default merge strategy |
| `allow_auto_merge` | true | needed by the dependabot auto-merge workflow in `atqamz/universe` |
| `allow_update_branch` | true | lets a stale dependabot PR be brought forward without a local checkout |
| `delete_branch_on_merge` | true | no stale branch pile-up |
| `has_wiki` | false | 13 repos had one enabled and empty |
| `has_projects` | false | unused |
| `archive_on_destroy` | true | removing a row archives the repo, it does not delete it |

Plus, as separate resources:

- vulnerability alerts on
- `default_workflow_permissions: read` and `can_approve_pull_request_reviews: false`, so a workflow with no `permissions:` block of its own gets read-only. `secondhand` overrides the approve flag via `CanApprovePRs`, because `release-please-action` opens pull requests and cannot work without it.
- dependabot security updates on, except where `NoDependabotSecurityUpdates` says otherwise (`universe`; see its AGENTS.md for why its npm fetcher cannot work). Public repos only - private ones report the feature as unavailable on a personal plan.
- a `main guard` branch ruleset on `~DEFAULT_BRANCH` with `deletion` and `non_fast_forward` blocked

Alerts are a separate `RepositoryVulnerabilityAlerts` resource rather than the `vulnerabilityAlerts` field on the repository, which the provider deprecates and will drop in its next major.

`~DEFAULT_BRANCH` rather than a literal branch name, so the five repos still on `master` are covered today and stay covered after they are renamed.

`RequirePR` additionally bars direct pushes, so changes have to arrive as a squash-merged PR. Only `secondhand` sets it, matching the rule it already had.

### Deliberately not managed

**`required_status_checks`** - it deadlocks direct pushes to the default branch, since a commit cannot have a passing check before it exists on the remote, and `atqamz/universe`'s package updater pushes straight to `main` on green.

**`required_signatures`** - tempting given the sign-always rule, but it would break two automations that commit unsigned by design: universe's package updater (`nix-update --commit` under a bot identity) and `zen-profile-sync` (`-c commit.gpgsign=false`, so its snapshots stay off the contribution graph).

**Labels** - dependabot creates `dependencies`, `nix` and `github_actions` on its own, so declaring the label set here would fight it on every alert.

**Archived repos** (`brain`, `mirufm`) - GitHub rejects field writes on an archived repo, so including them would mean a permanently failing diff for no benefit. Unarchive first if either needs to come back under management.

**`allow_forking`** - not settable at all here. GitHub answers `422 Allow forks setting can only be changed on org-owned private repositories` for every personal repo, public or private. Worse, the provider reports the update as succeeding while GitHub ignores it, which leaves Pulumi state holding a value that does not exist and a phantom diff on the next refresh. Do not add the field back.

**Secret scanning and push protection** - no field on `Repository` and no dedicated resource in the provider (checked in 6.14.1). Enabled by hand on the public repos:

```bash
gh api -X PATCH repos/atqamz/<repo> --input - <<'EOF'
{"security_and_analysis":{"secret_scanning":{"status":"enabled"},"secret_scanning_push_protection":{"status":"enabled"}}}
EOF
```

The five private repos report it unavailable - it needs Advanced Security, which a personal plan does not include.

**Interaction limits** - also absent from the provider, and applied by hand to the config and profile repos so only collaborators can open issues and pull requests:

```bash
gh api -X PUT repos/atqamz/<repo>/interaction-limits -f limit=collaborators_only -f expiry=six_months
```

`six_months` is GitHub's maximum; there is no permanent setting, so these **lapse on 2027-02-04** unless renewed.
That is a silent expiry, which is the failure mode worth wiring a timer against rather than trusting to memory.

Note what this does and does not buy: a public repo on GitHub can always be forked, so an outside PR can always be *opened* once the limit lapses. What stops one landing is that nobody else has write access and the ruleset governs the default branch. The interaction limit is the layer that stops it being opened at all.

**Rule types the provider does not model** - `secondhand` carries a `code_coverage` rule that the Go SDK has no field for. The provider leaves unmodelled rule types in place rather than stripping them, so it survives a run untouched. Do not assume that holds for every rule type; check a preview diff before trusting it.

## Usage

Requires a token with `repo` and `delete_repo`; the devshell has `gh`, so:

```bash
export GITHUB_TOKEN=$(gh auth token)
pulumi preview
pulumi up
```

Every repository resource carries `pulumi.Protect(true)`.
Deleting one therefore needs a deliberate `pulumi state unprotect`, which is the explicit intent such a change should require.

### Adopting a repo that already exists

A repo created outside Pulumi has to be imported, or the run will try to create it and fail:

```bash
pulumi config set adopt true
pulumi up
pulumi config rm adopt
```

`adopt` attaches `pulumi.Import` to the repository, its alert switch, and - where `RulesetID` is set - its existing ruleset.
Remove the flag afterwards so later runs are plain updates, and leave it unset in the committed stack file.
