package render

import (
	"fmt"
	"strings"

	"boottree/internal/update"
)

func UpdatePlan(plan update.Plan) string {
	lines := []string{
		"Update plan",
		fmt.Sprintf("  Current version: %s", plan.CurrentVersion),
		fmt.Sprintf("  Target version: %s", plan.TargetVersion),
		fmt.Sprintf("  Channel: %s", plan.Channel),
		fmt.Sprintf("  Platform: %s/%s", plan.GOOS, plan.GOARCH),
		fmt.Sprintf("  Install path: %s", plan.InstallPath),
		fmt.Sprintf("  Asset URL: %s", plan.Asset.URL),
		fmt.Sprintf("  Archive type: %s", plan.Asset.Archive),
	}
	if plan.NeedsElevation {
		lines = append(lines, "  Permissions: install path may require elevated or additional write permissions")
	}
	if plan.RequiresDeferredSwap {
		lines = append(lines, "  Install strategy: deferred replacement")
	} else {
		lines = append(lines, "  Install strategy: direct replacement")
	}
	if plan.IsNoop {
		lines = append(lines, "  Result: current binary is already up to date")
	}
	return strings.Join(lines, "\n")
}

func UpdateResult(result update.Result) string {
	lines := []string{
		"BootTree update completed.",
		fmt.Sprintf("  Previous version: %s", result.PreviousVersion),
		fmt.Sprintf("  Installed version: %s", result.InstalledVersion),
		fmt.Sprintf("  Install path: %s", result.InstallPath),
	}
	if result.BackupPath != "" {
		lines = append(lines, fmt.Sprintf("  Backup: %s", result.BackupPath))
	}
	if result.Deferred {
		lines = append(lines, "  Status: deferred replacement has been scheduled")
	}
	if result.RestartRequired {
		lines = append(lines, "  Next step: start a new terminal session before using the updated command")
	}
	return strings.Join(lines, "\n")
}
