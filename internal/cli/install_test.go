package cli

import (
	"bytes"
	"strings"
	"testing"

	"boottree/internal/platform"
)

type stubInstaller struct {
	state         platform.InstallState
	result        platform.InstallResult
	detectErr     error
	installErr    error
	installCalled bool
}

func (s *stubInstaller) Detect() (platform.InstallState, error) {
	return s.state, s.detectErr
}

func (s *stubInstaller) InstallForCurrentUser() (platform.InstallResult, error) {
	s.installCalled = true
	return s.result, s.installErr
}

func TestRunInstall_ExplicitCommandInstallsWithoutTTYConfirmation(t *testing.T) {
	installer := &stubInstaller{
		state: platform.InstallState{
			CommandName:         "boottree.exe",
			CurrentExecutable:   `C:\Temp\boottree.exe`,
			SuggestedInstallDir: `C:\Users\tester\AppData\Local\Programs\BootTree\bin`,
		},
		result: platform.InstallResult{
			InstalledExecutable:     `C:\Users\tester\AppData\Local\Programs\BootTree\bin\boottree.exe`,
			InstallDir:              `C:\Users\tester\AppData\Local\Programs\BootTree\bin`,
			PathUpdated:             true,
			ShellRestartRecommended: true,
		},
	}
	cmd := newInstallCommandWithDependencies(installer, surveyInitPrompter{})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !installer.installCalled {
		t.Fatal("expected install command to invoke installer")
	}
	text := out.String()
	if !strings.Contains(text, "BootTree installation completed.") {
		t.Fatalf("expected install summary, got %q", text)
	}
}

func TestShouldOfferInstall(t *testing.T) {
	tests := []struct {
		name        string
		commandPath string
		goos        string
		interactive bool
		want        bool
	}{
		{name: "windows init tty", commandPath: "boottree init", goos: "windows", interactive: true, want: true},
		{name: "windows install command", commandPath: "boottree install", goos: "windows", interactive: true, want: false},
		{name: "windows version command", commandPath: "boottree version", goos: "windows", interactive: true, want: false},
		{name: "non interactive", commandPath: "boottree init", goos: "windows", interactive: false, want: false},
		{name: "linux", commandPath: "boottree init", goos: "linux", interactive: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldOfferInstall(tt.commandPath, tt.goos, tt.interactive); got != tt.want {
				t.Fatalf("shouldOfferInstall(%q, %q, %v) = %v, want %v", tt.commandPath, tt.goos, tt.interactive, got, tt.want)
			}
		})
	}
}
