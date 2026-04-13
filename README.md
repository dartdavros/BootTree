# BootTree 🦜

BootTree is a cross-platform CLI for standardizing local project structure from presets.

## Install

### Download from GitHub Releases

Release builds are published automatically from Git tags.

Artifacts:
- Windows: standalone `boottree.exe`
- Linux: `.tar.gz` archive with the `boottree` binary
- macOS: `.tar.gz` archive with the `boottree` binary

### Windows

1. Download `boottree.exe` from the latest GitHub Release.
2. Run it once from the download folder.
3. If BootTree is not yet available in `PATH`, the CLI will offer to install itself for the current user.
4. After a successful install, close the current PowerShell window and open a new one.
5. Verify the install:

```powershell
boottree version
```

You can also install explicitly at any time:

```powershell
.\boottree.exe install
```

BootTree installs to the current-user location:

```text
%LocalAppData%\Programs\BootTree\bin
```

### Linux

1. Download the matching `.tar.gz` archive from the latest GitHub Release.
2. Extract it.
3. Move `boottree` to a directory already available in `PATH`, or add the chosen directory to `PATH` manually.
4. Verify the install:

```bash
tar -xzf boottree_<version>_linux_amd64.tar.gz
chmod +x boottree
mkdir -p ~/.local/bin
mv ./boottree ~/.local/bin/boottree
boottree version
```

### macOS

1. Download the matching `.tar.gz` archive from the latest GitHub Release.
2. Extract it.
3. Move `boottree` to a directory already available in `PATH`, or add the chosen directory to `PATH` manually.
4. Verify the install:

```bash
tar -xzf boottree_<version>_darwin_arm64.tar.gz
chmod +x boottree
mkdir -p ~/.local/bin
mv ./boottree ~/.local/bin/boottree
boottree version
```

## Build from source

### Requirements

- Go version from `go.mod`

### Build

```powershell
# Windows PowerShell
go mod tidy
go build -o .\dist\boottree.exe .\cmd\boottree
```

```bash
# Linux or macOS
go mod tidy
go build -o ./dist/boottree ./cmd/boottree
```

## Quick start

```powershell
boottree --help
boottree --version
boottree init
boottree init --preset software-product --mode folders-only --dry-run
boottree init --include business,engineering --yes
boottree tree --depth 2
boottree stats
boottree install
boottree completion powershell
```

## MVP commands

- `boottree init` — guided interactive setup with preset, mode, and section selection, then preview and apply on confirmation
- `boottree tree` — render the current project tree with ignore rules and optional depth limit
- `boottree stats` — print structure statistics, empty directories, and secret-like filenames
- `boottree install` — install the current executable for the current user and update `PATH` where supported
- `boottree version` — print build information
- `boottree completion <shell>` — generate shell completion scripts via Cobra

## Repository layout

```text
BootTree/
├─ cmd/
├─ internal/
├─ presets/
├─ templates/
├─ testdata/
└─ docs/
```

`presets/` and `templates/` live at the top level so the embedded data model stays discoverable, data-driven, and easy to evolve independently from CLI and core logic.

## Preset

MVP ships with the built-in preset `software-product`. It creates the following top-level project areas:

- `inbox`
- `business`
- `product`
- `marketing`
- `brand`
- `docs`
- `engineering`
- `repos`
- `deploy`
- `assets`
- `secrets`
- `archive`

## Init behavior

`boottree init` always follows the same model. With no flags it opens an interactive flow for preset, mode, and section selection:

1. load preset
2. scan current directory
3. build execution plan
4. render preview
5. apply only after confirmation, unless `--yes` is provided

### Modes

- `folders-only`
- `folders-and-templates`

### Safety rules

- existing directories are skipped
- existing files are skipped by default
- `--force` allows overwrite only for template files BootTree is about to create
- `--dry-run` renders the same plan without writing to disk

## Tree behavior

- ignore rules are enabled by default
- `--depth` limits traversal depth
- `--all` disables default ignore filtering for the current command run

## Stats behavior

`boottree stats` reports:

- total directories
- total files
- empty directories
- extension distribution
- secret-like filenames such as `.env`, `*.key`, `*.pem`, `secrets.*`, `credentials.*`

BootTree never reads or prints the contents of secret-like files.

## Version metadata

Default local builds print `dev` metadata. Release builds inject version information from the Git tag, commit, and build date.

```powershell
go build -ldflags="-X boottree/internal/buildinfo.Version=1.0.0 -X boottree/internal/buildinfo.Commit=abc1234 -X boottree/internal/buildinfo.BuildDate=2026-04-10T12:00:00Z" -o .\dist\boottree.exe .\cmd\boottree
```

## Releases for maintainers

- every push and pull request runs CI on Windows, Linux, and macOS
- pushing a tag like `v0.1.0` creates a GitHub Release through GoReleaser
- release artifacts include Windows binaries, Linux/macOS archives, and a checksum file

## License

BootTree is licensed under the MIT License. See `LICENSE`.
