//go:build windows

package update

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func installBinary(ctx context.Context, plan Plan, extractedPath string) (Result, error) {
	_ = ctx
	if !plan.RequiresDeferredSwap {
		if err := os.MkdirAll(filepath.Dir(plan.InstallPath), 0o755); err != nil {
			return Result{}, fmt.Errorf("create install directory %q: %w", filepath.Dir(plan.InstallPath), err)
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

	scriptPath := filepath.Join(plan.TempDir, "boottree-self-update.ps1")
	scriptContent := buildPowerShellSwapScript(int64(os.Getpid()), extractedPath, plan.InstallPath, plan.BackupPath)
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o600); err != nil {
		return Result{}, fmt.Errorf("write update helper script %q: %w", scriptPath, err)
	}

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	cmd.SysProcAttr = windowsDetachedProcessAttrs()
	if err := cmd.Start(); err != nil {
		return Result{}, fmt.Errorf("start deferred update helper %q: %w", scriptPath, err)
	}
	return Result{InstalledVersion: plan.TargetVersion, PreviousVersion: plan.CurrentVersion, InstallPath: plan.InstallPath, BackupPath: plan.BackupPath, RestartRequired: true, Deferred: true}, nil
}

func buildPowerShellSwapScript(pid int64, source string, target string, backup string) string {
	return strings.Join([]string{
		"$ErrorActionPreference = 'Stop'",
		"$parentPid = " + strconv.FormatInt(pid, 10),
		"$source = " + quotePowerShell(source),
		"$target = " + quotePowerShell(target),
		"$backup = " + quotePowerShell(backup),
		"for ($i = 0; $i -lt 300; $i++) {",
		"  if (-not (Get-Process -Id $parentPid -ErrorAction SilentlyContinue)) { break }",
		"  Start-Sleep -Milliseconds 200",
		"}",
		"$targetDir = Split-Path -Parent $target",
		"New-Item -ItemType Directory -Force -Path $targetDir | Out-Null",
		"if (Test-Path $backup) { Remove-Item -Force $backup }",
		"if (Test-Path $target) { Move-Item -Force $target $backup }",
		"Move-Item -Force $source $target",
		"Remove-Item -Force $MyInvocation.MyCommand.Path",
	}, "\n")
}

func quotePowerShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
