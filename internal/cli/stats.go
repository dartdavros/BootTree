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
		Use:   "stats",
		Short: "Show project structure statistics",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runStats(cmd, args); err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Error:", err)
				os.Exit(1)
			}
		},
	}
}

func runStats(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown argument: %s", args[0])
	}

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
