package buildinfo

import "fmt"

var (
	Version           = "dev"
	Commit            = "none"
	BuildDate         = "unknown"
	UpdateManifestURL = ""
)

func Short() string {
	return fmt.Sprintf("boottree %s", Version)
}

func Detailed() string {
	return fmt.Sprintf("boottree %s\ncommit: %s\nbuilt: %s", Version, Commit, BuildDate)
}
