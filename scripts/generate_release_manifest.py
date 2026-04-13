#!/usr/bin/env python3
from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Any

PROJECT_NAME = "boottree"
SUPPORTED_GOOS = {"linux", "darwin", "windows"}
SUPPORTED_GOARCH = {"amd64", "arm64"}
RUNTIME_ARTIFACT_TYPES = {"Archive", "Binary"}
RELEASE_ARTIFACT_ID = "releases"


@dataclass(frozen=True)
class AssetEntry:
    os: str
    arch: str
    archive: str
    binary: str
    filename: str
    sha256: str


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate BootTree release manifest from GoReleaser metadata.")
    parser.add_argument("--version", required=True, help="Release version, for example v0.4.0 or 0.4.0.")
    parser.add_argument("--dist-dir", required=True, help="Path to GoReleaser dist directory.")
    parser.add_argument("--base-url", required=True, help="Base URL where release assets are published.")
    parser.add_argument("--channel", default="stable", help="Release channel.")
    parser.add_argument("--published-at", default="", help="ISO-8601 timestamp. Defaults to current UTC time.")
    parser.add_argument("--output", required=True, help="Output manifest JSON file path.")
    return parser.parse_args()


def normalize_version(raw: str) -> str:
    value = raw.strip()
    if value.lower().startswith("v"):
        value = value[1:]
    if not value:
        raise ValueError("version is required")
    return value


def normalize_base_url(raw: str) -> str:
    value = raw.strip().rstrip("/")
    if not value:
        raise ValueError("base URL is required")
    return value


def resolve_published_at(raw: str) -> str:
    value = raw.strip()
    if value:
        datetime.fromisoformat(value.replace("Z", "+00:00"))
        return value.replace("+00:00", "Z")
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def load_checksums(dist_dir: pathlib.Path, version: str) -> dict[str, str]:
    checksum_path = dist_dir / f"{PROJECT_NAME}_{version}_checksums.txt"
    if not checksum_path.exists():
        raise FileNotFoundError(f"checksum file not found: {checksum_path}")

    mapping: dict[str, str] = {}
    for raw_line in checksum_path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line:
            continue
        parts = line.split()
        if len(parts) < 2:
            raise ValueError(f"invalid checksum line: {raw_line!r}")
        mapping[parts[1].lstrip("*")] = parts[0].lower()
    return mapping


def load_artifacts_metadata(dist_dir: pathlib.Path) -> list[dict[str, Any]]:
    metadata_path = dist_dir / "artifacts.json"
    if not metadata_path.exists():
        raise FileNotFoundError(f"artifacts metadata file not found: {metadata_path}")

    payload = json.loads(metadata_path.read_text(encoding="utf-8"))
    if not isinstance(payload, list):
        raise ValueError("artifacts metadata must be a JSON array")
    return payload


def normalize_relative_artifact_path(raw: str) -> pathlib.Path:
    value = raw.strip()
    if not value:
        raise ValueError("artifact path is required")

    normalized = pathlib.PurePosixPath(value)
    parts = list(normalized.parts)
    if parts and parts[0] == "dist":
        parts = parts[1:]
    if not parts:
        raise ValueError(f"invalid artifact path: {raw}")
    return pathlib.Path(*parts)


def detect_archive(filename: str, artifact_type: str, extra: dict[str, Any]) -> str:
    fmt = str(extra.get("Format", "")).strip().lower()
    if artifact_type == "Archive":
        if fmt in {"tar.gz", "zip"}:
            return fmt
        raise ValueError(f"unsupported archive format for {filename}: {fmt or '<empty>'}")
    if artifact_type == "Binary":
        return "binary"
    raise ValueError(f"unsupported artifact type for runtime manifest: {artifact_type}")


def detect_binary(goos: str, extra: dict[str, Any]) -> str:
    binary_name = str(extra.get("Binary", "")).strip()
    if binary_name:
        return binary_name
    return "boottree.exe" if goos == "windows" else "boottree"


def resolve_sha256(filename: str, extra: dict[str, Any], checksums: dict[str, str], file_path: pathlib.Path) -> str:
    raw_checksum = str(extra.get("Checksum", "")).strip()
    if raw_checksum:
        if ":" not in raw_checksum:
            raise ValueError(f"invalid checksum format for {filename}: {raw_checksum}")
        algorithm, digest = raw_checksum.split(":", 1)
        if algorithm.lower() != "sha256":
            raise ValueError(f"unsupported checksum algorithm for {filename}: {algorithm}")
        digest = digest.strip().lower()
        if not digest:
            raise ValueError(f"empty checksum digest for {filename}")
        return digest

    sha256_value = checksums.get(filename)
    if sha256_value:
        return sha256_value
    return hashlib.sha256(file_path.read_bytes()).hexdigest()


def discover_assets(dist_dir: pathlib.Path, version: str, checksums: dict[str, str]) -> list[AssetEntry]:
    metadata_entries = load_artifacts_metadata(dist_dir)
    assets: list[AssetEntry] = []
    seen_targets: set[tuple[str, str]] = set()

    for entry in metadata_entries:
        if not isinstance(entry, dict):
            continue

        artifact_type = str(entry.get("type", "")).strip()
        if artifact_type not in RUNTIME_ARTIFACT_TYPES:
            continue

        extra = entry.get("extra")
        if extra is None:
            extra = {}
        if not isinstance(extra, dict):
            raise ValueError(f"artifact extra metadata must be an object for {entry.get('name', '<unknown>')}")

        artifact_id = str(extra.get("ID", "")).strip()
        if artifact_id != RELEASE_ARTIFACT_ID:
            continue

        goos = str(entry.get("goos", "")).strip()
        goarch = str(entry.get("goarch", "")).strip()
        if goos not in SUPPORTED_GOOS or goarch not in SUPPORTED_GOARCH:
            continue

        name = str(entry.get("name", "")).strip()
        raw_path = str(entry.get("path", "")).strip()
        if not name:
            raise ValueError("artifact entry is missing name")
        if not raw_path:
            raise ValueError(f"artifact entry is missing path for {name}")

        target = (goos, goarch)
        if target in seen_targets:
            raise ValueError(f"duplicate runtime artifact for {goos}/{goarch}: {name}")
        seen_targets.add(target)

        artifact_path = dist_dir / normalize_relative_artifact_path(raw_path)
        if not artifact_path.exists():
            raise FileNotFoundError(f"artifact file not found: {artifact_path}")

        assets.append(
            AssetEntry(
                os=goos,
                arch=goarch,
                archive=detect_archive(name, artifact_type, extra),
                binary=detect_binary(goos, extra),
                filename=name,
                sha256=resolve_sha256(name, extra, checksums, artifact_path),
            )
        )

    if not assets:
        raise ValueError(f"no packaged release artifacts with extra.ID={RELEASE_ARTIFACT_ID!r} found in {dist_dir}/artifacts.json")
    return sorted(assets, key=lambda item: (item.os, item.arch))


def build_manifest(version: str, channel: str, published_at: str, base_url: str, assets: list[AssetEntry]) -> dict[str, Any]:
    return {
        "channel": channel,
        "latest": version,
        "publishedAt": published_at,
        "releases": [
            {
                "version": version,
                "publishedAt": published_at,
                "assets": [
                    {
                        "os": asset.os,
                        "arch": asset.arch,
                        "url": f"{base_url}/{asset.filename}",
                        "sha256": asset.sha256,
                        "archive": asset.archive,
                        "binary": asset.binary,
                    }
                    for asset in assets
                ],
            }
        ],
    }


def main() -> int:
    args = parse_args()
    try:
        version = normalize_version(args.version)
        dist_dir = pathlib.Path(args.dist_dir).resolve()
        base_url = normalize_base_url(args.base_url)
        published_at = resolve_published_at(args.published_at)
        checksums = load_checksums(dist_dir, version)
        assets = discover_assets(dist_dir, version, checksums)
        manifest = build_manifest(version, args.channel.strip() or "stable", published_at, base_url, assets)

        output_path = pathlib.Path(args.output).resolve()
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
        return 0
    except Exception as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
