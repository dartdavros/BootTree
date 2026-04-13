# Self-update release flow

## Goal

Produce release artifacts and one stable GitHub Pages manifest that BootTree can consume via `boottree update`.

## Selected hosting model

BootTree uses GitHub Pages as the canonical host for the stable release manifest.

Stable manifest path:

```text
https://<owner>.github.io/BootTree/updates/stable/manifest.json
```

The workflow also uploads the same `manifest.json` to the tagged GitHub Release as an auxiliary asset, but GitHub Pages is the runtime source of truth.

## What is automated now

The GitHub release workflow now does the following on every `v*` tag:

1. builds and publishes release artifacts with GoReleaser;
2. generates `dist/manifest.json` from the produced artifacts;
3. uploads that manifest to the same GitHub Release as an additional asset;
4. publishes the same manifest to GitHub Pages at `/updates/stable/manifest.json`.

This removes manual manifest generation and manual manifest publishing from the normal tag-based release path.

## Inputs

- GoReleaser `dist/` directory
- Git tag version, for example `v0.4.0`
- Tagged GitHub Release download base URL
- Stable GitHub Pages manifest URL

## GitHub repository setup

Enable GitHub Pages for this repository and allow GitHub Actions to publish it.

Recommended settings:

1. `Settings` → `Pages`
2. `Build and deployment` → `Source` = `GitHub Actions`

Optional repository variable:

- `BOOTTREE_UPDATE_MANIFEST_URL`

If this variable is not set, the release workflow falls back to:

```text
https://<owner>.github.io/BootTree/updates/stable/manifest.json
```

Use the variable only when you want to override the default Pages URL, for example for a custom domain.

## Generate the manifest manually

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

## Release workflow behavior

Current workflow behavior:

- artifacts are published by GoReleaser;
- asset URLs inside the manifest point to the tagged GitHub Release download URLs;
- the generated manifest is uploaded to the tagged GitHub Release as `manifest.json`;
- the same manifest is published to GitHub Pages as `/updates/stable/manifest.json`;
- GoReleaser injects `boottree/internal/buildinfo.UpdateManifestURL` with the stable GitHub Pages manifest URL.

This is enough for:

- baked-in `boottree update` on tagged releases;
- explicit HTTPS `--manifest-url` usage;
- testing the full update pipeline against a concrete tagged release;
- stable production self-update via GitHub Pages.

## Publish requirements for production self-update

For self-update to work in a production release without manually passing `--manifest-url`, the following must all be true:

1. release assets are published at stable HTTPS URLs;
2. `manifest.json` is published at one canonical stable HTTPS URL;
3. release builds inject `boottree/internal/buildinfo.UpdateManifestURL` with that manifest URL.

With GitHub Pages selected, the remaining operational requirement is simply that Pages stays enabled and healthy for the repository.

## Operational notes

- The stable manifest always tracks the latest tagged release published by the workflow.
- Asset URLs inside the manifest remain version-specific GitHub Release download URLs.
- The stable Pages manifest is intentionally channel-scoped at `/updates/stable/manifest.json` so later channels such as `/updates/beta/manifest.json` can be added without changing the runtime URL model.


## Manifest generation source of truth

`manifest.json` is generated from GoReleaser metadata in `dist/artifacts.json`, not by scanning every file in `dist`. This keeps the update manifest aligned with actual release artifacts and avoids coupling to incidental files like changelogs or future metadata files.
