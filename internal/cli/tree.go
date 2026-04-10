package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"boottree/internal/app"
	"boottree/internal/core/scanner"
	"boottree/internal/fs"
	"boottree/internal/render"

	"github.com/spf13/cobra"
)

type treeFlagValues struct {
	Depth int
	All   bool
}

type parsedTreeArgs struct {
	Depth int
	All   bool
}

func newTreeCommand() *cobra.Command {
	var flags treeFlagValues
	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Render the current project tree",
		Long:  "Scan the current directory, apply ignore rules by default, and render a stable text tree view.",
		Example: "  boottree tree\n" +
			"  boottree tree --depth 2\n" +
			"  boottree tree --all",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runTree(cmd, flags, args); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().IntVar(&flags.Depth, "depth", 0, "Maximum depth to render; 0 means unlimited")
	cmd.Flags().BoolVar(&flags.All, "all", false, "Include entries that are ignored by default")
	return cmd
}

func runTree(cmd *cobra.Command, flagValues treeFlagValues, args []string) error {
	parsed, err := parseTreeFlags(args, flagValues)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve current directory: %w", err)
	}

	builder := app.TreeBuilder{Scanner: scanner.Service{FS: fs.OSFileSystem{}}}
	snapshot, err := builder.BuildSnapshot(context.Background(), cwd, app.TreeOptions{IncludeIgnored: parsed.All})
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), render.RenderTree(snapshot, render.TreeRenderOptions{MaxDepth: parsed.Depth}))
	return nil
}

func parseTreeFlags(args []string, flagValues treeFlagValues) (parsedTreeArgs, error) {
	if len(args) > 0 {
		return parsedTreeArgs{}, fmt.Errorf("unknown argument: %s", args[0])
	}
	if flagValues.Depth < 0 {
		return parsedTreeArgs{}, fmt.Errorf("invalid depth %q: must be a non-negative integer", strconv.Itoa(flagValues.Depth))
	}
	return parsedTreeArgs{Depth: flagValues.Depth, All: flagValues.All}, nil
}

