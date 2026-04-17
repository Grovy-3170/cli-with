# Releasing

This is the checklist for cutting a new release of `cli-with`. Follow it top to bottom every time — skipping a step usually shows up as a broken install, a stale `@latest`, or a tag with no binaries.

## TL;DR

```bash
# 1. Write your feature / fix, commit local changes
# 2. Update CHANGELOG.md and (if user-facing) README.md
# 3. Run tests + build
make test && make build

# 4. Commit
git add <files>
git commit -m "feat(scope): one-line summary"

# 5. Push main
git push origin main --no-verify

# 6. Tag and push the tag (this triggers the release workflow)
VERSION=v0.3.2
git tag -a $VERSION -m "$VERSION — short summary"
git push origin $VERSION --no-verify

# 7. Prime the Go module proxy so `go install @latest` picks up the new version
curl -sS "https://proxy.golang.org/github.com/!grovy-3170/cli-with/@v/$VERSION.info"

# 8. Verify
# - https://github.com/Grovy-3170/cli-with/actions → workflow green
# - https://github.com/Grovy-3170/cli-with/releases → binaries present
# - https://proxy.golang.org/github.com/!grovy-3170/cli-with/@latest → returns new version
```

---

## Step-by-step

### 1. Decide the version number

We follow [Semantic Versioning](https://semver.org/):

| Change | Bump | Example |
|---|---|---|
| Breaking change (CLI flag removed, behavior change users relied on) | **major** | `v0.3.1 → v1.0.0` |
| New feature, backwards-compatible | **minor** | `v0.3.1 → v0.4.0` |
| Bug fix, docs, internal refactor, CI tweak | **patch** | `v0.3.1 → v0.3.2` |

Pre-1.0 we tolerate breaking changes in minor bumps, but flag them loudly in the CHANGELOG.

For pre-release tags (`v0.4.0-rc1`, `v0.4.0-beta1`), GoReleaser automatically marks the GitHub release as a pre-release.

### 2. Update `CHANGELOG.md`

Add a new section at the top, under the header. Date is ISO (`YYYY-MM-DD`).

```markdown
## [v0.3.2] - 2026-04-20

### Added
- One-line description of each new thing.

### Changed
- Things whose behavior changed.

### Fixed
- Bugs fixed.

### Security
- Security-relevant fixes.
```

Only include the sections that apply. Don't ship an empty `### Added` block.

### 3. Update `README.md` (if the change is user-facing)

- New command or flag → add it to the relevant section and to the commands table.
- Changed behavior → update the existing description.
- Internal-only changes → skip the README.

### 4. Run tests + build

```bash
make test       # all packages must pass
make build      # binary compiles cleanly
```

If anything fails, fix it before going further. A broken tag is painful to recover from.

> **Do not commit the `with` binary.** `make build` regenerates it with whatever version string Git currently reports (usually `vX.Y.Z-dirty`), which is stale the moment you tag. Users install from source via `go install` or from the Release archives built by CI — neither path uses the committed binary. Leave it in the working tree; it's already tracked for historical reasons but don't stage changes to it.

### 5. Commit

Use [Conventional Commits](https://www.conventionalcommits.org/) prefixes so the changelog stays readable:

| Prefix | When |
|---|---|
| `feat(scope):` | New user-facing feature |
| `fix(scope):` | Bug fix |
| `docs:` | Docs-only change |
| `ci:` | CI / release infra |
| `chore:` | Housekeeping, version bumps, dependency updates |
| `refactor(scope):` | Internal restructure, no behavior change |
| `test(scope):` | Test-only changes |

Stage files explicitly — **never** use `git add .` or `-A`, to avoid accidentally committing the stale binary, a local password file, or an alias file.

```bash
git add CHANGELOG.md README.md internal/... cmd/with/...
git commit -m "feat(alias): with alias command for saved exec shortcuts"
```

Multi-paragraph commit bodies with context are encouraged for substantial changes.

### 6. Push to `main`

```bash
git push origin main --no-verify
```

Why `--no-verify`:
- It skips **client-side** pre-push hooks (none exist in this repo today, so it's a no-op now — but the flag is safe to keep as a habit).
- The "Bypassed rule violations" message you'll see from GitHub is a **server-side** bypass of the "Changes must be made through a pull request" ruleset. That bypass happens because your GitHub account has admin permissions on this repo, not because of `--no-verify`. Regular contributors would have to open a PR.

If you ever add pre-push hooks (e.g., running `make test` before push), you'll want to drop `--no-verify` for real releases so those hooks actually run.

### 7. Tag and push the tag

```bash
VERSION=v0.3.2
git tag -a $VERSION -m "$VERSION — short summary"
git push origin $VERSION --no-verify
```

Rules:
- Tags are **always** annotated (`-a`), never lightweight. GoReleaser and Go modules both want annotation metadata.
- Tag name **must** start with `v` (e.g., `v0.3.2`, not `0.3.2`). The Release workflow trigger is `tags: ["v*"]` and Go modules require the `v` prefix.
- Never re-tag a version that already has a Release or that the Go proxy has cached. If you need to replace it, bump the patch version instead. Moving a tag breaks anyone who already pulled the old SHA.

### 8. Prime the Go module proxy

```bash
curl -sS "https://proxy.golang.org/github.com/!grovy-3170/cli-with/@v/$VERSION.info"
```

Note the `!grovy-3170` spelling: the Go proxy lowercases uppercase letters and prefixes them with `!`, so `Grovy-3170` becomes `!grovy-3170`.

This HTTP call makes the proxy fetch the tag immediately. Without it, `go install github.com/Grovy-3170/cli-with/cmd/with@latest` may resolve to an older version for up to ~10 minutes (the first user to request the new version triggers indexing, but that's a race).

Verify it took:
```bash
curl -sS "https://proxy.golang.org/github.com/!grovy-3170/cli-with/@latest"
```

Should return the new version.

### 9. Verify CI and the release

After you push the tag, the `Release` workflow in `.github/workflows/release.yml` triggers automatically. It runs GoReleaser, which cross-compiles for Linux/macOS/Windows (amd64 + arm64) and uploads archives to the Releases page.

Watch it:
- **Actions:** https://github.com/Grovy-3170/cli-with/actions (expect green in ~2-3 min)
- **Releases:** https://github.com/Grovy-3170/cli-with/releases (should show the new version with ~10 assets: 6 binaries + checksums.txt + source archives)

Spot-check the install one-liner from the README on your own machine.

---

## What's automated vs manual

| Piece | Who does it |
|---|---|
| Cross-platform binary builds | GoReleaser (triggered by tag) |
| GitHub Release page + archive uploads | GoReleaser |
| `checksums.txt` generation | GoReleaser |
| Release notes (auto-generated from commit log) | GoReleaser |
| Go proxy indexing (`@latest` resolution) | Manual `curl` (step 8) |
| CHANGELOG.md updates | Manual |
| README.md updates | Manual |
| Version tag creation | Manual |

---

## If something goes wrong

### The workflow failed red

Open the failing job in the Actions tab, read the error, fix it. Then:

1. Delete the broken tag locally and on remote:
   ```bash
   git tag -d v0.3.2
   git push origin :refs/tags/v0.3.2
   ```
2. Commit the fix.
3. Re-tag with the **same** version — this is the one time re-tagging is acceptable, because no Release was successfully published, so nobody could have consumed it.
4. Re-push the tag. Workflow re-runs.

**Exception:** if the tag was live long enough that the Go proxy indexed it (step 8 happens immediately, so it usually did), bump the patch version instead (`v0.3.3`). The proxy never forgets a version, and publishing different code under the same tag is the one thing that poisons your users' caches.

### Binaries are missing from the Release page

The workflow succeeded but assets are missing → check the `archives` section of `.goreleaser.yml` for a typo. Fix, commit, bump patch version, tag.

### `go install @latest` installs the old version

The proxy cache didn't update. Re-run step 8. If it still returns the old version after 5 minutes, try:
```bash
curl -sS "https://proxy.golang.org/github.com/!grovy-3170/cli-with/@v/list"
```
to see all known versions. If the new one is there but `@latest` doesn't return it, the Go proxy's `@latest` logic prefers semver — check you didn't tag something like `v0.3.2-dirty` by accident.

### You tagged but forgot to update CHANGELOG

Ship a patch release (`v0.3.3`) with just the CHANGELOG fix. GoReleaser will publish binaries identical in behavior to the previous tag. Don't move the old tag.

---

## Release cadence

No fixed cadence. Release when:
- A feature is complete and tested.
- A bug fix is worth shipping.
- You've accumulated enough small changes that tagging them together is useful.

Batching unrelated changes into one release is fine; splitting a big feature across multiple pre-release tags (`-rc1`, `-rc2`) is also fine. Prefer shipping small and often over big-bang releases.
