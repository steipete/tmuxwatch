# Release Guide

This project ships as a tagged GitHub release and as a Homebrew formula in `~/Projects/homebrew-tap`.

## 1. Prep the repository
1. Update `cmd/tmuxwatch/main.go` with the new semantic version.
2. Move unreleased notes into `CHANGELOG.md` under the new version heading.
3. Run the full suite: `go test ./...` (add `staticcheck ./...` once the toolchain matches).
4. Regenerate docs or assets if required, then commit the changes.
5. Build release zips (from repo root):
   - `make dist` (or `go build` equivalents); ensure `dist/tmuxwatch_darwin_amd64.zip` and `dist/tmuxwatch_darwin_arm64.zip` exist.
   - Create checksums: `shasum -a 256 dist/tmuxwatch_darwin_*.zip > dist/tmuxwatch_darwin_sha256.txt` (include both architectures).

## 2. Tag and publish on GitHub
1. Create an annotated tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`.
2. Push main and the tag: `git push origin main --follow-tags`.
3. Draft the GitHub release (notes can reuse the changelog entry) and attach the built artifacts:
   - `tmuxwatch_darwin_arm64.zip`
   - `tmuxwatch_darwin_amd64.zip`
   - `tmuxwatch_darwin_sha256.txt`

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
