package cli

import "testing"

func TestRunStats_RejectsUnexpectedArgs(t *testing.T) {
	cmd := newStatsCommand()
	if err := runStats(cmd, []string{"unexpected"}); err == nil {
		t.Fatal("expected runStats to reject unexpected args")
	}
}
