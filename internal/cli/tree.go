package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"boottree/internal/app"
	"boottree/internal/core/scanner"
	"boottree/internal/fs"
	"boottree/internal/render"

	"github.com/spf13/cobra"
)

type parsedTreeArgs struct {
	Depth int
	All   bool
}

func newTreeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tree",
		Short: "Render the current project tree",
		Long:  "Scan the current directory, apply ignore rules by default, and render a stable text tree view.",
		Example: "  boottree tree\n" +
			"  boottree tree --depth 2\n" +
			"  boottree tree --all",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runTree(cmd, args); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
				os.Exit(1)
			}
		},
	}
}

func runTree(cmd *cobra.Command, args []string) error {
	parsed, err := parseTreeArgs(args)
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

func parseTreeArgs(args []string) (parsedTreeArgs, error) {
	parsed := parsedTreeArgs{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--all":
			parsed.All = true
		case strings.HasPrefix(arg, "--depth="):
			depth, err := parseDepthValue(strings.TrimPrefix(arg, "--depth="))
			if err != nil {
				return parsedTreeArgs{}, err
			}
			parsed.Depth = depth
		case arg == "--depth":
			i++
			if i >= len(args) {
				return parsedTreeArgs{}, fmt.Errorf("flag --depth requires a value")
			}
			depth, err := parseDepthValue(args[i])
			if err != nil {
				return parsedTreeArgs{}, err
			}
			parsed.Depth = depth
		default:
			return parsedTreeArgs{}, fmt.Errorf("unknown argument: %s", arg)
		}
	}
	return parsed, nil
}

func parseDepthValue(value string) (int, error) {
	depth, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("invalid depth %q: must be a non-negative integer", value)
	}
	if depth < 0 {
		return 0, fmt.Errorf("invalid depth %q: must be a non-negative integer", value)
	}
	return depth, nil
}
