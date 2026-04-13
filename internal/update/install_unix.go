//go:build !windows

package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func installBinary(ctx context.Context, plan Plan, extractedPath string) (Result, error) {
	_ = ctx
	if err := os.MkdirAll(filepath.Dir(plan.InstallPath), 0o755); err != nil {
		return Result{}, fmt.Errorf("create install directory %q: %w", filepath.Dir(plan.InstallPath), err)
	}
	if err := os.Chmod(extractedPath, 0o755); err != nil {
		return Result{}, fmt.Errorf("chmod extracted binary %q: %w", extractedPath, err)
	}

	_ = os.Remove(plan.BackupPath)
	if _, err := os.Stat(plan.InstallPath); err == nil {
		if err := os.Rename(plan.InstallPath, plan.BackupPath); err != nil {
			return Result{}, fmt.Errorf("backup existing executable %q to %q: %w", plan.InstallPath, plan.BackupPath, err)
		}
	}
	if err := os.Rename(extractedPath, plan.InstallPath); err != nil {
		if _, restoreErr := os.Stat(plan.BackupPath); restoreErr == nil {
			_ = os.Rename(plan.BackupPath, plan.InstallPath)
		}
		return Result{}, fmt.Errorf("install new executable to %q: %w", plan.InstallPath, err)
	}
	return Result{InstalledVersion: plan.TargetVersion, PreviousVersion: plan.CurrentVersion, InstallPath: plan.InstallPath, BackupPath: plan.BackupPath}, nil
}
