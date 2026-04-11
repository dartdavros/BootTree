//go:build !windows

package platform

func ensureUserPathEntry(dir string) (bool, error) {
	return false, nil
}
