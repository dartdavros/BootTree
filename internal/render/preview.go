package render

import (
	"fmt"
	"strings"

	"boottree/internal/core/model"
)

func RenderExecutionPlan(plan model.ExecutionPlan) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Plan summary\n")
	fmt.Fprintf(&b, "  Directories to create: %d\n", len(plan.DirectoriesToCreate))
	fmt.Fprintf(&b, "  Files to create: %d\n", len(plan.FilesToCreate))
	fmt.Fprintf(&b, "  Skipped existing: %d\n", len(plan.SkippedExisting))
	fmt.Fprintf(&b, "  Conflicts: %d\n", len(plan.Conflicts))
	fmt.Fprintf(&b, "  Warnings: %d\n", len(plan.Warnings))

	appendSection(&b, "Directories to create", plan.DirectoriesToCreate)
	appendSection(&b, "Files to create", plan.FilesToCreate)
	appendSection(&b, "Skipped existing", plan.SkippedExisting)
	appendSection(&b, "Conflicts", plan.Conflicts)
	if len(plan.Warnings) > 0 {
		b.WriteString("\nWarnings\n")
		for _, warning := range plan.Warnings {
			fmt.Fprintf(&b, "  - %s\n", warning)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func appendSection(b *strings.Builder, title string, actions []model.PlanAction) {
	if len(actions) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s\n", title)
	for _, action := range actions {
		fmt.Fprintf(b, "  - %s (%s)\n", action.Path, action.Reason)
	}
}
