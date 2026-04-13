# Self-update release flow

## Goal

Produce release artifacts and a release manifest that BootTree can consume via `boottree update`.

## What is automated now

The GitHub release workflow now does the following on every `v*` tag:

1. builds and publishes release artifacts with GoReleaser;
2. generates `dist/manifest.json` from the produced artifacts;
3. uploads that manifest to the same GitHub Release as an additional asset.

This removes manual manifest generation from the normal tag-based release path.

## Inputs

- GoReleaser `dist/` directory
- Git tag version, for example `v0.4.0`
- Stable base URL where the release assets will be reachable over HTTPS

## Generate the manifest manually

Example:

```bash
python3 scripts/generate_release_manifest.py   --version v0.4.0   --dist-dir dist   --base-url https://github.com/<owner>/<repo>/releases/download/v0.4.0   --output dist/manifest.json
```

The generated manifest contains:

- channel
- latest version
- release list
- per-platform asset URL
- SHA-256 checksum
- archive type
- binary name

## Release workflow behavior

Current workflow behavior:

- artifacts are published by GoReleaser;
- asset URLs inside the manifest point to the tagged GitHub Release download URLs;
- the generated manifest is uploaded to the same GitHub Release as `manifest.json`.

This is enough for:

- manual verification;
- explicit HTTPS `--manifest-url` usage;
- testing the full update pipeline against a concrete tagged release.

## Publish requirements for production self-update

For self-update to work in a production release without manually passing `--manifest-url`, the following must all be true:

1. release assets are published at stable HTTPS URLs;
2. `manifest.json` is published at one canonical stable HTTPS URL;
3. release builds inject `boottree/internal/buildinfo.UpdateManifestURL` with that manifest URL.

## Current practical options

### Option A — explicit manifest URL

Ship the binary without a baked-in manifest URL and document an HTTPS manifest endpoint:

```powershell
boottree update --check --manifest-url <https-url>
boottree update --yes --manifest-url <https-url>
```

This works immediately and is already supported.

### Option B — baked-in stable manifest URL

Publish the manifest to a stable location such as:

```text
https://downloads.example.com/boottree/stable/manifest.json
```

Then set repository variable `BOOTTREE_UPDATE_MANIFEST_URL` so GoReleaser injects it into the binary during tagged releases.

## What is still not solved automatically

This repository now automates manifest generation and release attachment, but it still does not provision hosting for one canonical stable manifest URL.

That hosting decision is still operational. Until it is made, production binaries should rely on an explicit HTTPS `--manifest-url` or on a separately managed stable manifest host.
