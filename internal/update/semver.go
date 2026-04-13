package update

import (
	"fmt"
	"strconv"
	"strings"
)

type semanticVersion struct {
	Major int
	Minor int
	Patch int
	Pre   string
}

func normalizeVersion(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "v")
	return value
}

func parseSemanticVersion(value string) (semanticVersion, error) {
	value = normalizeVersion(value)
	if value == "" {
		return semanticVersion{}, fmt.Errorf("empty version")
	}

	parts := strings.SplitN(value, "-", 2)
	core := strings.Split(parts[0], ".")
	if len(core) != 3 {
		return semanticVersion{}, fmt.Errorf("version %q must use major.minor.patch", value)
	}

	major, err := strconv.Atoi(core[0])
	if err != nil {
		return semanticVersion{}, fmt.Errorf("parse major version %q: %w", core[0], err)
	}
	minor, err := strconv.Atoi(core[1])
	if err != nil {
		return semanticVersion{}, fmt.Errorf("parse minor version %q: %w", core[1], err)
	}
	patch, err := strconv.Atoi(core[2])
	if err != nil {
		return semanticVersion{}, fmt.Errorf("parse patch version %q: %w", core[2], err)
	}

	version := semanticVersion{Major: major, Minor: minor, Patch: patch}
	if len(parts) == 2 {
		version.Pre = parts[1]
	}
	return version, nil
}

func compareVersions(left, right string) (int, error) {
	l, err := parseSemanticVersion(left)
	if err != nil {
		return 0, err
	}
	r, err := parseSemanticVersion(right)
	if err != nil {
		return 0, err
	}

	if l.Major != r.Major {
		if l.Major < r.Major {
			return -1, nil
		}
		return 1, nil
	}
	if l.Minor != r.Minor {
		if l.Minor < r.Minor {
			return -1, nil
		}
		return 1, nil
	}
	if l.Patch != r.Patch {
		if l.Patch < r.Patch {
			return -1, nil
		}
		return 1, nil
	}

	switch {
	case l.Pre == "" && r.Pre == "":
		return 0, nil
	case l.Pre == "":
		return 1, nil
	case r.Pre == "":
		return -1, nil
	case l.Pre < r.Pre:
		return -1, nil
	case l.Pre > r.Pre:
		return 1, nil
	default:
		return 0, nil
	}
}
