package cli

import (
	"context"
	"fmt"
	"os"

	"boottree/internal/app"
	"boottree/internal/core/scanner"
	"boottree/internal/fs"
	"boottree/internal/render"

	"github.com/spf13/cobra"
)

type treeCommandFlags struct {
	Depth int
	All   bool
}

func newTreeCommand() *cobra.Command {
	flags := &treeCommandFlags{}
	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Render the current project tree",
		Long:  "Scan the current directory, apply ignore rules by default, and render a stable text tree view.",
		Example: "  boottree tree\n" +
			"  boottree tree --depth 2\n" +
			"  boottree tree --all",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTree(cmd, flags)
		},
	}
	cmd.Flags().IntVar(&flags.Depth, "depth", 0, "Maximum directory depth to render (0 means unlimited)")
	cmd.Flags().BoolVar(&flags.All, "all", false, "Include entries normally filtered by default ignore rules")
	return cmd
}

func runTree(cmd *cobra.Command, flags *treeCommandFlags) error {
	if flags.Depth < 0 {
		return fmt.Errorf("invalid depth %q: must be a non-negative integer", cmd.Flag("depth").Value.String())
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve current directory: %w", err)
	}

	builder := app.TreeBuilder{Scanner: scanner.Service{FS: fs.OSFileSystem{}}}
	snapshot, err := builder.BuildSnapshot(context.Background(), cwd, app.TreeOptions{IncludeIgnored: flags.All})
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), render.RenderTree(snapshot, render.TreeRenderOptions{MaxDepth: flags.Depth}))
	return nil
}
