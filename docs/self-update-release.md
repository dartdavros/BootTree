# Self-update release flow

## Goal

Produce release artifacts and a release manifest that BootTree can consume via `boottree update`.

## Inputs

- GoReleaser `dist/` directory
- Git tag version, for example `v0.4.0`
- Stable base URL where the release assets will be reachable over HTTPS

## Generate the manifest

Example:

```bash
python3 scripts/generate_release_manifest.py \
  --version v0.4.0 \
  --dist-dir dist \
  --base-url https://github.com/<owner>/<repo>/releases/download/v0.4.0 \
  --output dist/manifest.json
```

The generated manifest contains:

- channel
- latest version
- release list
- per-platform asset URL
- SHA-256 checksum
- archive type
- binary name

## Publish requirements

For self-update to work in a production release, the following must all be true:

1. release assets are published at stable HTTPS URLs;
2. `manifest.json` is published at a known HTTPS URL;
3. release builds inject `boottree/internal/buildinfo.UpdateManifestURL` with that manifest URL.

## Current practical options

### Option A — explicit manifest URL

Ship the binary without a baked-in manifest URL and document:

```powershell
boottree update --check --manifest-url <https-url>
boottree update --yes --manifest-url <https-url>
```

This is the simplest option and works immediately.

### Option B — baked-in stable manifest URL

Publish the manifest to a stable location such as:

```text
https://downloads.example.com/boottree/stable/manifest.json
```

Then set `BOOTTREE_UPDATE_MANIFEST_URL` in the release environment so GoReleaser injects it into the binary.

## What this document does not solve

This repository now includes manifest generation, but hosting strategy is still an operational decision. The project still needs one canonical stable HTTPS location for the manifest used by production binaries.
