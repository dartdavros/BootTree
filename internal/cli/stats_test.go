package cli

import "testing"

func TestStatsCommand_RejectsUnexpectedArgs(t *testing.T) {
	cmd := newStatsCommand()
	cmd.SetArgs([]string{"unexpected"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected stats command to reject unexpected args")
	}
}
