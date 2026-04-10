package cli

import (
	"context"
	"fmt"
	"os"

	"boottree/internal/app"
	"boottree/internal/core/scanner"
	corestats "boottree/internal/core/stats"
	"boottree/internal/fs"
	"boottree/internal/render"

	"github.com/spf13/cobra"
)

func newStatsCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "stats",
		Short:   "Show project structure statistics",
		Long:    "Scan the current directory and print a human-readable summary of directories, files, extensions, empty folders, and secret-like filenames.",
		Example: "  boottree stats",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStats(cmd)
		},
	}
}

func runStats(cmd *cobra.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve current directory: %w", err)
	}

	builder := app.StatsBuilder{
		Scanner: scanner.Service{FS: fs.OSFileSystem{}},
		Stats:   corestats.Service{},
	}
	stats, err := builder.Build(context.Background(), cwd)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), render.RenderStats(stats))
	return nil
}
