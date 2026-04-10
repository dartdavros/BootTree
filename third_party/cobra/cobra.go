package cobra

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Command struct {
	Use     string
	Short   string
	Long    string
	Example string
	Run     func(cmd *Command, args []string)

	children []*Command
	parent   *Command
	in       io.Reader
	out      io.Writer
	err      io.Writer
}

func (c *Command) AddCommand(children ...*Command) {
	for _, child := range children {
		if child == nil {
			continue
		}
		child.parent = c
		c.children = append(c.children, child)
	}
}

func (c *Command) Execute() error { return c.execute(os.Args[1:]) }

func (c *Command) execute(args []string) error {
	if len(args) == 0 {
		if c.Run != nil {
			c.Run(c, nil)
			return nil
		}
		return c.Help()
	}

	switch args[0] {
	case "-h", "--help":
		return c.Help()
	case "help":
		return c.executeHelp(args[1:])
	case "-v", "--version":
		if version := c.findChild("version"); version != nil {
			version.Run(version, nil)
			return nil
		}
	}

	for _, child := range c.children {
		if child.Name() == args[0] {
			if len(args) > 1 {
				switch args[1] {
				case "-h", "--help":
					return child.Help()
				case "-v", "--version":
					if version := child.findChild("version"); version != nil {
						version.Run(version, nil)
						return nil
					}
				}
			}
			if child.Run != nil {
				child.Run(child, args[1:])
				return nil
			}
			return child.execute(args[1:])
		}
	}
	return fmt.Errorf("unknown command: %s", args[0])
}

func (c *Command) executeHelp(args []string) error {
	if len(args) == 0 {
		return c.Help()
	}

	child := c.findChild(args[0])
	if child == nil {
		return fmt.Errorf("unknown help topic: %s", args[0])
	}
	if len(args) == 1 {
		return child.Help()
	}
	return child.executeHelp(args[1:])
}

func (c *Command) findChild(name string) *Command {
	for _, child := range c.children {
		if child.Name() == name {
			return child
		}
	}
	return nil
}

func (c *Command) Help() error {
	_, err := io.WriteString(c.OutOrStdout(), c.helpText())
	return err
}

func (c *Command) helpText() string {
	var b strings.Builder

	title := strings.TrimSpace(c.Long)
	if title == "" {
		title = strings.TrimSpace(c.Short)
	}
	if title != "" {
		fmt.Fprintf(&b, "%s\n\n", title)
	}

	fmt.Fprintf(&b, "Usage:\n  %s", c.CommandPath())
	if len(c.children) > 0 && c.Run == nil {
		b.WriteString(" <command>")
	}
	b.WriteString("\n")

	if len(c.children) > 0 {
		b.WriteString("\nAvailable Commands:\n")
		children := append([]*Command(nil), c.children...)
		sort.Slice(children, func(i, j int) bool { return children[i].Name() < children[j].Name() })
		for _, child := range children {
			fmt.Fprintf(&b, "  %-10s %s\n", child.Name(), child.Short)
		}
	}

	b.WriteString("\nFlags:\n")
	b.WriteString("  -h, --help      Show help for this command\n")
	if c.parent == nil && c.findChild("version") != nil {
		b.WriteString("  -v, --version   Print version information\n")
	}

	if example := strings.TrimSpace(c.Example); example != "" {
		fmt.Fprintf(&b, "\nExamples:\n%s\n", example)
	}

	return b.String()
}

func (c *Command) Name() string {
	name := strings.TrimSpace(c.Use)
	if idx := strings.IndexRune(name, ' '); idx >= 0 {
		name = name[:idx]
	}
	return name
}

func (c *Command) CommandPath() string {
	if c.parent == nil {
		return c.Name()
	}
	return c.parent.CommandPath() + " " + c.Name()
}

func (c *Command) InOrStdin() io.Reader {
	if c.in != nil {
		return c.in
	}
	if c.parent != nil {
		return c.parent.InOrStdin()
	}
	return os.Stdin
}

func (c *Command) OutOrStdout() io.Writer {
	if c.out != nil {
		return c.out
	}
	if c.parent != nil {
		return c.parent.OutOrStdout()
	}
	return os.Stdout
}

func (c *Command) ErrOrStderr() io.Writer {
	if c.err != nil {
		return c.err
	}
	if c.parent != nil {
		return c.parent.ErrOrStderr()
	}
	return os.Stderr
}

func (c *Command) SetIn(r io.Reader)  { c.in = r }
func (c *Command) SetOut(w io.Writer) { c.out = w }
func (c *Command) SetErr(w io.Writer) { c.err = w }
