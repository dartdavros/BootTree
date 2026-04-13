#!/usr/bin/env python3
from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import re
import sys
from dataclasses import dataclass
from datetime import datetime, timezone

PROJECT_NAME = "boottree"


@dataclass(frozen=True)
class AssetEntry:
    os: str
    arch: str
    archive: str
    binary: str
    filename: str
    sha256: str


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate BootTree release manifest from GoReleaser dist artifacts.")
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


def detect_archive(filename: str) -> str:
    lower = filename.lower()
    if lower.endswith(".tar.gz"):
        return "tar.gz"
    if lower.endswith(".zip"):
        return "zip"
    if lower.endswith(".exe"):
        return "binary"
    raise ValueError(f"unsupported artifact format: {filename}")


def detect_binary(goos: str) -> str:
    return "boottree.exe" if goos == "windows" else "boottree"


def parse_asset_filename(filename: str, version: str) -> tuple[str, str]:
    escaped_version = re.escape(version)
    match = re.fullmatch(rf"{PROJECT_NAME}_{escaped_version}_(linux|darwin|windows)_(amd64|arm64)(?:\.tar\.gz|\.zip|\.exe)", filename)
    if not match:
        raise ValueError(f"artifact does not match expected naming convention: {filename}")
    return match.group(1), match.group(2)


def discover_assets(dist_dir: pathlib.Path, version: str, checksums: dict[str, str]) -> list[AssetEntry]:
    assets: list[AssetEntry] = []
    seen_targets: set[tuple[str, str]] = set()

    for path in sorted(dist_dir.iterdir()):
        if not path.is_file():
            continue
        if path.name.endswith("_checksums.txt"):
            continue
        if path.name.startswith("README"):
            continue
        goos, goarch = parse_asset_filename(path.name, version)
        target = (goos, goarch)
        if target in seen_targets:
            raise ValueError(f"duplicate asset for {goos}/{goarch}: {path.name}")
        seen_targets.add(target)

        sha256_value = checksums.get(path.name)
        if not sha256_value:
            sha256_value = hashlib.sha256(path.read_bytes()).hexdigest()

        assets.append(
            AssetEntry(
                os=goos,
                arch=goarch,
                archive=detect_archive(path.name),
                binary=detect_binary(goos),
                filename=path.name,
                sha256=sha256_value,
            )
        )

    if not assets:
        raise ValueError(f"no release artifacts found in {dist_dir}")
    return assets


def build_manifest(version: str, channel: str, published_at: str, base_url: str, assets: list[AssetEntry]) -> dict:
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
