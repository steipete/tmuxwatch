# Release Guide

Releases use the shared Go CLI workflow from `openclaw/release-workflows`, pinned to v1.9.0 (`f613cbfed2b043159c850c353e7facb8c89833b0`). Dispatch `.github/workflows/release.yml` from the current protected `main` head; the workflow owns annotated tag creation. Pushing a tag no longer starts a release.

## 1. Prepare and merge
1. Update `cmd/tmuxwatch/main.go`, `package.json`, `flake.nix`, and the README version example.
2. Finalize `CHANGELOG.md` under a dated `## [X.Y.Z] - YYYY-MM-DD` heading, including highlights. The shared workflow extracts that exact section for the release body and `RELEASE-NOTES.md`.
3. Run `python3 scripts/check-release-metadata.py`, `python3 scripts/test-release-target.py`, `actionlint`, `make check`, `go test ./...`, `go test -race ./...`, and `govulncheck ./...`.
4. On macOS, run `goreleaser check` and `goreleaser release --snapshot --clean --parallelism=2`. Smoke-test the native snapshot against a disposable tmux session.
5. Independently review the changes, open a PR, require green CI on its exact head, and squash-merge. Wait for green CI on `main` before dispatching.

## 2. Signing and deployment target

The build matrix remains macOS, Linux, and Windows on amd64 and arm64. macOS requires **12.0 or newer**. Builds run on macOS so the GoReleaser post-build hook can inspect every Darwin binary with `otool -arch all -l`; `scripts/check-macos-target` rejects absent or mismatched deployment targets. CI builds the same matrix without signing credentials.

`CGO_ENABLED=0` keeps these pure Go builds independent of the host SDK. `MACOSX_DEPLOYMENT_TARGET=12.0` records the build contract; the actual Mach-O load commands are the authority. If cgo is introduced, also set `CGO_CFLAGS=-mmacosx-version-min=12.0` and `CGO_LDFLAGS=-mmacosx-version-min=12.0` and retain the binary check.

The caller uses `repository-type: personal`: **Developer ID Application: Peter Steinberger (Y5PE65HELJ)**. Repository secrets map as follows:

| Repository secret | Shared workflow secret |
| --- | --- |
| `MACOS_SIGN_P12` | `MACOS_SIGNING_P12` |
| `MACOS_SIGN_P12_PASSWORD` | `MACOS_SIGNING_P12_PASSWORD` |
| `ASC_KEY_ID` | `ASC_KEY_ID` |
| `ASC_ISSUER_ID` | `ASC_ISSUER_ID` |
| `ASC_PRIVATE_KEY` | `ASC_PRIVATE_KEY_P8` |
| `HOMEBREW_TAP_TOKEN` | `TAP_TOKEN` |

The signer receives only frozen build artifacts. It signs and notarizes both native macOS binaries. Separate credential-free arm64 and Intel verification jobs check the inventory, hashes, signatures, notarization, stable identifiers, and native execution before the publisher can undraft the release. `ASSET-INVENTORY.json` binds the payload to the source tag/commit; `SIGNING-MANIFEST.json` records signing provenance.

Repository prerequisites: protected `main` with required CI checks; Actions `default_workflow_permissions=write` and `can_approve_pull_request_reviews=true` for the automatic closeout PR. Workflows retain explicit least-privilege permissions.

## 3. Publish and Homebrew handoff
1. Confirm the version is absent from `gh release list --limit 3 --json tagName`, local tags, and remote tags. Never move or replace a release tag.
2. Dispatch: `gh workflow run release.yml --ref main -f version=X.Y.Z`.
3. Watch the exact release run through publication and the Homebrew handoff. On retry, the shared workflow reuses the immutable annotated tag and its frozen source.
4. The six existing archive names remain `tmuxwatch_X.Y.Z_<os>_<arch>.tar.gz` (Windows uses `.zip`). Archives retain `CHANGELOG.md`, `LICENSE`, and `README.md`. The public checksum filename stays `checksums.txt`; no universal archive is added.
5. The workflow dispatches `steipete/homebrew-tap`, waits for its correlated run, and checks that `Formula/tmuxwatch.rb` URLs and hashes match the verified assets. The tap token needs Contents read and Actions write on that tap.

## 4. Verify and close out
- Download all published assets freshly and run `shasum -a 256 -c checksums.txt`. Compare the REST asset digests and inventory with the files and frozen tag.
- Inspect both macOS archives with `codesign -dvv`, `codesign --verify --deep --strict --verbose=4`, and `scripts/check-macos-target`.
- Test a quarantined native download with `spctl -a -vv -t open --context context:primary-signature` and `tmuxwatch --version`. Command-line curl may not set quarantine; when absent, add a test quarantine attribute explicitly and record that fact. Never remove quarantine to make the verification pass.
- Verify the new module tag through `GOPROXY=https://proxy.golang.org go list -m github.com/steipete/tmuxwatch@vX.Y.Z`.
- Run `brew update`, `brew upgrade steipete/tap/tmuxwatch`, `tmuxwatch --version`, and `brew test steipete/tap/tmuxwatch`.
- Review and merge the workflow-created Unreleased PR (or restore `## [Unreleased]` if needed), then leave `main` clean and equal to `origin/main`.
