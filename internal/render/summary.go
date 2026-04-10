package render

import (
	"fmt"
	"strings"

	"boottree/internal/core/model"
)

func RenderApplySummary(plan model.ExecutionPlan, dryRun bool) string {
	var b strings.Builder
	if dryRun {
		b.WriteString("Dry-run completed\n")
	} else {
		b.WriteString("Apply completed\n")
	}
	fmt.Fprintf(&b, "  Directories handled: %d\n", len(plan.DirectoriesToCreate)+countDirectorySkips(plan.SkippedExisting))
	fmt.Fprintf(&b, "  Files handled: %d\n", len(plan.FilesToCreate)+countFileSkips(plan.SkippedExisting))
	fmt.Fprintf(&b, "  Conflicts: %d\n", len(plan.Conflicts))
	fmt.Fprintf(&b, "  Warnings: %d", len(plan.Warnings))
	return b.String()
}

func countDirectorySkips(actions []model.PlanAction) int {
	count := 0
	for _, action := range actions {
		if strings.Contains(action.Reason, "directory") {
			count++
		}
	}
	return count
}

func countFileSkips(actions []model.PlanAction) int {
	count := 0
	for _, action := range actions {
		if strings.Contains(action.Reason, "file") {
			count++
		}
	}
	return count
}
