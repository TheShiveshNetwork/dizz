# Release Process

This project produces two binaries with independent release pipelines:

| Binary | Tag pattern | Build method | Workflow |
|--------|------------|-------------|----------|
| `dizz` | `v*` | GoReleaser (`.goreleaser.yaml`) | `.github/workflows/release.yaml` |
| `dizzie` | `dizzie-v*` | Go build + `gh release` | `.github/workflows/release-dizzie.yaml` |

Both binaries are published to the same GitHub Releases page on [TheShiveshNetwork/dizz](https://github.com/TheShiveshNetwork/dizz/releases).

---

## Releasing dizz

### Tag format

```
v<major>.<minor>.<patch>
```

Examples: `v1.0.0`, `v0.3.1`, `v2.0.0-rc1`

### Steps

1. Ensure all changes are committed and pushed to `main`.
2. Create and push the tag:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```
3. GitHub Actions triggers `.github/workflows/release.yaml`, which runs GoReleaser with `.goreleaser.yaml`.
4. GoReleaser builds **both `dizz` and `dizzie`** binaries for linux/darwin/windows (amd64 + arm64) and publishes them to GitHub Releases.

### What gets built

- `dizz-linux-amd64`, `dizz-linux-arm64`
- `dizz-darwin-amd64`, `dizz-darwin-arm64`
- `dizz-windows-amd64`
- `dizzie-linux-amd64`, `dizzie-linux-arm64`
- `dizzie-darwin-amd64`, `dizzie-darwin-arm64`
- `dizzie-windows-amd64`

---

## Releasing dizzie (standalone)

Use this when shipping a dizzie-only update that does not include dizz changes.

### Tag format

```
dizzie-v<major>.<minor>.<patch>
```

Examples: `dizzie-v1.0.0`, `dizzie-v0.5.2`, `dizzie-v2.0.0-beta1`

### Steps

1. Ensure all dizzie changes are committed and pushed to `main`.
2. Run the release script:
   ```bash
   ./scripts/release-dizzie.sh dizzie-v1.0.0
   ```
   This validates the tag format, builds the binary, runs tests, creates the git tag, and pushes it.

   Or manually:
   ```bash
   git tag -a dizzie-v1.0.0 -m "Release dizzie-v1.0.0"
   git push origin dizzie-v1.0.0
   ```
3. GitHub Actions triggers `.github/workflows/release-dizzie.yaml`, which builds dizzie binaries directly and publishes them to GitHub Releases.

### What gets built

- `dizzie-linux-amd64`, `dizzie-linux-arm64`
- `dizzie-darwin-amd64`, `dizzie-darwin-arm64`
- `dizzie-windows-amd64`

---

## Tag reference

| Tag | What it releases |
|-----|-----------------|
| `v1.0.0` | dizz + dizzie (full release) |
| `dizzie-v1.0.0` | dizzie only (standalone TUI release) |

Tags are prefixed to avoid collisions. A `v*` tag always builds both binaries. A `dizzie-v*` tag builds only the TUI.
