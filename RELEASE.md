# Release Guide

This project ships from an annotated Git tag through GitHub Actions, GoReleaser, and the Homebrew tap update workflow.

## 1. Prep the repository
1. Update `cmd/tmuxwatch/main.go` with the new semantic version.
2. Update package metadata such as `flake.nix` and version examples in `README.md`.
3. Move unreleased notes into `CHANGELOG.md` under the new version heading.
4. Run formatting, `go test ./...`, `golangci-lint run`, cross-platform builds, and a GoReleaser snapshot.
5. Live-test the snapshot binary against a real tmux session, then commit and push the release prep.
6. Require green `main` CI, a clean checkout, and an empty GitHub issue/PR queue.

## 2. Tag and publish on GitHub
1. Create an annotated tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`.
2. Push the tag: `git push origin vX.Y.Z`.
3. Watch the `release` workflow. GoReleaser publishes macOS, Linux, and Windows archives plus `checksums.txt`.
4. Set the GitHub Release notes from the matching changelog entry and verify the new release is marked latest.

## 3. Update the Homebrew tap
1. The release workflow dispatches `steipete/homebrew-tap` after GoReleaser succeeds.
2. Verify `Formula/tmuxwatch.rb` references the new release assets, checksums, and version assertion.
3. Run `brew reinstall steipete/tap/tmuxwatch` and `brew test steipete/tap/tmuxwatch`.

## 4. Post-release checks
- Verify the tag, GitHub Release, assets, checksums, release notes, release workflow, and Homebrew workflow.
- Restore an empty `Unreleased` changelog section for the next patch, commit, and push.
- Confirm `main` CI is green and the final checkout is clean.
