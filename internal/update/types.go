package update

import "time"

type Options struct {
	CheckOnly   bool
	Yes         bool
	Version     string
	Channel     string
	ManifestURL string
	InstallPath string
}

type Manifest struct {
	Channel     string    `json:"channel"`
	Latest      string    `json:"latest"`
	PublishedAt time.Time `json:"publishedAt"`
	Assets      []Asset   `json:"assets,omitempty"`
	Releases    []Release `json:"releases,omitempty"`
}

type Release struct {
	Version     string    `json:"version"`
	PublishedAt time.Time `json:"publishedAt"`
	Assets      []Asset   `json:"assets"`
}

type Asset struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	URL     string `json:"url"`
	SHA256  string `json:"sha256"`
	Archive string `json:"archive"`
	Binary  string `json:"binary"`
}

type Plan struct {
	CurrentVersion       string
	TargetVersion        string
	GOOS                 string
	GOARCH               string
	Channel              string
	ManifestURL          string
	InstallPath          string
	Asset                Asset
	TempDir              string
	TempArchivePath      string
	ExtractedBinaryPath  string
	BackupPath           string
	NeedsElevation       bool
	RequiresDeferredSwap bool
	IsNoop               bool
}

type Result struct {
	InstalledVersion string
	PreviousVersion  string
	InstallPath      string
	BackupPath       string
	RestartRequired  bool
	Deferred         bool
}
