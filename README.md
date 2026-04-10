# 🦜 BootTree

BootTree is a cross-platform CLI for standardizing local project structure from presets.

## MVP commands

- `boottree init` — build a safe execution plan, preview it, and apply it on confirmation
- `boottree tree` — render the current project tree with ignore rules and optional depth limit
- `boottree stats` — print structure statistics, empty directories, and secret-like filenames
- `boottree version` — print build information

## Quick start

### Build

```powershell
# Windows PowerShell
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o .\dist\boottree.exe .\cmd\boottree
```

```bash
# Linux or macOS
GOOS=linux GOARCH=amd64 go build -o ./dist/boottree ./cmd/boottree
GOOS=darwin GOARCH=amd64 go build -o ./dist/boottree-darwin ./cmd/boottree
```

### Local usage

```powershell
boottree --help
boottree --version
boottree init
boottree init --preset software-product --mode folders-only --dry-run
boottree init --include 01_business,06_engineering --yes
boottree tree --depth 2
boottree stats
```

## Preset

MVP ships with the built-in preset `software-product`. It creates the following top-level project areas:

- `00_inbox`
- `01_business`
- `02_product`
- `03_marketing`
- `04_brand`
- `05_docs`
- `06_engineering`
- `07_repos`
- `08_deploy`
- `09_assets`
- `10_secrets`
- `99_archive`

## Init behavior

`boottree init` always follows the same model:

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

Default local builds print `dev` metadata. Release builds can inject version information with linker flags:

```powershell
go build -ldflags="-X boottree/internal/buildinfo.Version=1.0.0 -X boottree/internal/buildinfo.Commit=abc1234 -X boottree/internal/buildinfo.BuildDate=2026-04-10T12:00:00Z" -o .\dist\boottree.exe .\cmd\boottree
```
