# Release Guide

This project ships as a tagged GitHub release and as a Homebrew formula in `~/Projects/homebrew-tap`.

## 1. Prep the repository
1. Update `cmd/tmuxwatch/main.go` with the new semantic version.
2. Move unreleased notes into `CHANGELOG.md` under the new version heading.
3. Run the full suite: `go test ./...` (add `staticcheck ./...` once the toolchain matches).
4. Regenerate docs or assets if required, then commit the changes.

## 2. Tag and publish on GitHub
1. Create an annotated tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`.
2. Push main and the tag: `git push origin main --follow-tags`.
3. Draft the GitHub release (notes can reuse the changelog entry) and attach any prebuilt binaries from `dist/` if needed.

## 3. Update the Homebrew tap
1. Switch to `~/Projects/homebrew-tap`.
2. Edit `Formula/tmuxwatch.rb`:
   - Set `url` to `https://github.com/steipete/tmuxwatch/archive/refs/tags/vX.Y.Z.tar.gz`.
   - Recalculate `sha256`: `curl -L $url | shasum -a 256`.
   - Update the `test do` assertion to expect the new version string.
3. Commit and push: `git commit -am "Update tmuxwatch formula to X.Y.Z" && git push`.
4. Verify the formula builds: `brew install --build-from-source steipete/tap/tmuxwatch`.

## 4. Post-release checks
- Confirm Dependabot/GH Actions succeed.
- Announce the release or update internal docs as needed.
- If issues arise, cut a follow-up patch release and repeat these steps.
